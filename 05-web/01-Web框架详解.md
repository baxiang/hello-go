# 5.1 Web 框架

## Gin 基础

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()

    // GET
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    // POST
    r.POST("/users", func(c *gin.Context) {
        name := c.PostForm("name")
        c.JSON(201, gin.H{"name": name})
    })

    r.Run(":8080")
}
```

---

## Gin 路由

```go
// 路径参数
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

// 可选参数
r.GET("/users/:id/:action", func(c *gin.Context) {
    id := c.Param("id")
    action := c.Param("action")
})

// 通配符
r.GET("/assets/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.File("./assets" + filepath)
})

// 路由分组
public := r.Group("/public")
{
    public.GET("/health", healthHandler)
}

auth := r.Group("/api")
auth.Use(AuthMiddleware())
{
    auth.GET("/users", getUsersHandler)
    auth.POST("/users", createUserHandler)
}
```

---

## 参数绑定与验证

```go
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"required,min=1,max=150"`
    Password string `json:"password" binding:"required,min=6"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, gin.H{"user": req})
})

// Query 参数
type QueryRequest struct {
    Page     int `form:"page" binding:"omitempty,min=1"`
    PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

r.GET("/users", func(c *gin.Context) {
    var req QueryRequest
    c.ShouldBindQuery(&req)
})
```

---

## Echo 框架

```go
import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())

e.GET("/", func(c echo.Context) error {
    return c.String(200, "Hello")
})

e.GET("/users/:id", func(c echo.Context) error {
    return c.JSON(200, map[string]string{"id": c.Param("id")})
})

e.Start(":8080")
```

---

## Fiber 框架

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
)

app := fiber.New()
app.Use(cors.New())

app.Get("/", func(c *fiber.Ctx) error {
    return c.SendString("Hello")
})

app.Get("/users/:id", func(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"id": c.Params("id")})
})

app.Listen(":8080")
```
