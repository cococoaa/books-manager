package main

import (
	"log"
	"net/http"

	"dgo.baisic.print/GIN/dao"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var booklist = make([]book, 0)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `josn:"data"`
}
type book struct {
	ID     string `json:"id"`
	Title  string `json:"title"binding:"required"`
	Author string `json:"author"binding:"required"`
	Year   int    `json:"year" binding:"required,min=1,max=2026"`
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
	var req book
	err := c.ShouldBindJSON(&req)
	if err != nil {
		Fail(c, 400, "参数效验失败"+err.Error())
		return
	}
	req.ID = uuid.NewString()
	booklist = append(booklist, req)
	success(c, req)
}
func renew(c *gin.Context) {
	id := c.Param("id")
	var req book
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误"+err.Error())
		return
	}
	find := false
	for i, item := range booklist {
		if item.ID == id {
			req.ID = id
			booklist[i] = req
			find = true
			break
		}
	}
	if !find {
		Fail(c, 400, "图书不存在")
		return
	}
	success(c, req)
}
func getbooklist(c *gin.Context) {
	success(c, booklist)
}
func detail(c *gin.Context) {
	id := c.Param("id")
	for _, item := range booklist {
		if item.ID == id {
			success(c, item)
			return
		}
	}
	Fail(c, 400, "图书不存在")
}
func delete(c *gin.Context) {
	id := c.Param("id")
	find := false
	newbooklist := make([]book, 0)
	for _, item := range booklist {
		if item.ID != id {
			newbooklist = append(newbooklist, item)
		} else {
			find = true
		}
	}
	if !find {
		Fail(c, 400, "图书不存在")
		return
	}
	booklist = newbooklist
	success(c, nil)
}
func initDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Asia/Shanghai"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败", err)
	}
	// 自动建表
	db.AutoMigrate(&model.User{})
	return db
}
func main() {
	//1.初始化mysql + GORM
	db := initDB()
	//2.把db注入dao层，所有dao共用一个连接池
	dao.InitDao(db)
	r := gin.Default()

	bookroute := r.Group("/books")
	{
		bookroute.POST("", create)
		bookroute.PUT("/:id", renew)
		bookroute.GET("", getbooklist)
		bookroute.GET("/:id", detail)
		bookroute.DELETE("/:id", delete)
	}
	err := r.Run(":8080")
	if err != nil {
		panic("服务启动失败：" + err.Error())
	}
}
