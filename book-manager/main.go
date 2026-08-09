package main

import (
	"log"
	"net/http"
	"os"

	"dgo.baisic.print/GIN/book-manager/dao"
	"dgo.baisic.print/GIN/book-manager/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	var req model.Book
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数校验失败: "+err.Error())
		return
	}
	req.ID = uuid.NewString()
	if err := dao.CreateBook(&req); err != nil {
		Fail(c, 500, "新增图书失败: "+err.Error())
		return
	}
	success(c, req)
}

func renew(c *gin.Context) {
	id := c.Param("id")
	book, err := dao.GetBookById(id)
	if err != nil {
		Fail(c, 400, "图书不存在")
		return
	}
	var req model.Book
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	book.Title = req.Title
	book.Author = req.Author
	book.Year = req.Year
	if err := dao.UpdateBook(book); err != nil {
		Fail(c, 500, "更新图书失败: "+err.Error())
		return
	}
	success(c, book)
}

func getbooklist(c *gin.Context) {
	books, err := dao.GetAllBooks()
	if err != nil {
		Fail(c, 500, "查询图书列表失败: "+err.Error())
		return
	}
	success(c, books)
}

func detail(c *gin.Context) {
	id := c.Param("id")
	book, err := dao.GetBookById(id)
	if err != nil {
		Fail(c, 400, "图书不存在")
		return
	}
	success(c, book)
}

func deleteBook(c *gin.Context) {
	id := c.Param("id")
	if _, err := dao.GetBookById(id); err != nil {
		Fail(c, 400, "图书不存在")
		return
	}
	if err := dao.DeleteBook(id); err != nil {
		Fail(c, 500, "删除图书失败: "+err.Error())
		return
	}
	success(c, nil)
}

func initDB() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/books?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}
	// 自动建表
	db.AutoMigrate(&model.Book{})
	return db
}

func main() {
	// 1. 初始化 mysql + GORM
	db := initDB()
	// 2. 把 db 注入 dao 层，所有 dao 共用一个连接池
	dao.InitDao(db)

	r := gin.Default()

	bookroute := r.Group("/books")
	{
		bookroute.POST("", create)
		bookroute.PUT("/:id", renew)
		bookroute.GET("", getbooklist)
		bookroute.GET("/:id", detail)
		bookroute.DELETE("/:id", deleteBook)
	}

	if err := r.Run(":8080"); err != nil {
		panic("服务启动失败: " + err.Error())
	}
}
