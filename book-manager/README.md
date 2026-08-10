# Gin + GORM + MySQL 图书管理系统

## 项目结构

```
GIN/
├── main.go          # 入口 + HTTP 路由 + Handler 层
├── model/
│   └── model.go     # 数据模型定义
└── dao/
    └── dao.go       # 数据库访问层 (Data Access Object)
```

## 技术栈

| 层级 | 技术 | 作用 |
|------|------|------|
| HTTP 框架 | Gin | 路由、请求解析、响应返回 |
| ORM | GORM | Go 结构体 ↔ SQL 的翻译层 |
| 数据库 | MySQL | 持久化存储 |

## 分层职责

### 1. model 层 — 数据模型

**只做一件事：定义数据长什么样。** Go struct 与数据库表一一对应。

```go
type Book struct {
    gorm.Model
    ID     string `gorm:"not null;comment:'图书ID'"`
    Title  string `gorm:"not null;comment:'图书名称'"`
    Author string `gorm:"not null;comment:'作者'"`
    Year   int    `gorm:"not null;comment:'图书出版日期'"`
}
```

- `gorm.Model` 内嵌了 `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt` 四个字段
- `` `gorm:"..."` `` struct tag 告诉 GORM 列的类型、约束和注释
- model 包不关心 HTTP 请求、不关心数据库连接，它只定义结构

### 2. dao 层 — 数据访问

**只做一件事：封装所有数据库操作。** 每个函数对应一次 SQL 操作。

| 函数 | 对应 SQL | 说明 |
|------|----------|------|
| `CreateBook(book)` | `INSERT INTO books (...) VALUES (...)` | 新增 |
| `GetBookById(id)` | `SELECT * FROM books WHERE id = ?` | 单条查询 |
| `GetAllBooks()` | `SELECT * FROM books` | 全部查询 |
| `UpdateBook(book)` | `UPDATE books SET ... WHERE id = ?` | 更新 |
| `DeleteBook(id)` | `DELETE FROM books WHERE id = ?` | 删除 |

全局 `db` 变量由 `InitDao()` 注入，整个 dao 包共用一个数据库连接池：

```go
var db *gorm.DB

func InitDao(database *gorm.DB) {
    db = database
}
```

### 3. main.go (Handler 层) — 请求入口

**只做一件事：接收 HTTP 请求、调 dao、返回响应。** 不写 SQL，不碰数据库连接。

```go
func create(c *gin.Context) {
    var req model.Book         // 用 model 包的结构体
    c.ShouldBindJSON(&req)     // Gin 解析请求体 JSON
    req.ID = uuid.NewString()  // 服务端生成唯一 ID
    dao.CreateBook(&req)       // 调用 DAO 层存库
    success(c, req)            // 返回 JSON 响应
}
```

## 一条请求的全流程

以 `POST /books` 为例：

```
浏览器/curl
  │  POST /books  {"title":"三体","author":"刘慈欣","year":2008}
  ▼
Gin Router
  │  匹配路由 → 调用 create()
  │  c.ShouldBindJSON(&req)  把 JSON 解析成 model.Book
  │  uuid.NewString()        生成唯一 ID
  ▼
dao.CreateBook(&req)
  │  db.Create(book)         调 GORM
  ▼
GORM
  │  翻译成 SQL: INSERT INTO books (id, title, author, year) VALUES (...)
  ▼
MySQL 驱动
  │  通过 database/sql 标准接口发送 SQL
  ▼
MySQL
  │  写入磁盘，提交事务
  ●  原路返回 → 浏览器收到 {"code":0,"msg":"success","data":{...}}
```

## API 接口

| 方法 | 路径 | 处理函数 | 说明 |
|------|------|----------|------|
| GET | `/books` | `getbooklist` | 获取全部图书 |
| GET | `/books/:id` | `detail` | 获取单本图书 |
| POST | `/books` | `create` | 新增图书 |
| PUT | `/books/:id` | `renew` | 更新图书 |
| DELETE | `/books/:id` | `deleteBook` | 删除图书 |

## 开发中遇到的问题

### 编译/语法类

| # | 位置 | 问题 | 类型 |
|---|------|------|------|
| 1 | model.go | `Title` 字段 struct tag 缺闭合双引号 | 语法错误 |
| 2 | dao.go | `import "project/model"` — 模块路径不存在 | 编译错误 |
| 3 | dao.go | `createBook` 小写开头，外部包无法调用 | 可见性错误 |
| 4 | dao.go | 注释复制粘贴未修改 | 代码规范 |
| 5 | dao.go | 只有 Create 和 Query，缺 Update 和 Delete | 功能缺失 |
| 6 | main.go | 未 import `gorm.io/driver/mysql` 却使用 `mysql.Open()` | 编译错误 |
| 7 | main.go | `AutoMigrate(&model.User{})` — User 类型不存在 | 编译错误 |
| 8 | main.go | main 包内重复定义 `book` struct，与 `model.Book` 冲突 | 重复定义 |
| 9 | main.go | Handler 直接操作内存 slice，DAO 层形同虚设 | 分层失效 |
| 10 | main.go | `delete` 函数名与内置函数重名 | 命名冲突 |
| 11 | main.go | `Fail(c, 400, msg)` 未加 `return`，代码向下穿透导致重复响应 | 逻辑错误 |
| 12 | main.go | `r.Group("/books")` 花括号未另起一行 | 语法错误 |
| 13 | main.go | 包级别使用 `:=` 短声明 | 语法错误 |
| 14 | main.go | PowerShell 中使用 `curl` 被别名劫持为 `Invoke-WebRequest` | 环境问题 |
| 15 | main.go | JSON 字段注释 `json` 误写为 `josn` | 拼写错误 |
| 16 | main.go | DSN 中 `loc=Asia/Shanghai` 的 `/` 未 URL 编码，导致连接失败 | 配置错误 |

### 环境变量与多项目配置（重点）

**这是多项目共存时最隐蔽的一类问题。**

#### 问题 17：两个项目共用一个 `DB_DSN` 导致建表串库

图书管理系统和学生管理系统最初都读 `DB_DSN`。图书项目把 `DB_DSN` 指向了 `books` 库，学生项目启动时也读到同一个值，导致学生表建到了 `books` 库下面。

**根因**：`os.Getenv("DB_DSN")` 是进程级别全局的，谁设了都一样。

**解决**：拆成两个独立的环境变量：
- `DB_BOOKS_DSN` → 图书管理系统 → `books` 库
- `DB_STUDENTS_DSN` → 学生管理系统 → `students` 库

#### 问题 18：终端 go run 不读 launch.json

VS Code 的 `launch.json` 里配好的 `env` **只在按 F5 启动时生效**。如果从终端直接 `go run main.go`，进程拿不到 launch.json 里的环境变量，只能拿到终端会话里 `$env:XXX=` 设置的值。

#### 问题 19：环境变量优先级链

```
launch.json env > 终端 $env:XXX > 系统环境变量 > 代码默认值
```

一个变量如果被设过（哪怕设错了），默认值永远不会生效。排查时先确认 `os.Getenv()` 到底读到了什么。

> **关键经验**：多项目共存时，不要共用环境变量名。排查建表问题先确认 DSN 实际指向了哪个库。

## 为什么用服务端生成 UUID，而不允许客户端传 ID？

| | 客户端传 ID | 服务端生成 UUID |
|---|---|---|
| ID 冲突 | 容易冲突 (多个客户端传相同 ID) | 全局唯一 |
| 安全性 | 可遍历猜测，越权访问 | 不可预测 |
| 职责划分 | 客户端不关心 ID 生成 | 服务端统一管理 |

客户端传 ID 的唯一场景：更新/删除时用服务端之前返回的 ID（URL 中的 `:id`）。

## 为什么分层？

- **换数据库**（MySQL → PostgreSQL）：只改 DAO 层，Handler 一行不动
- **加字段**：只改 model 层
- **加业务规则**：加 service 层，DAO 和 Handler 都不动
- **写测试**：每层可以独立 mock，不依赖真实数据库

每一层只管自己的事，改动影响被限制在一层内。
