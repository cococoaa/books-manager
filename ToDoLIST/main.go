package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"dgo.baisic.print/GIN/ToDoLIST/dao"
	"dgo.baisic.print/GIN/ToDoLIST/model"
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
func CreateTodo(c *gin.Context) {
	var req model.Todo
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数校验失败: "+err.Error())
		return
	}
	if err := dao.CreateTodo(&req); err != nil {
		Fail(c, 500, "新增代办事项失败: "+err.Error())
		return
	}
	success(c, req)
}
func QueryTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, 400, "无效的ID")
		return
	}
	Todo, err := dao.GetTodoById(id)
	if err != nil {
		Fail(c, 400, "待办不存在")
		return
	}
	success(c, Todo)
}
func QueryALLTodo(c *gin.Context) {
	todos, err := dao.GetAllTodo()
	if err != nil {
		Fail(c, 500, "查询待办列表失败: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, todos)
}
func DeleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, 400, "无效的ID")
		return
	}
	if _, err := dao.GetTodoById(id); err != nil {
		Fail(c, 400, "待办不存在")
		return
	}
	if err := dao.DeleteTodo(id); err != nil {
		Fail(c, 500, "删除待办失败: "+err.Error())
		return
	}
	success(c, nil)
}
func RenewTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		Fail(c, 400, "无效的ID")
		return
	}
	todo, err := dao.GetTodoById(id)
	if err != nil {
		Fail(c, 400, "待办不存在")
		return
	}
	var req model.Todo
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, 400, "参数错误: "+err.Error())
		return
	}
	todo.Title = req.Title
	todo.Status = req.Status
	if err := dao.UpdateTodo(todo); err != nil {
		Fail(c, 500, "更新待办信息失败: "+err.Error())
		return
	}
	success(c, todo)
}
func initDB() *gorm.DB {
	dsn := os.Getenv("DB_TODOLIST_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/todolist?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败: ", err)
	}
	// 自动建表
	db.AutoMigrate(&model.Todo{})
	return db
}
func main() {
	// 1. 初始化 mysql + GORM
	db := initDB()
	// 2. 把 db 注入 dao 层，所有 dao 共用一个连接池
	dao.InitDao(db)

	r := gin.Default()
	//告诉GIN框架模板文件引用的静态文件去哪里找
	// 将 /static 映射到 dist/static 目录
	r.Static("/static", "dist/static")
	// 单独映射 favicon
	r.StaticFile("/favicon.ico", "dist/templates/favicon.ico")
	// 告诉GIN框架去哪里找模板文件
	r.LoadHTMLGlob("dist/templates/*")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	todoroute := r.Group("/v1/todo")
	{
		todoroute.POST("", CreateTodo)
		todoroute.PUT("/:id", RenewTodo)
		todoroute.GET("", QueryALLTodo)
		todoroute.GET("/:id", QueryTodo)
		todoroute.DELETE("/:id", DeleteTodo)
	}
	if err := r.Run(":8080"); err != nil {
		panic("服务启动失败: " + err.Error())
	}
}
