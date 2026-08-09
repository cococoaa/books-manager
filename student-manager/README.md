# Gin + GORM 学生管理系统

## 项目结构

```
studentguanlixitong/
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

## 遇到的问题

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
**错误**：

```go
import (
    "dgo.baisic.print/GIN/dao"                          // ← 图书管理系统的
    "dgo.baisic.print/GIN/model"                        // ← 图书管理系统的
    "dgo.baisic.print/baisic/studentguanlixitong/dao"   // ← 本项目的
    "dgo.baisic.print/baisic/studentguanlixitong/model" // ← 本项目的
)
```

两个 `dao`、两个 `model` 导入路径不同但包名相同，编译器分不清谁是谁。本项目只保留自己的 `dao` 和 `model`。

### 4. Gin Handler 签名错误

**文件**：`main.go`
**错误**：

```go
func detail(id int, c *gin.Context)  { ... }
func deleteBook(id int, c *gin.Context) { ... }
```

Gin 的 handler 签名必须是 `func(*gin.Context)`，不能自定义参数。路由中的 `:id` 应该在函数内通过 `c.Param("id")` 获取，而不是从参数传入。

### 5. c.Param 返回 string，DAO 需要 int

**文件**：`main.go`
**错误**：路由 `/students/:id` → `c.Param("id")` 返回字符串 `"1"`，但 `dao.GETStudentbyID(id int)` 要求 `int` 类型。

**解决**：使用 `strconv.Atoi()` 转换：

```go
id, err := strconv.Atoi(c.Param("id"))
```

### 6. gorm.Model 与自定义 ID 字段重复

**文件**：`model/mod.go`
**错误**：

```go
type Student struct {
    gorm.Model           // ← 自带 ID uint（自增主键）
    ID    int   `gorm:"..."` // ← 又定义了一个 ID，冲突
    ...
}
```

`gorm.Model` 已经包含 `ID`、`CreatedAt`、`UpdatedAt`、`DeletedAt`，不需要再定义 ID。如果要用自增 ID，直接删掉自定义的 `ID` 字段即可。

### 7. DSN loc 参数未 URL 编码

**错误**：`loc=Asia/Shanghai` 中的 `/` 被 MySQL DSN 解析为路径分隔符，导致连接失败。

**解决**：`/` 改为 `%2F`：`loc=Asia%2FShanghai`。

### 8. 密码硬编码在代码中

**解决**：改为环境变量 `DB_DSN`，避免密码提交到 Git。

### 9. 根目录非 git 仓库导致 VS Code 调试编译失败

**错误**：`error obtaining VCS status`

**解决**：`.vscode/launch.json` 加 `"buildFlags": "-buildvcs=false"`。

## c.Param vs uuid.NewString 的取舍

| | c.Param("id") | uuid.NewString() |
|---|---|---|
| 数据来源 | URL 路径 `/students/1` | 服务端自动生成 |
| ID 类型 | 自增 ID (int)，由数据库管理 | UUID (string)，由应用层生成 |
| 冲突风险 | 无（数据库自增保证唯一） | 极低 |
| 安全性 | 可遍历 `1,2,3...` | 不可预测 |
| 适用场景 | 内部系统、学习项目 | 公开 API、需要隐藏 ID 数量 |

本项目使用自增 ID + `c.Param("id")`，Url 路径中的 ID 由 `gorm.Model` 自增生成，新增时不需要客户端传 ID。
