# 5.4 API 设计

## RESTful API 设计规范

```go
// 资源命名 - 使用名词复数
// GET    /api/users      - 获取用户列表
// POST   /api/users      - 创建用户
// GET    /api/users/123  - 获取单个用户
// PUT    /api/users/123  - 更新用户
// DELETE /api/users/123  - 删除用户

r.GET("/api/users", func(c *gin.Context) {
    // 支持分页、排序、过滤
    // GET /api/users?page=1&limit=10&sort=name
    c.JSON(200, gin.H{
        "data": []gin.H{},
        "meta": gin.H{
            "page":        1,
            "limit":       10,
            "total":       100,
            "total_pages": 10,
        },
    })
})

r.POST("/api/users", func(c *gin.Context) {
    c.Header("Location", "/api/users/123")
    c.JSON(201, gin.H{"id": 123, "name": "Alice"})
})

r.PUT("/api/users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"updated": true})
})

r.PATCH("/api/users/:id", func(c *gin.Context) {
    c.JSON(200, gin.H{"patched": true})
})

r.DELETE("/api/users/:id", func(c *gin.Context) {
    c.JSON(204, nil)
})
```

---

## API 版本管理

```go
// URL 路径版本
v1 := r.Group("/api/v1")
{
    v1.GET("/users", getUsersV1)
}

v2 := r.Group("/api/v2")
{
    v2.GET("/users", getUsersV2)
}

// Header 版本
r.GET("/api/users", func(c *gin.Context) {
    version := c.GetHeader("API-Version")
    switch version {
    case "2":
        getUsersV2(c)
    default:
        getUsersV1(c)
    }
})

// Query 参数版本
// GET /api/users?version=2
```

---

## 统一响应格式

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(200, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(201, Response{
        Code:    0,
        Message: "created",
        Data:    data,
    })
}

func Error(c *gin.Context, httpStatus int, code int, message string) {
    c.JSON(httpStatus, Response{
        Code:    code,
        Message: message,
    })
}

func BadRequest(c *gin.Context, message string) {
    Error(c, 400, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
    Error(c, 401, 401, message)
}

func NotFound(c *gin.Context, message string) {
    Error(c, 404, 404, message)
}

// 分页响应
type PaginatedResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
    Meta    Pagination  `json:"meta"`
}

type Pagination struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}
```

---

## 接口文档 (Swagger)

```go
// @Title 创建用户
// @Description 创建新用户
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} User
// @Failure 400 {object} Error
// @Router /users [post]
func createUser(c *gin.Context) {}

// 安装 swag
// go install github.com/swaggo/swag/cmd/swag@latest

// 生成文档
// swag init

// 添加 Swagger UI
import "github.com/swaggo/gin-swagger"
r.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "SWAGGER_DISABLE"))
```
