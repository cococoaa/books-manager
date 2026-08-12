# ToDoLIST — Gin + GORM + MySQL 待办清单管理系统

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
        ├── js/
        └── fonts/
```

## API 接口

| 方法 | 路径 | 处理函数 | 说明 |
|------|------|----------|------|
| GET | `/` | 匿名函数 | 渲染 index.html 页面 |
| GET | `/todos` | `QueryALLTodo` | 获取全部待办 |
| GET | `/todos/:id` | `QueryTodo` | 获取单个待办 |
| POST | `/todos` | `CreateTodo` | 新增待办 |
| PUT | `/todos/:id` | `RenewTodo` | 更新待办 |
| DELETE | `/todos/:id` | `DeleteTodo` | 删除待办 |

---

## 问题总结

### 问题 1：package 名写错，编译器不认入口

**错误代码：**

```go
package todolist   // ❌ 入口文件必须是 package main
```

**错误信息：**

```
package command-line-arguments is not a main package
```

Go 编译器只会从 `package main` 里找 `func main()`。文件夹名可以随便取，但入口文件的 `package` 必须是 `main`。

**修正：**

```go
package main
```

---

### 问题 2：静态文件路径两处都错，页面空白

**错误代码：**

```go
r.Static("/dist/static", "static")
```

**问题分析：**

`r.Static(URL路径, 本地目录)` 两个参数都有问题：

| 参数 | 错误值 | 为什么错 |
|------|--------|----------|
| URL 路径 | `/dist/static` | HTML 里引用的是 `/static/css/...`，不匹配 |
| 本地目录 | `"static"` | 文件在 `dist/static/` 下，少了一层 `dist/` |

浏览器根据 HTML 里 `<link href="/static/css/app.css">` 请求 `/static/css/...`，但 Gin 没映射 `/static`，返回 404。即便映射对了，本地目录 `static/` 也不存在。

**修正：**

```go
r.Static("/static", "dist/static")   // URL /static → 本地 dist/static
r.StaticFile("/favicon.ico", "dist/templates/favicon.ico")
```

---

### 问题 3：launch.json 的 JSON 格式错误（缺逗号）

**错误代码：**

```json
"env": {
    "DB_BOOKS_DSN": "...",
    "DB_STUDENTS_DSN": "..."
    "DB_TODOLIST_DSN": "..."        ← 上一行末尾缺了逗号
}
```

JSON 里同一对象内的多个字段必须用逗号分隔。少一个逗号，VS Code 解析整个 `launch.json` 失败，F5 启动时所有配置都不生效。

---

### 问题 4：环境变量名还是旧的 `DB_DSN`（本次核心问题）

**错误代码：**

```go
dsn := os.Getenv("DB_DSN")   // ❌ 跟图书、学生项目共用同一个变量名
```

这个错误反复出现三次了——图书系统、学生系统、ToDoLIST 最初都读 `DB_DSN`。环境变量是进程级全局的，三个项目共用一个变量名，读到的一定是同一个值，必然有人连错库：

```
终端里 $env:DB_DSN 指向 /books
         ↓
三个项目 os.Getenv("DB_DSN") 都读到 /books
         ↓
学生系统的表建到了 books 库，ToDoLIST 也是
         ↓
报错: Table 'books.todos' doesn't exist (表建错库了)
```

**修正：**

```go
dsn := os.Getenv("DB_TODOLIST_DSN")  // 每个项目独立命名
```

launch.json 里也要配：

```json
"env": {
    "DB_BOOKS_DSN": "root:Zsh...@.../books?...",
    "DB_STUDENTS_DSN": "root:Zsh...@.../students?...",
    "DB_TODOLIST_DSN": "root:Zsh...@.../todolist?..."
}
```

**三个项目的环境变量对照：**

| 项目 | 环境变量名 | 指向库 | 表名 |
|------|-----------|--------|------|
| book-manager | `DB_BOOKS_DSN` | `books` | `books` |
| student-manager | `DB_STUDENTS_DSN` | `students` | `students` |
| ToDoLIST | `DB_TODOLIST_DSN` | `todolist` | `todos` |

---

### 问题 5：终端 go run vs F5 的环境变量来源不同

这是多次踩坑后的核心经验：

```
终端 go run:
  读当前终端 $env:XXX → 读系统环境变量 → 读代码默认值

VS Code F5:
  读 launch.json env → 读系统环境变量 → 读代码默认值
```

**两者互不相干。** launch.json 里配得再全，终端 `go run` 也拿不到。终端需要手动 `$env:DB_TODOLIST_DSN="..."` 或者等默认值兜底。

**推荐使用 F5 启动**——launch.json 里 `"program": "${fileDirname}"` 自动切到当前文件所在目录，`env` 自动注入，一步到位。

---

### 问题 6：go run 的工作目录不对导致模板加载失败

**错误信息：**

```
panic: html/template: pattern matches no files: `dist/templates/*`
```

**原因：** 在 `print/` 目录下执行 `go run d:/...ToDoLIST/main.go` 时，Go 程序的工作目录是 `print/`（当前 shell 所在目录），相对路径 `dist/templates/*` 就去找 `print/dist/templates/`——这个目录不存在。

**解决：**

```bash
cd 项目目录 && go run main.go    # 先 cd 进去
```

或者直接 F5 —— `${fileDirname}` 自动处理工作目录。

---

### 问题 7：端口 8080 被占

好几个项目都用 8080，每次切换项目要先关掉旧的：

```powershell
# 查看谁占了 8080
netstat -ano | findstr :8080
# 杀掉对应 PID
taskkill /F /PID 进程号
```

---

## 环境变量优先级链（总图）

```
launch.json "env": {...}        ← 最高优先级（仅 F5 启动时）
       │
       ▼ 没设或终端启动
终端 $env:XXX = "..."           ← 当前 PowerShell 会话
       │
       ▼ 没设
系统环境变量 [Environment]::Set     ← 永久生效
       │
       ▼ 没设
代码中的默认值                    ← 最低优先级
```

**排查口诀：** 先查 `os.Getenv()` 实际读到了什么值，再往下排查。

---

## 关键语法

### `r.Static(relativePath, root)` — 静态文件目录

```go
r.Static("/static", "dist/static")
//         ↑ URL 前缀    ↑ 本地目录
```

浏览器请求 `/static/css/app.css` → Gin 读 `dist/static/css/app.css`。

### `r.StaticFile(relativePath, filepath)` — 单个静态文件

```go
r.StaticFile("/favicon.ico", "dist/templates/favicon.ico")
```

### `r.LoadHTMLGlob(pattern)` — 加载模板

```go
r.LoadHTMLGlob("dist/templates/*")  // 加载目录下所有文件
```

### `c.HTML(code, name, data)` — 渲染模板

```go
c.HTML(http.StatusOK, "index.html", nil)
```

---

## PowerShell `$env:` 和 launch.json `env` 的区别

### 问题：第二行 `$env:DB_TODOLIST_DSN = "..."` 是在配置环境变量吗？

**是的。** `$env:XXX = "值"` 就是在当前 PowerShell 终端设置环境变量。

```powershell
$env:DB_TODOLIST_DSN = "root:Zsh1314520@tcp(127.0.0.1:3306)/todolist?..."
#                          └──────────────────────────────────────────┘
#                                       数据库连接字符串

# 设好之后：
go run main.go
# → 程序里 os.Getenv("DB_TODOLIST_DSN") 就能读到这个值
```

这和 launch.json 里配的 `env` 是一个意思，只是方式不同：

| 方式 | 语法 | 生效范围 |
|------|------|----------|
| PowerShell | `$env:XXX = "值"` | 当前终端会话，关掉窗口就没了 |
| launch.json | `"env": { "XXX": "值" }` | F5 启动时自动注入进程 |

### 为什么推荐 F5 而不是终端？

终端启动需要每次手动敲一遍 `$env:...`，换了窗口就失效。launch.json 配好之后，F5 启动自动注入，一劳永逸。

而且 launch.json 里 `"program": "${fileDirname}"` 会自动把工作目录切到当前文件所在目录，不会出现"在根目录执行导致相对路径找不到文件"的问题。

### F5 启动的原理

```
你按 F5
  → VS Code 读取 .vscode/launch.json
  → 展开 ${fileDirname} = 当前文件所在目录
  → 注入 "env" 里的所有环境变量 (set DB_TODOLIST_DSN=...)
  → 在该目录下执行 go build + 运行
  → 程序 os.Getenv("DB_TODOLIST_DSN") → 读到 launch.json 的值 ✅
```

整个过程**只在这次启动的进程内生效**，不影响系统其他任何地方，退出程序就没了，干净且不污染。

---

## 新建项目 checklist

以后每加一个项目，按这个清单逐项确认：

- [ ] `package main`（不是文件夹名）
- [ ] `os.Getenv("DB_项目名_DSN")`（独立命名，不共用 `DB_DSN`）
- [ ] launch.json 的 `env` 里补一行（注意逗号）
- [ ] launch.json 逗号语法正确
- [ ] `r.Static` 的 URL 路径和 HTML 里引用的一致
- [ ] `r.Static` 的本地目录有 `dist/` 前缀（如果文件在 dist 里）
- [ ] 按 F5 启动，而不是终端 `go run`
- [ ] 端口是否被占

---

## 前后端对接问题（本次核心）

### 问题 8：前后端 API 路径不匹配 → 前端所有请求 404

**错误现象：** 前端 `http://localhost:8080/#/` 页面正常显示，但所有的待办操作（添加、删除、修改）都失败。

**前端 JS 中发出的请求：**

```javascript
GET    /v1/todo          // 获取全部
POST   /v1/todo          // 新增
PUT    /v1/todo/:id      // 更新
DELETE /v1/todo/:id      // 删除
```

**后端最初的路由：**

```go
todoroute := r.Group("/todos")    // 匹配 /todos，不是 /v1/todo
```

**根因：** 这套前端模板是别人的 Vue 项目，API 设计是 `/v1/todo` + `title` + `status`。后端是自己写的 `/todos` + `Content`。两边完全对不上。

**修正：** 统一改为 `/v1/todo`，model 里 `Content` → `Title`，加 `Status bool` 字段。

---

### 问题 9：JSON 字段名不匹配 → 前端拿到数据但渲染不出来

**错误代码：**

```go
type Todo struct {
    gorm.Model
    Content string `gorm:"...;comment:'待办事情内容'"`
}
```

**前端发的 JSON：**

```json
{"title": "打游戏"}
```

**后端期望的 JSON：**

```json
{"Content": "打游戏"}
```

前端传 `title`，后端等 `Content`，Gin 的 `ShouldBindJSON` 解析不到值，存进数据库的是空字符串。

**修正：**

```go
Title string `gorm:"...;comment:'待办事情内容'" json:"title"`
```

---

### 问题 10：GET 返回 Response 包装对象 → el-table 渲染为空

**错误代码：**

```go
func QueryALLTodo(c *gin.Context) {
    todos, _ := dao.GetAllTodo()
    success(c, todos)    // → {"code":0,"msg":"success","data":[...]}
}
```

**前端代码：**

```javascript
this.axios.get("/v1/todo").then(function(t) {
    return e.tableData = t.data    // t.data 是 {code:0, msg:"success", data:[...]}
})
```

el-table 期望 `tableData` 是数组 `[{...}, {...}]`，但 `t.data` 是整个 Response 对象，不是数组。el-table 不会报错，但一行都渲染不出来。

**修正：**

```go
func QueryALLTodo(c *gin.Context) {
    todos, _ := dao.GetAllTodo()
    c.JSON(http.StatusOK, todos)   // 直接返回裸数组
}
```

---

### 问题 11：gorm.Model 字段无 json tag → 前端调 `row.id` 拿到 undefined

**现象：** 待办列表能显示了，但点击删除按钮没反应。

**排查：** 后端返回的 JSON：

```json
{"ID":1, "CreatedAt":"...", "UpdatedAt":"..."}
//  ↑ 大写的 ID，JSON 里是大写
```

前端 JS 调的是：

```javascript
handleDelete: function(e, t) {
    this.axios.delete("/v1/todo/" + t)   // t = row.id
    //                               ↑ 小写 id
}
```

JavaScript 严格区分大小写：`row.id` 去匹配 `"ID"` → `undefined` → 发送 `DELETE /v1/todo/undefined` → 400 错误。

**修正：** 不用 `gorm.Model`，自己定义每个字段并加 json tag：

```go
type Todo struct {
    ID        uint           `gorm:"primaryKey" json:"id"`         // id 小写
    CreatedAt time.Time      `json:"created_at"`                   // 蛇形
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`              // 隐藏
    Title     string         `gorm:"not null;comment:'待办事情内容'" json:"title"`
    Status    bool           `gorm:"not null;default:false;comment:'是否完成'" json:"status"`
}
```

**对比：**

| | gorm.Model（之前） | 自定义字段 + json tag（之后） |
|---|---|---|
| JSON 输出 | `"ID":1` | `"id":1` |
| 前端 `row.id` | `undefined` | `1` ✅ |
| 删除按钮 | 失效 | 正常 |

---

### 问题 12：AutoMigrate 只加列不删列 → Error 1364

**错误日志：**

```
Error 1364 (HY000): Field 'content' doesn't have a default value
INSERT INTO `todos` (`created_at`,`updated_at`,`deleted_at`,`title`,`status`)
VALUES ('2026-08-12 21:02:26',...,NULL,'打游戏',false)
```

**根因追踪：**

```
第一版 model:  Content string → AutoMigrate 建表: id, content, ...
第二版 model:  Title  string → AutoMigrate 行为: 保留 content 列 + 新增 title 列
               
新 INSERT: INSERT INTO todos (title) VALUES ('打游戏')
           → content 列 NOT NULL 且无默认值 → MySQL 拒绝 → Error 1364
```

**AutoMigrate 的核心设计：**
- ✅ 新增列
- ✅ 修改列类型/约束
- ✅ 新增索引
- ❌ **从不删列**（怕数据丢失）
- ❌ 从不改列名（认为是"删旧列 + 加新列"）

改字段名 = 残留旧列。社区推荐的生产环境做法是手写 SQL 迁移脚本。

**修正：** 在 `initDB` 的 `AutoMigrate` 之前加清理逻辑：

```go
// 删除旧版本的 content 列（从 Content 改名 Title 后残留）
if db.Migrator().HasColumn(&model.Todo{}, "content") {
    db.Migrator().DropColumn(&model.Todo{}, "content")
}
// 然后再建表
db.AutoMigrate(&model.Todo{})
```

---

### 前后端对接总结

**前端模板里硬编码了三样东西，后端每一个都得对上：**

| 前端期望 | 后端第一版 | 后端最终版 | 
|----------|-----------|-----------|
| URL 路径 | `/todos` ❌ | `/v1/todo` ✅ |
| JSON 字段名 | `Content` ❌ | `Title` → `"title"` ✅ |
| 状态字段 | 无 ❌ | `Status` → `"status"` ✅ |
| 返回格式 | `{code, msg, data}` ❌ | 裸数组 `[...]` ✅ |
| ID 字段 JSON | `"ID"` ❌ | `"id"` ✅ |
| 旧列 content | 残留 ❌ | 启动时自动删 ✅ |


