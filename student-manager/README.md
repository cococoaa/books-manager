# Gin + GORM 学生管理系统

## 项目结构

```
student-manager/
├── main.go          # 入口 + HTTP 路由 + Handler 层
├── model/
│   └── mod.go       # 数据模型定义
└── dao/
    └── dao.go       # 数据库访问层
```

## API 接口

| 方法 | 路径 | 处理函数 | 说明 |
|------|------|----------|------|
| GET | `/students` | `getstudentlist` | 获取全部学生 |
| GET | `/students/:id` | `detail` | 获取单个学生 |
| POST | `/students` | `create` | 新增学生 |
| PUT | `/students/:id` | `renew` | 更新学生 |
| DELETE | `/students/:id` | `deleteBook` | 删除学生 |

## 编译/语法类问题

### 1. package 名与文件夹名不一致

**文件**：`model/mod.go`
**错误**：`package mod`

文件夹名叫 `model`，`package` 就必须是 `model`。导入路径用的是文件夹名，`package` 声明的是命名空间——两者必须一致，否则引用方写 `model.Student` 会报错。

### 2. import 路径少了一层目录

**文件**：`dao/dao.go`
**错误**：`"dgo.baisic.print/studentguanlixitong/model"`

`go.mod` 在 `print/` 目录下，`studentguanlixitong` 还嵌套在 `baisic/` 里，正确的路径是：

```
"dgo.baisic.print/baisic/studentguanlixitong/model"
```

规则：**import 路径 = module名 + 相对于 go.mod 所在目录的文件路径。**

### 3. import 了两个项目的包导致冲突

**文件**：`main.go`

```go
import (
    "dgo.baisic.print/GIN/dao"                          // ← 图书管理系统的
    "dgo.baisic.print/GIN/model"                        // ← 图书管理系统的
    "dgo.baisic.print/baisic/studentguanlixitong/dao"   // ← 本项目的
    "dgo.baisic.print/baisic/studentguanlixitong/model" // ← 本项目的
)
```

两个 `dao`、两个 `model` 导入路径不同但包名相同，编译器分不清谁是谁。

### 4. Gin Handler 签名错误

```go
func detail(id int, c *gin.Context)  { ... }    // ❌
func deleteBook(id int, c *gin.Context) { ... }  // ❌
```

Gin 的 handler 签名必须是 `func(*gin.Context)`，不能自定义参数。路由中的 `:id` 应该在函数内通过 `c.Param("id")` 获取。

### 5. c.Param 返回 string，DAO 需要 int

路由 `/students/:id` → `c.Param("id")` 返回字符串 `"1"`，但 `dao.GETStudentbyID(id int)` 要求 `int` 类型。使用 `strconv.Atoi()` 转换。

### 6. gorm.Model 与自定义 ID 字段重复

`gorm.Model` 自带 `ID uint`（自增主键），又定义了 `ID int` 导致冲突。如果要用自增 ID，直接删掉自定义的 `ID` 字段。

### 7. DSN loc 参数未 URL 编码

`loc=Asia/Shanghai` 中的 `/` 被 MySQL DSN 解析为路径分隔符，改为 `loc=Asia%2FShanghai`。

## 环境变量与多项目配置（重点）

这是多项目共存时最隐蔽的一类问题，也是本项目调试耗时最长的问题。

### 8. 两个项目共用一个 `DB_DSN` 导致建表串库

**现象**：学生管理系统的 POST 报错 `Table 'books.students' doesn't exist`。明明代码里默认 DSN 写的是 `/students`，为什么 GORM 却往 `books` 库写？

**排查过程**：

```
代码默认值: /students    ← 以为是这个
系统环境变量: (空)         ← 第 3 方案设的已删掉
终端 $env:DB_DSN = ?     ← 之前设过，还挂在终端里！
launch.json: DB_BOOKS_DSN + DB_STUDENTS_DSN  ← F5 能读到，终端读不到
```

**根因**：图书管理系统和学生管理系统最初都读 `DB_DSN` 这一个变量。终端里设过 `$env:DB_DSN = .../books`，学生项目启动时 `os.Getenv("DB_DSN")` 读到的还是 `books`，默认值被环境变量覆盖了。

**解决**：拆成两个独立的环境变量：

| 项目 | 环境变量名 | 指向 |
|------|-----------|------|
| book-manager | `DB_BOOKS_DSN` | `/books` |
| student-manager | `DB_STUDENTS_DSN` | `/students` |

### 9. AutoMigrate 是启动时执行，不是 POST 请求时

`db.AutoMigrate(&model.Student{})` 在 `go run main.go` 那一刻就执行了，不是在收到第一个 POST 请求时。所以表建没建成，看启动日志就知道了。如果连接失败，服务直接 `log.Fatal` 退出，根本不会监听端口。

### 10. 终端 go run vs F5 启动的环境变量来源不同

| 启动方式 | 环境变量来源 |
|----------|-------------|
| F5 (VS Code) | launch.json 的 `"env": {...}` |
| 终端 `go run` | 当前终端 `$env:XXX` + 系统环境变量 |

**两者互不相干。** 在终端启动需要手动设 `$env:DB_STUDENTS_DSN="..."`，F5 则是 launch.json 自动注入。

### 11. 环境变量优先级链

```
launch.json env > 终端 $env:XXX > 系统环境变量 > 代码默认值
```

一个变量被上层设过后（哪怕设错了），下层默认值永远不会生效。排查建表问题第一步先确认 `os.Getenv()` 到底读到了什么值。

> **核心经验**：多项目共存时，每个项目用独立的环境变量名。先从 `os.Getenv` 返回值反向确认连到了哪个库，再查 GORM 的建表和插入行为。

## c.Param vs uuid.NewString 的取舍

| | c.Param("id") | uuid.NewString() |
|---|---|---|
| 数据来源 | URL 路径 `/students/1` | 服务端自动生成 |
| ID 类型 | 自增 ID (int)，由数据库管理 | UUID (string)，由应用层生成 |
| 冲突风险 | 无（数据库自增保证唯一） | 极低 |
| 安全性 | 可遍历 `1,2,3...` | 不可预测 |
| 适用场景 | 内部系统、学习项目 | 公开 API、需要隐藏 ID 数量 |

本项目使用自增 ID + `c.Param("id")`。

## JWT 登录功能遇到的问题

### 12. User 表字段类型错误：string 默认映射成 longtext

**错误信息**：

```
Error 1170 (42000): BLOB/TEXT column 'username' used in key specification without a key length
```

**根因**：Go 的 `string` 字段如果没写 `gorm:"type:..."`，GORM 默认映射成 `longtext` 类型。而 `longtext` 不能直接作为唯一索引（`uniqueIndex`）的键，MySQL 拒绝建表。

**解决**：给字符串字段显式指定 `type:varchar(长度)`：

```go
Username string `gorm:"type:varchar(64);not null;uniqueIndex;comment:'用户名'" json:"username"`
Password string `gorm:"type:varchar(255);not null;comment:'密码'" json:"-"`
```

> 经验：凡是加 `uniqueIndex` 或 `index` 的字符串字段，都要写 `type:varchar(...)`，不能用默认的 longtext。

### 13. 新增学生失败：time 字段零值问题

**错误信息**：

```
Error 1292 (22007): Incorrect datetime value: '0000-00-00' for column 'time'
```

**根因**：`Student.Time` 字段是 `time.Time` 类型且标注 `not null`。前端新增学生时没传 `time`，GORM 用零值 `0000-00-00` 插入，MySQL 的 datetime 类型不接受这个值。

**解决（两步）**：

1. model 里去掉 `not null`，加 json tag：

```go
Time  time.Time `gorm:"comment:'录入时间'" json:"time"`
```

2. create handler 里判断零值，自动填当前时间：

```go
if req.Time.IsZero() {
    req.Time = time.Now()
}
```

> 经验：`time.Time` 字段若允许空，去掉 `not null`；若语义是"创建/录入时间"，服务端自动生成，别让前端传。

### 14. 残留 __debug_bin 进程占用端口

**现象**：改了代码、重启服务，但接口行为还是旧的，报错依旧。

**根因**：按 F5 调试时，VS Code 会启动 `__debug_bin.exe` 进程。调试结束（点停止/重启 VS Code）后，这个进程有时不会自动退出，继续占着 8080 端口。之后 `go run` 新代码要么启动失败（端口被占），要么你以为重启了其实还在请求旧进程。

**排查**：

```powershell
# 查端口被谁占
Get-NetTCPConnection -LocalPort 8080 | Select-Object OwningProcess
# 看进程名
Get-Process -Id <PID> | Select-Object ProcessName
# 看到 __debug_bin 就是残留调试进程，杀掉
taskkill /F /PID <PID>
```

> 经验：改了代码不生效，先查端口是不是被 `__debug_bin` 残留进程占了，杀干净再启动。

## JWT 登录功能说明

- **注册** `POST /register`：用户名 + 密码，bcrypt 加密后存库
- **登录** `POST /login`：校验密码，返回 JWT token（24 小时有效）
- **鉴权中间件**：`/students` 所有接口需带 `Authorization: Bearer <token>`
- **前端** `static/index.html`：登录/注册 + 学生增删查界面，token 存 localStorage

## 项目结构（仓库级）

```
books-manager/
├── book-manager/          ← Gin + GORM 图书管理系统
│   ├── main.go
│   ├── dao/dao.go
│   └── model/model.go
└── student-manager/       ← Gin + GORM 学生管理系统
    ├── main.go
    ├── dao/dao.go
    └── model/mod.go

根目录 .vscode/launch.json 配置了两个项目的环境变量
```
