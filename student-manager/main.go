package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"dgo.baisic.print/GIN/student-manager/dao"
	"dgo.baisic.print/GIN/student-manager/model"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

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
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/students?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}
	db.AutoMigrate(&model.Student{})
	return db
}

func main() {
	db := initDB()
	dao.InitDao(db)

	r := gin.Default()

	studentroute := r.Group("/students")
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
