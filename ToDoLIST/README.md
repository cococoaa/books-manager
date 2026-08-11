# ToDoLIST — Gin 静态文件未加载问题

## 项目结构

```
ToDoLIST/
├── main.go
├── dao/
│   └── dao.go
├── model/
│   └── model.go
└── dist/
    ├── templates/
    │   ├── favicon.ico
    │   └── index.html
    └── static/
        ├── css/
        │   ├── app.8eeeaf31.css
        │   └── chunk-vendors.57db8905.css
        ├── js/
        │   ├── app.007f9690.js
        │   └── chunk-vendors.ddcb6f91.js
        └── fonts/
            ├── element-icons.535877f5.woff
            └── element-icons.732389de.ttf
```

## 错误现象

`go run main.go` 启动后，浏览器访问 `http://localhost:8080`，页面一片空白，CSS 和 JS 全部加载失败。

## 错误代码

```go
func main() {
    r := gin.Default()
    r.Static("/dist/static", "static")     // ❌ 两处都错
    r.LoadHTMLGlob("dist/templates/*")
    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })
    r.Run(":8080")
}
```

## 错误分析

### 错误 1：本地目录路径不存在

```go
r.Static("/dist/static", "static")
//  ↑ URL 前缀           ↑ 本地目录
```

**`r.Static(relativePath, root)` 的参数含义：**

| 参数 | 说明 |
|------|------|
| `relativePath` | 浏览器访问的 URL 路径 |
| `root` | 本地磁盘上的文件夹路径 |

代码写的是 `"static"`，程序从 `ToDoLIST/` 目录下找 `static/` 文件夹。

实际文件在 `ToDoLIST/dist/static/` 下面，多了一层 `dist/`。所以 Gin 永远找不到 CSS 和 JS 文件。

### 错误 2：URL 路径与 HTML 中引用的路径不匹配

index.html 里是这样引用资源的：

```html
<link href="/static/css/app.8eeeaf31.css" rel="stylesheet">
<script src="/static/js/app.007f9690.js"></script>
```

浏览器请求的 URL 是 `/static/css/...`，但 Gin 映射的 URL 是 `/dist/static`：

```
浏览器请求:  GET /static/css/app.8eeeaf31.css   ← HTML 写的
Gin 监听:    r.Static("/dist/static", ...)       ← 不匹配 /static
             结果：404 找不到
```

## 正确代码

```go
func main() {
    r := gin.Default()

    // 将 /static URL 映射到本地 dist/static 目录
    r.Static("/static", "dist/static")

    // 单独映射 favicon（它在 templates 目录下，不是 static）
    r.StaticFile("/favicon.ico", "dist/templates/favicon.ico")

    // 告诉 Gin 去哪里找 HTML 模板
    r.LoadHTMLGlob("dist/templates/*")

    r.GET("/", func(c *gin.Context) {
        c.HTML(http.StatusOK, "index.html", nil)
    })

    r.Run(":8080")
}
```

## 关键语法讲解

### `r.Static(relativePath, root)` — 静态文件服务

```go
r.Static("/static", "dist/static")
//         ↑ URL 路径    ↑ 本地相对路径
```

当浏览器请求 `/static/css/app.css` 时，Gin 去 `dist/static/css/app.css` 读文件。

**一定要两边对上：**

```
URL 请求:    /static/css/app.css
                ↓ r.Static 匹配 /static
本地文件:  dist/static/css/app.css
                ↓ 根目录是 dist/static
最终读取:     css/app.css  → 拼到 dist/static/ 后面
```

### `r.StaticFile(relativePath, filepath)` — 单个静态文件

```go
r.StaticFile("/favicon.ico", "dist/templates/favicon.ico")
//            ↑ 浏览器请求的 URL  ↑ 本地文件路径
```

和 `r.Static` 不同，`StaticFile` 只映射一个文件。favicon 放在 `templates` 目录下，不在 `static` 里，所以需要单独处理。

### `r.LoadHTMLGlob(pattern)` — 加载 HTML 模板

```go
r.LoadHTMLGlob("dist/templates/*")
//                     ↑ 通配符，匹配目录下所有文件
```

Gin 会在启动时一次性加载所有匹配的 HTML 文件到内存，`c.HTML()` 时直接渲染。

### `c.HTML(code, name, data)` — 渲染 HTML 模板

```go
c.HTML(http.StatusOK,   // HTTP 状态码
       "index.html",    // 模板文件名（LoadHTMLGlob 加载的）
       nil)             // 模板数据，没有时传 nil
```

## 排查静态文件 404 的口诀

```
1. 浏览器 F12 → Network 看哪个文件 404
2. 404 的 URL 路径是什么？→ 核对 r.Static 的 relativePath
3. 本地文件在哪个目录？→ 核对 r.Static 的 root
4. HTML 里引用的路径和你 Gin 映射的路径对上了吗？
```

## 补充：之前还遇到了 package 名写错的问题

```go
package todolist   // ❌ 入口文件必须是 package main
```

Go 编译器只会从 `package main` 里找 `func main()` 作为程序入口。文件夹里所有 `.go` 文件可以有不同的 package，但必须有一个 `package main` 包含 `func main()`。
