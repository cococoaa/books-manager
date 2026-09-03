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

## 二、Gin → GORM → MySQL 各层关系和职责

用这个项目画一张图：



```
┌─────────────────────────────────────────────────────────┐
│  浏览器 / curl                                          │
│  POST /books  {"title":"三体","author":"刘慈欣","year":2008} │
└──────────────────────┬──────────────────────────────────┘
                       │ HTTP 请求
                       ▼
┌─────────────────────────────────────────────────────────┐
│  Gin (main.go)        ← 接口层 / Handler 层              │
│                                                         │
│  职责：接收请求、参数校验、调用 DAO、返回响应              │
│  不关心：SQL 怎么写、数据库在哪                          │
│                                                         │
│  func create(c *gin.Context) {                          │
│      var req model.Book         ← 用 model 包的结构体    │
│      c.ShouldBindJSON(&req)     ← Gin 帮你解析 JSON      │
│      dao.CreateBook(&req)       ← 扔给 DAO，具体怎么存不管│
│      success(c, req)            ← 返回 JSON 响应         │
│  }                                                      │
└──────────────────────┬──────────────────────────────────┘
                       │ 函数调用
                       ▼
┌─────────────────────────────────────────────────────────┐
│  model/model.go        ← 数据模型层                       │
│                                                         │
│  职责：定义数据长什么样（Go struct ↔ 数据库表）            │
│  不关心：怎么查、怎么存                                   │
│                                                         │
│  type Book struct {                                     │
│      gorm.Model                                         │
│      ID     string `gorm:"not null"`                    │
│      Title  string `gorm:"not null"`   ← 跟数据库列一一对应│
│      Author string                                     │
│      Year   int                                        │
│  }                                                      │
└──────────────────────┬──────────────────────────────────┘
                       │ DAO 引用 model.Book
                       ▼
┌─────────────────────────────────────────────────────────┐
│  dao/dao.go            ← 数据访问层                       │
│                                                         │
│  职责：只管数据库的增删改查，封装所有 SQL（GORM）操作       │
│  不关心：HTTP 请求、JSON 解析                            │
│                                                         │
│  func CreateBook(book *model.Book) error {               │
│      return db.Create(book).Error     ← GORM 帮你生成 SQL│
│  }                                                      │
│  func GetBookById(id string) (*model.Book, error) {      │
│      return db.Where("id=?", id).First(&b).Error         │
│  }                                                      │
└──────────────────────┬──────────────────────────────────┘
                       │ GORM 翻译成 SQL
                       ▼
┌─────────────────────────────────────────────────────────┐
│  GORM                   ← ORM 框架                       │
│                                                         │
│  职责：Go 代码 ↔ SQL 的翻译官                            │
│  db.Create(&book)  →  INSERT INTO books (...) VALUES (..)│
│  db.First(&book)   →  SELECT * FROM books WHERE ...      │
└──────────────────────┬──────────────────────────────────┘
                       │ 标准 database/sql 驱动
                       ▼
┌─────────────────────────────────────────────────────────┐
│  MySQL                  ← 数据库                          │
│                                                         │
│  磁盘上真正存数据的地方，重启也不会丢                      │
└─────────────────────────────────────────────────────────┘
```

## 三、一条请求怎么走完全程

以 **新增图书** `POST /books` 为例：



```
1. Gin 收到 POST /books
   → c.ShouldBindJSON(&req)  把 JSON 转成 model.Book 结构体
   → req.ID = uuid.NewString()  服务端生成唯一 ID

2. Gin 调 dao.CreateBook(&req)
   → DAO 不关心数据哪来的，只管存

3. DAO 调 db.Create(book)
   → GORM 翻译成 SQL: INSERT INTO books (id, title, author, year) VALUES (...)

4. GORM 通过 mysql driver 把 SQL 发给 MySQL
   → MySQL 写入磁盘，返回成功

5. 原路返回
   → DAO 返回 nil（没出错）
   → Gin 调 success(c, req)
   → 浏览器收到 {"code":0,"msg":"success","data":{...}}
```

## 四、为什么要分三层

**不分层会怎样？** 你最初代码就是这样——所有东西塞在 main.go：



```go
// Handler + 模型 + 数据库操作 + 业务逻辑全混在一起
func create(c *gin.Context) {
    var req book              // 自己定义的模型
    booklist = append(...)    // 直接操作内存 slice
    // 换个数据库？全部重写
}
```

**分层以后：**

- **换数据库**（MySQL → PostgreSQL）？只改 DAO 层，Handler 一行不动
- **加字段**？只改 model 层，其他地方自动生效
- **改业务规则**（比如新增图书要发通知）？加一个 service 层，DAO 和 Handler 都不动
- **写测试**？每一层可以独立 mock，不用真的连数据库测

每一层只管自己的事，改动的影响被限制在一层内，这就是分层的意义。

哪怕现在规模小，DA之层实际的职责不是"一行代码的厚度"，而是**边界**：

1. **换库** — 以后从 GORM 切到 sqlc 或者 raw SQL，只改 DAO，Handler 不动。DAO 文件就是你的修改清单。
2. **加缓存/日志** — 在 `GetAllBooks` 里加 Redis 缓存，改一个 DAO 函数就够了，不用到 Handler 里到处翻。
3. **不同项目需要不同的库** — book-manager 和 student-manager 接到不同的数据库，现在靠各自 DAO 注入各自的 `db`，合并后这个边界变模糊了。
4. **测试** — 测 DAO 只需要 mock 数据库，测 Handler + DAO 就要 mock HTTP 请求 + 数据库。
