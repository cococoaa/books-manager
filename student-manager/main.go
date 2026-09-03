package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"dgo.baisic.print/GIN/student-manager/dao"
	"dgo.baisic.print/GIN/student-manager/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// JWT 密钥（生产环境应从环境变量读取）
var jwtKey = []byte("student-manager-secret-key")

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  "success",
		Data: data,
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(200, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

func create(c *gin.Context) {
	var req model.Student
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数校验失败: "+err.Error())
		return
	}
	if err := dao.CreateStudent(&req); err != nil {
		Fail(c, 500, "新增学生失败: "+err.Error())
		return
	}
	success(c, req)
}

func renew(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, 400, "无效的ID")
		return
	}
	student, err := dao.GETStudentbyID(id)
	if err != nil {
		Fail(c, 400, "学生不存在")
		return
	}
	var req model.Student
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	student.Age = req.Age
	student.Name = req.Name
	student.Score = req.Score
	student.Time = req.Time
	if err := dao.UpdateStudent(student); err != nil {
		Fail(c, 500, "更新学生信息失败: "+err.Error())
		return
	}
	success(c, student)
}

func getstudentlist(c *gin.Context) {
	students, err := dao.GETALLStudent()
	if err != nil {
		Fail(c, 500, "查询学生列表失败: "+err.Error())
		return
	}
	success(c, students)
}

func detail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, 400, "无效的ID")
		return
	}
	student, err := dao.GETStudentbyID(id)
	if err != nil {
		Fail(c, 400, "学生不存在")
		return
	}
	success(c, student)
}

func deleteBook(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, 400, "无效的ID")
		return
	}
	if _, err := dao.GETStudentbyID(id); err != nil {
		Fail(c, 400, "学生不存在")
		return
	}
	if err := dao.DeleteStudentByID(id); err != nil {
		Fail(c, 500, "删除学生失败: "+err.Error())
		return
	}
	success(c, nil)
}

func initDB() *gorm.DB {
	dsn := os.Getenv("DB_STUDENTS_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/students?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}
	db.AutoMigrate(&model.Student{}, &model.User{})
	return db
}

// 注册
func register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	if req.Username == "" || req.Password == "" {
		Fail(c, 400, "用户名和密码不能为空")
		return
	}
	// 检查用户名是否已存在
	if _, err := dao.GetUserByUsername(req.Username); err == nil {
		Fail(c, 400, "用户名已存在")
		return
	}
	// 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Fail(c, 500, "密码加密失败: "+err.Error())
		return
	}
	user := model.User{
		Username: req.Username,
		Password: string(hashed),
	}
	if err := dao.CreateUser(&user); err != nil {
		Fail(c, 500, "注册失败: "+err.Error())
		return
	}
	success(c, gin.H{"username": user.Username})
}

// 登录
func login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	user, err := dao.GetUserByUsername(req.Username)
	if err != nil {
		Fail(c, 400, "用户名或密码错误")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		Fail(c, 400, "用户名或密码错误")
		return
	}
	// 生成 token（有效期 24 小时）
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		Fail(c, 500, "生成token失败: "+err.Error())
		return
	}
	success(c, gin.H{"token": tokenString})
}

// JWT 鉴权中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			Fail(c, 401, "未登录")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			Fail(c, 401, "token格式错误")
			c.Abort()
			return
		}
		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("无效的签名方法")
			}
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			Fail(c, 401, "token无效或已过期")
			c.Abort()
			return
		}
		c.Next()
	}
}

func main() {
	db := initDB()
	dao.InitDao(db)

	r := gin.Default()

	// 静态前端页面
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// 注册和登录不需要鉴权
	r.POST("/register", register)
	r.POST("/login", login)

	// 学生接口需要 JWT 鉴权
	studentroute := r.Group("/students", authMiddleware())
	{
		studentroute.POST("", create)
		studentroute.PUT("/:id", renew)
		studentroute.GET("", getstudentlist)
		studentroute.GET("/:id", detail)
		studentroute.DELETE("/:id", deleteBook)
	}

	if err := r.Run(":8080"); err != nil {
		panic("服务启动失败: " + err.Error())
	}
}
