package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
func main() {
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
