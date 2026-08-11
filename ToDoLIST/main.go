package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
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
	r.Run(":8080")
}
