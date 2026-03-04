# 第五部分：Web 开发

## 5.1 Web 框架

### Gin 框架

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    // ========== 基础示例 ==========
    r := gin.Default()

    // GET 请求
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })

    // POST 请求
    r.POST("/users", func(c *gin.Context) {
        name := c.PostForm("name")
        c.JSON(201, gin.H{"name": name})
    })

    r.Run(":8080")
}
```

### Gin 路由与分组

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // ========== 路径参数 ==========
    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id")
        c.JSON(200, gin.H{"id": id})
    })

    // ========== 可选参数 ==========
    r.GET("/users/:id/:action", func(c *gin.Context) {
        id := c.Param("id")
        action := c.Param("action")
        c.JSON(200, gin.H{"id": id, "action": action})
    })

    // ========== 通配符 ==========
    r.GET("/assets/*filepath", func(c *gin.Context) {
        filepath := c.Param("filepath")
        c.File("./assets" + filepath)
    })

    // ========== 路由分组 ==========
    // 公开路由
    public := r.Group("/public")
    {
        public.GET("/health", healthHandler)
        public.GET("/version", versionHandler)
    }

    // 需要认证的路由
    auth := r.Group("/api")
    auth.Use(AuthMiddleware())
    {
        auth.GET("/users", getUsersHandler)
        auth.POST("/users", createUserHandler)
        auth.GET("/users/:id", getUserHandler)
        auth.PUT("/users/:id", updateUserHandler)
        auth.DELETE("/users/:id", deleteUserHandler)
    }

    // v1 API
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", getUsersHandler)

        // 嵌套分组
        admin := v1.Group("/admin")
        admin.Use(AdminMiddleware())
        {
            admin.GET("/stats", getStatsHandler)
        }
    }

    r.Run(":8080")
}

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "未授权"})
            return
        }
        c.Next()
    }
}

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 检查管理员权限
        c.Next()
    }
}

// Handler 函数
func healthHandler(c *gin.Context)     { c.JSON(200, gin.H{"status": "ok"}) }
func versionHandler(c *gin.Context)    { c.JSON(200, gin.H{"version": "1.0.0"}) }
func getUsersHandler(c *gin.Context)   { c.JSON(200, gin.H{"users": []string{}}) }
func createUserHandler(c *gin.Context) { c.JSON(201, gin.H{"id": 1}) }
func getUserHandler(c *gin.Context)    { c.JSON(200, gin.H{"id": 1}) }
func updateUserHandler(c *gin.Context) { c.JSON(200, gin.H{"updated": true}) }
func deleteUserHandler(c *gin.Context) { c.JSON(200, gin.H{"deleted": true}) }
func getStatsHandler(c *gin.Context)   { c.JSON(200, gin.H{"stats": "admin only"}) }
```

### Gin 参数绑定与验证

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/go-playground/validator/v10"
)

// ========== 结构体定义 ==========
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=2,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"required,min=1,max=150"`
    Password string `json:"password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
    Name  string `json:"name" binding:"omitempty,min=2,max=50"`
    Email string `json:"email" binding:"omitempty,email"`
}

type QueryRequest struct {
    Page     int    `form:"page" binding:"omitempty,min=1"`
    PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
    Keyword  string `form:"keyword"`
}

type PathRequest struct {
    ID uint64 `uri:"id" binding:"required,min=1"`
}

func main() {
    r := gin.Default()

    // ========== JSON 绑定 ==========
    r.POST("/users", func(c *gin.Context) {
        var req CreateUserRequest

        // 绑定并验证 JSON
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        c.JSON(201, gin.H{"user": req})
    })

    // ========== Form 绑定 ==========
    r.POST("/login", func(c *gin.Context) {
        var req struct {
            Username string `form:"username" binding:"required"`
            Password string `form:"password" binding:"required"`
        }

        if err := c.ShouldBind(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"logged_in": true})
    })

    // ========== Query 参数绑定 ==========
    r.GET("/users", func(c *gin.Context) {
        var req QueryRequest

        if err := c.ShouldBindQuery(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // 设置默认值
        if req.Page == 0 {
            req.Page = 1
        }
        if req.PageSize == 0 {
            req.PageSize = 10
        }

        c.JSON(200, req)
    })

    // ========== 路径参数绑定 ==========
    r.GET("/users/:id", func(c *gin.Context) {
        var req PathRequest

        if err := c.ShouldBindUri(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, req)
    })

    // ========== 自定义验证 ==========
    // 注册自定义验证器
    if v, ok := r.Engine.Validator.Engine.(*validator.Validate); ok {
        v.RegisterValidation("name_tag", nameTagValidation)
    }

    r.Run(":8080")
}

// 自定义验证函数
func nameTagValidation(fl validator.FieldLevel) bool {
    name := fl.Field().String()
    // 自定义验证逻辑
    return name != "admin"
}
```

### Gin 错误处理

```go
package main

import (
    "errors"
    "net/http"
    "github.com/gin-gonic/gin"
)

// ========== 自定义错误类型 ==========
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Err     error  `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

var (
    ErrNotFound     = &AppError{Code: 404, Message: "资源未找到"}
    ErrUnauthorized = &AppError{Code: 401, Message: "未授权"}
    ErrBadRequest   = &AppError{Code: 400, Message: "请求参数错误"}
)

func main() {
    r := gin.Default()

    // ========== 统一错误处理 ==========
    r.Use(ErrorHandlerMiddleware())

    r.GET("/users/:id", func(c *gin.Context) {
        id := c.Param("id")

        if id == "" {
            c.Error(ErrBadRequest)
            return
        }

        if id == "999" {
            c.Error(ErrNotFound)
            return
        }

        c.JSON(200, gin.H{"id": id})
    })

    r.Run(":8080")
}

// ========== 错误处理中间件 ==========
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        // 处理错误
        for _, err := range c.Errors {
            switch e := err.Err.(type) {
            case *AppError:
                c.JSON(e.Code, gin.H{
                    "code":    e.Code,
                    "message": e.Message,
                })
            default:
                c.JSON(http.StatusInternalServerError, gin.H{
                    "code":    500,
                    "message": "服务器内部错误",
                })
            }
            return
        }
    }
}
```

### Echo 框架

```go
package main

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
)

func main() {
    e := echo.New()

    // ========== 中间件 ==========
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())

    // ========== 路由 ==========
    e.GET("/", func(c echo.Context) error {
        return c.String(http.StatusOK, "Hello, World!")
    })

    e.GET("/users/:id", func(c echo.Context) error {
        id := c.Param("id")
        return c.JSON(http.StatusOK, map[string]string{"id": id})
    })

    e.POST("/users", func(c echo.Context) error {
        u := new(User)
        if err := c.Bind(u); err != nil {
            return err
        }
        return c.JSON(http.StatusCreated, u)
    })

    // ========== 路由分组 ==========
    api := e.Group("/api")
    api.Use(middleware.JWT([]byte("secret")))
    {
        api.GET("/users", getUsers)
        api.POST("/users", createUser)
    }

    e.Start(":8080")
}

type User struct {
    Name  string `json:"name" validate:"required"`
    Email string `json:"email" validate:"required,email"`
}

func getUsers(c echo.Context) error {
    return c.JSON(http.StatusOK, []User{})
}

func createUser(c echo.Context) error {
    return c.JSON(http.StatusCreated, User{})
}
```

### Fiber 框架

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
    app := fiber.New(fiber.Config{
        // 配置
    })

    // ========== 中间件 ==========
    app.Use(logger.New())
    app.Use(recover.New())
    app.Use(cors.New())

    // ========== 路由 ==========
    app.Get("/", func(c *fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    app.Get("/users/:id", func(c *fiber.Ctx) error {
        id := c.Params("id")
        return c.JSON(fiber.Map{"id": id})
    })

    app.Post("/users", func(c *fiber.Ctx) error {
        type Req struct {
            Name  string `json:"name"`
            Email string `json:"email"`
        }
        var req Req
        if err := c.BodyParser(&req); err != nil {
            return err
        }
        return c.JSON(fiber.Map{"user": req})
    })

    app.Listen(":8080")
}
```

---

## 5.2 中间件开发

### 中间件原理

```go
package main

import (
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

// ========== 中间件本质 ==========
// 中间件是一个函数，接收 HandlerFunc，返回 HandlerFunc
// 可以在请求处理前后执行逻辑

type Middleware func(http.Handler) http.Handler

// Gin 中间件
type GinMiddleware func(*gin.Context)

// ========== 中间件执行流程 ==========
/*
请求 -> Middleware1 -> Middleware2 -> Handler -> Middleware2 -> Middleware1 -> 响应
*/
```

### 自定义中间件

```go
package main

import (
    "fmt"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
)

// ========== 基础中间件 ==========
func SimpleMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 请求前
        fmt.Println("请求前:", c.Request.URL.Path)

        c.Next()  // 执行后续处理

        // 请求后
        fmt.Println("请求后:", c.Writer.Status())
    }
}

// ========== 日志中间件 ==========
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        fmt.Printf("[%s] %s %d %v\n", c.Request.Method, path, status, latency)
    }
}

// ========== 认证中间件 ==========
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        if token == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "未提供 token",
            })
            return
        }

        // 验证 token
        if !validateToken(token) {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "token 无效",
            })
            return
        }

        // 将用户信息存入上下文
        c.Set("userID", "123")
        c.Next()
    }
}

func validateToken(token string) bool {
    return token == "valid-token"
}

// ========== 限流中间件 ==========
func RateLimitMiddleware(maxReqs int, window time.Duration) gin.HandlerFunc {
    requests := make(map[string][]time.Time)

    return func(c *gin.Context) {
        ip := c.ClientIP()
        now := time.Now()

        // 清理过期记录
        windowStart := now.Add(-window)
        var validReqs []time.Time
        for _, t := range requests[ip] {
            if t.After(windowStart) {
                validReqs = append(validReqs, t)
            }
        }

        // 检查是否超限
        if len(validReqs) >= maxReqs {
            c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
                "error": "请求过于频繁",
            })
            return
        }

        // 记录请求
        validReqs = append(validReqs, now)
        requests[ip] = validReqs

        c.Next()
    }
}

func main() {
    r := gin.Default()

    // 全局中间件
    r.Use(LoggerMiddleware())

    // 路由组中间件
    api := r.Group("/api")
    api.Use(AuthMiddleware())
    api.Use(RateLimitMiddleware(10, time.Minute))
    {
        api.GET("/data", func(c *gin.Context) {
            c.JSON(200, gin.H{"data": "sensitive"})
        })
    }

    r.Run(":8080")
}
```

### CORS 中间件

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "*")
        c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
        c.Header("Access-Control-Max-Age", "86400")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}

// ========== 使用 gin-contrib/cors ==========
/*
import "github.com/gin-contrib/cors"

func main() {
    r := gin.Default()

    config := cors.DefaultConfig()
    config.AllowOrigins = []string{"http://localhost:3000"}
    config.AllowCredentials = true

    r.Use(cors.New(config))
    r.Run()
}
*/
```

---

## 5.3 认证与授权

### Session/Cookie 认证

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("secret-key"))

func main() {
    r := gin.Default()

    r.POST("/login", func(c *gin.Context) {
        username := c.PostForm("username")
        password := c.PostForm("password")

        // 验证用户
        if username != "admin" || password != "password" {
            c.JSON(401, gin.H{"error": "用户名或密码错误"})
            return
        }

        // 创建 session
        session, _ := store.Get(c.Request, "session")
        session.Options = &sessions.Options{
            Path:   "/",
            MaxAge: 86400,
        }
        session.Values["username"] = username
        session.Save(c.Request, c.Writer)

        c.JSON(200, gin.H{"message": "登录成功"})
    })

    r.GET("/logout", func(c *gin.Context) {
        session, _ := store.Get(c.Request, "session")
        session.Options.MaxAge = -1
        session.Save(c.Request, c.Writer)
        c.JSON(200, gin.H{"message": "已退出"})
    })

    r.Run(":8080")
}
```

### JWT Token 认证

```go
package main

import (
    "errors"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secret-key")

type Claims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    jwt.RegisteredClaims
}

func main() {
    r := gin.Default()

    // 公开路由
    r.POST("/login", loginHandler)

    // 需要认证的路由
    auth := r.Group("")
    auth.Use(JWTAuthMiddleware())
    {
        auth.GET("/profile", profileHandler)
        auth.POST("/refresh", refreshHandler)
    }

    r.Run(":8080")
}

// ========== 登录生成 Token ==========
func loginHandler(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")

    // 验证用户 (实际应从数据库查询)
    if username != "admin" || password != "password" {
        c.JSON(401, gin.H{"error": "用户名或密码错误"})
        return
    }

    // 生成 Token
    claims := Claims{
        UserID:   1,
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(jwtKey)

    c.JSON(200, gin.H{"token": tokenString})
}

// ========== JWT 认证中间件 ==========
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := c.GetHeader("Authorization")
        if tokenString == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未提供 token"})
            return
        }

        // 去掉 Bearer 前缀
        if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
            tokenString = tokenString[7:]
        }

        // 解析 Token
        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效"})
            return
        }

        // 将用户信息存入上下文
        c.Set("userID", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}

// ========== 刷新 Token ==========
func refreshHandler(c *gin.Context) {
    username, _ := c.Get("username")

    // 生成新 Token
    claims := Claims{
        UserID:   1,
        Username: username.(string),
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString(jwtKey)

    c.JSON(200, gin.H{"token": tokenString})
}

func profileHandler(c *gin.Context) {
    userID, _ := c.Get("userID")
    username, _ := c.Get("username")

    c.JSON(200, gin.H{
        "user_id":  userID,
        "username": username,
    })
}

// ========== 工具函数 ==========
func ParseToken(tokenString string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        return jwtKey, nil
    })

    if err != nil {
        return nil, err
    }

    if !token.Valid {
        return nil, errors.New("token 无效")
    }

    return claims, nil
}
```

### OAuth2.0 集成

```go
package main

import (
    "context"
    "encoding/json"
    "io"
    "net/http"
    "github.com/gin-gonic/gin"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/github"
)

var (
    oauthConfig = &oauth2.Config{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        RedirectURL:  "http://localhost:8080/auth/callback",
        Scopes:       []string{"read:user"},
        Endpoint:     github.Endpoint,
    }
)

func main() {
    r := gin.Default()

    r.GET("/auth/login", func(c *gin.Context) {
        url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
        c.Redirect(http.StatusTemporaryRedirect, url)
    })

    r.GET("/auth/callback", func(c *gin.Context) {
        code := c.Query("code")

        // 换取 Token
        token, err := oauthConfig.Exchange(context.Background(), code)
        if err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        // 获取用户信息
        client := oauthConfig.Client(context.Background(), token)
        resp, _ := client.Get("https://api.github.com/user")
        defer resp.Body.Close()

        body, _ := io.ReadAll(resp.Body)

        var user map[string]interface{}
        json.Unmarshal(body, &user)

        c.JSON(200, gin.H{
            "token": token,
            "user":  user,
        })
    })

    r.Run(":8080")
}
```

### RBAC 权限模型

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// ========== RBAC 模型 ==========
// Role-Based Access Control

type Role string

const (
    RoleAdmin    Role = "admin"
    RoleEditor   Role = "editor"
    RoleViewer   Role = "viewer"
)

// 权限定义
type Permission string

const (
    PermRead    Permission = "read"
    PermWrite   Permission = "write"
    PermDelete  Permission = "delete"
)

// 角色 - 权限映射
var rolePermissions = map[Role][]Permission{
    RoleAdmin:   {PermRead, PermWrite, PermDelete},
    RoleEditor:  {PermRead, PermWrite},
    RoleViewer:  {PermRead},
}

// ========== 权限检查中间件 ==========
func RequirePermission(requiredPerm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从上下文获取用户角色
        roleStr, exists := c.Get("role")
        if !exists {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
            return
        }

        role := Role(roleStr.(string))
        perms := rolePermissions[role]

        // 检查权限
        hasPermission := false
        for _, p := range perms {
            if p == requiredPerm {
                hasPermission = true
                break
            }
        }

        if !hasPermission {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足"})
            return
        }

        c.Next()
    }
}

func main() {
    r := gin.Default()

    // 使用角色中间件 (假设已认证)
    r.Use(func(c *gin.Context) {
        c.Set("role", string(RoleAdmin))
        c.Next()
    })

    // 需要读权限
    r.GET("/users", RequirePermission(PermRead), func(c *gin.Context) {
        c.JSON(200, gin.H{"users": []string{}})
    })

    // 需要写权限
    r.POST("/users", RequirePermission(PermWrite), func(c *gin.Context) {
        c.JSON(201, gin.H{"created": true})
    })

    // 需要删除权限
    r.DELETE("/users/:id", RequirePermission(PermDelete), func(c *gin.Context) {
        c.JSON(200, gin.H{"deleted": true})
    })

    r.Run(":8080")
}
```

### Casbin 使用

```go
package main

import (
    "github.com/casbin/casbin/v2"
    "github.com/gin-gonic/gin"
)

// ========== Casbin 配置 ==========
// model.conf
/*
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
*/

// policy.csv
/*
p, admin, data1, read
p, admin, data1, write
p, editor, data1, read
p, viewer, data1, read

g, alice, admin
g, bob, editor
g, charlie, viewer
*/

func main() {
    // 初始化 Casbin
    enforcer, _ := casbin.NewEnforcer("model.conf", "policy.csv")

    r := gin.Default()

    // Casbin 中间件
    r.Use(func(c *gin.Context) {
        user := c.GetHeader("X-User")  // 从请求头获取用户
        obj := c.Request.URL.Path
        act := c.Request.Method

        allowed, _ := enforcer.Enforce(user, obj, act)
        if !allowed {
            c.AbortWithStatusJSON(403, gin.H{"error": "权限不足"})
            return
        }
        c.Next()
    })

    r.Run(":8080")
}
```

---

## 5.4 API 设计

### RESTful API 设计规范

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // ========== 资源命名 ==========
    // 使用名词复数
    // GET /api/users
    // POST /api/users
    // GET /api/users/123
    // PUT /api/users/123
    // DELETE /api/users/123

    api := r.Group("/api")
    {
        // ========== 获取资源列表 ==========
        api.GET("/users", func(c *gin.Context) {
            // 支持分页、排序、过滤
            // GET /api/users?page=1&limit=10&sort=name&order=desc
            c.JSON(http.StatusOK, gin.H{
                "data": []gin.H{},
                "meta": gin.H{
                    "page":     1,
                    "limit":    10,
                    "total":    100,
                    "total_pages": 10,
                },
            })
        })

        // ========== 创建资源 ==========
        api.POST("/users", func(c *gin.Context) {
            // 请求体
            // { "name": "Alice", "email": "alice@example.com" }

            // 返回 201 Created + Location 头
            c.Header("Location", "/api/users/123")
            c.JSON(http.StatusCreated, gin.H{
                "id": 123,
                "name": "Alice",
            })
        })

        // ========== 获取单个资源 ==========
        api.GET("/users/:id", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{
                "id": 123,
                "name": "Alice",
            })
        })

        // ========== 更新资源 ==========
        // PUT - 完整更新
        api.PUT("/users/:id", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"updated": true})
        })

        // PATCH - 部分更新
        api.PATCH("/users/:id", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"patched": true})
        })

        // ========== 删除资源 ==========
        api.DELETE("/users/:id", func(c *gin.Context) {
            c.JSON(http.StatusNoContent, nil)
        })

        // ========== 嵌套资源 ==========
        // GET /api/users/123/posts
        api.GET("/users/:userId/posts", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"posts": []gin.H{}})
        })
    }

    r.Run(":8080")
}
```

### API 版本管理

```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // ========== URL 路径版本 ==========
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", getUsersV1)
    }

    v2 := r.Group("/api/v2")
    {
        v2.GET("/users", getUsersV2)
    }

    // ========== Header 版本 ==========
    r.GET("/api/users", func(c *gin.Context) {
        version := c.GetHeader("API-Version")
        switch version {
        case "2":
            getUsersV2(c)
        default:
            getUsersV1(c)
        }
    })

    // ========== Query 参数版本 ==========
    // GET /api/users?version=2

    r.Run(":8080")
}

func getUsersV1(c *gin.Context) {
    c.JSON(200, gin.H{"version": "v1", "users": []string{}})
}

func getUsersV2(c *gin.Context) {
    c.JSON(200, gin.H{"version": "v2", "users": []gin.H{}})
}
```

### 接口文档 (Swagger/OpenAPI)

```go
// ========== 使用 swag 生成文档 ==========
/*
1. 安装
go install github.com/swaggo/swag/cmd/swag@latest

2. 添加注释
// @Title 创建用户
// @Description 创建新用户
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "用户信息"
// @Success 201 {object} User
// @Failure 400 {object} Error
// @Router /users [post]
func createUser(c *gin.Context) {}

3. 生成文档
swag init

4. 添加 Swagger UI
import "github.com/swaggo/gin-swagger"
r.GET("/swagger/*any", ginSwagger.DisablingWrapHandler(swaggerFiles.Handler, "SWAGGER_DISABLE"))
*/
```

### 统一响应格式

```go
package main

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

// ========== 统一响应结构 ==========
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// ========== 成功响应 ==========
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

func Created(c *gin.Context, data interface{}) {
    c.JSON(http.StatusCreated, Response{
        Code:    0,
        Message: "created",
        Data:    data,
    })
}

// ========== 错误响应 ==========
func Error(c *gin.Context, httpStatus int, code int, message string) {
    c.JSON(httpStatus, Response{
        Code:    code,
        Message: message,
    })
}

func BadRequest(c *gin.Context, message string) {
    Error(c, http.StatusBadRequest, 400, message)
}

func Unauthorized(c *gin.Context, message string) {
    Error(c, http.StatusUnauthorized, 401, message)
}

func NotFound(c *gin.Context, message string) {
    Error(c, http.StatusNotFound, 404, message)
}

func InternalError(c *gin.Context, message string) {
    Error(c, http.StatusInternalServerError, 500, message)
}

// ========== 分页响应 ==========
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

func Paginated(c *gin.Context, data interface{}, page, pageSize, total int) {
    totalPages := (total + pageSize - 1) / pageSize
    c.JSON(http.StatusOK, PaginatedResponse{
        Code:    0,
        Message: "success",
        Data:    data,
        Meta: Pagination{
            Page:       page,
            PageSize:   pageSize,
            Total:      total,
            TotalPages: totalPages,
        },
    })
}

// ========== 使用示例 ==========
func main() {
    r := gin.Default()

    r.GET("/users/:id", func(c *gin.Context) {
        // 模拟查询
        user := gin.H{"id": 1, "name": "Alice"}
        Success(c, user)
    })

    r.Run(":8080")
}
```

---

## 5.5 微服务基础

### gRPC 基础

```go
// ========== 定义 proto 文件 ==========
// proto/user.proto
/*
syntax = "proto3";

package user;

option go_package = "example.com/user";

service UserService {
    rpc GetUser(GetUserRequest) returns (User);
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
    rpc CreateUser(CreateUserRequest) returns (User);
}

message User {
    int32 id = 1;
    string name = 2;
    string email = 3;
}

message GetUserRequest {
    int32 id = 1;
}

message ListUsersRequest {
    int32 page = 1;
    int32 page_size = 2;
}

message ListUsersResponse {
    repeated User users = 1;
    int32 total = 2;
}

message CreateUserRequest {
    string name = 1;
    string email = 2;
}
*/

// ========== 生成 Go 代码 ==========
// go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
// go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
// protoc --go_out=. --go_opt=paths=source_relative \
//        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
//        user.proto

// ========== gRPC 服务器 ==========
/*
package main

import (
    "context"
    "net"
    "google.golang.org/grpc"
    pb "example.com/user"
)

type UserService struct {
    pb.UnimplementedUserService
}

func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
    return &pb.User{
        Id:    req.Id,
        Name:  "Alice",
        Email: "alice@example.com",
    }, nil
}

func main() {
    lis, _ := net.Listen("tcp", ":50051")
    server := grpc.NewServer()
    pb.RegisterUserService(server, &UserService{})
    server.Serve(lis)
}
*/

// ========== gRPC 客户端 ==========
/*
conn, _ := grpc.Dial("localhost:50051", grpc.WithInsecure())
client := pb.NewUserServiceClient(conn)

resp, _ := client.GetUser(context.Background(), &pb.GetUserRequest{Id: 1})
fmt.Println(resp.Name)
*/
```

### Protocol Buffers

```go
// ========== Proto 语法 ==========
/*
// 标量类型
double, float, int32, int64, uint32, uint64
sint32, sint64, fixed32, fixed64, sfixed32, sfixed64
bool, string, bytes

// 字段规则
optional  // 可有可无
repeated  // 数组
oneof     // 多选一

// 默认值
string name = 1 [default = "unknown"];

// 枚举
enum Status {
    STATUS_UNKNOWN = 0;
    STATUS_ACTIVE = 1;
    STATUS_INACTIVE = 2;
}

// Map
map<string, int32> scores = 1;

// 嵌套消息
message Address {
    string street = 1;
    string city = 2;
}

message Person {
    string name = 1;
    Address address = 2;
}
*/
```

### 服务发现

```go
// ========== etcd 服务发现 ==========
/*
import (
    "github.com/etcd-io/etcd/client/v3"
)

// 服务注册
func RegisterService(etcdClient *clientv3.Client, serviceName, serviceAddr string) {
    key := fmt.Sprintf("/services/%s/%s", serviceName, serviceAddr)
    etcdClient.Put(context.Background(), key, serviceAddr)
}

// 服务发现
func DiscoverService(etcdClient *clientv3.Client, serviceName string) ([]string, error) {
    resp, err := etcdClient.Get(context.Background(),
        fmt.Sprintf("/services/%s/", serviceName),
        clientv3.WithPrefix())

    var addrs []string
    for _, kv := range resp.Kvs {
        addrs = append(addrs, string(kv.Value))
    }
    return addrs, err
}
*/
```

### 负载均衡

```go
// ========== gRPC 负载均衡 ==========
/*
// 使用 gRPC 内置的负载均衡
conn, _ := grpc.Dial(
    "etcd:///myservice",  // etcd 作为 resolver
    grpc.WithInsecure(),
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
*/

// ========== 客户端负载均衡 ==========
/*
// 1. 获取服务列表
// 2. 选择策略：随机、轮询、加权
// 3. 健康检查
*/
```

---

## 第五部分完

接下来可以继续学习第六部分：工程实践
