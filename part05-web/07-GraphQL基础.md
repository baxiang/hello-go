# 5.7 GraphQL 基础

## GraphQL 简介

### GraphQL vs REST

```
REST API:
GET /users/123          # 获取用户
GET /users/123/posts    # 获取用户文章
GET /posts/456/comments # 获取文章评论
(需要多次请求，可能过度获取数据)

GraphQL:
query {
  user(id: 123) {
    name
    email
    posts {
      title
      comments {
        content
        author {
          name
        }
      }
    }
  }
}
(一次请求，精确获取所需数据)
```

### 核心概念

```
Schema:     API 的契约，定义所有可用的数据类型和操作
Query:      查询操作 (类似 REST 的 GET)
Mutation:   修改操作 (类似 REST 的 POST/PUT/DELETE)
Subscription: 订阅 (实时推送)
Resolver:   解析器，实现数据获取逻辑
```

---

## gqlgen 使用

### 安装

```bash
# 初始化 Go 模块
go mod init example.com/graphql

# 安装 gqlgen
go get github.com/99designs/gqlgen
go install github.com/99designs/gqlgen@latest

# 初始化项目
gqlgen init
```

### 项目结构

```
graphql/
├── graph/
│   ├── generated/      # 生成的代码
│   ├── model/          # 数据模型
│   └── resolver.go     # 解析器
├── gqlgen.yml          # 配置文件
├── schema.graphql      # GraphQL Schema
└── server.go           # 服务器入口
```

---

## Schema 定义

### 基础类型

```graphql
# schema.graphql

# 自定义类型
type User {
  id: ID!           # 必填
  name: String!
  email: String!
  age: Int
  role: Role!
  posts: [Post!]    # 数组
  createdAt: Time!
}

type Post {
  id: ID!
  title: String!
  content: String!
  author: User!
  comments: [Comment!]
  published: Boolean!
}

type Comment {
  id: ID!
  content: String!
  author: User!
  post: Post!
}

# 枚举类型
enum Role {
  USER
  ADMIN
  GUEST
}

# 自定义标量
scalar Time

# 查询操作
type Query {
  # 获取单个用户
  user(id: ID!): User

  # 获取所有用户
  users(limit: Int, offset: Int): [User!]!

  # 获取文章
  post(id: ID!): Post

  # 搜索
  searchPosts(keyword: String!, limit: Int): [Post!]!
}

# 修改操作
type Mutation {
  # 创建用户
  createUser(input: CreateUserInput!): User!

  # 更新用户
  updateUser(id: ID!, input: UpdateUserInput!): User!

  # 删除用户
  deleteUser(id: ID!): Boolean!

  # 创建文章
  createPost(input: CreatePostInput!): Post!
}

# 输入类型
input CreateUserInput {
  name: String!
  email: String!
  age: Int
  role: Role
}

input UpdateUserInput {
  name: String
  email: String
  age: Int
  role: Role
}

input CreatePostInput {
  title: String!
  content: String!
  authorId: ID!
}

# 订阅操作
type Subscription {
  # 新用户加入
  userCreated: User!

  # 文章更新
  postUpdated(id: ID!): Post!
}
```

### 指令 (Directives)

```graphql
# 内置指令
type Query {
  # 条件包含
  user(id: ID!): User @deprecated(reason: "Use viewer instead")

  # 跳过
  posts(includeDraft: Boolean!): [Post!]! @skip(if: $includeDraft)

  # 仅当为真时包含
  adminUsers: [User!]! @include(if: $isAdmin)
}
```

---

## Resolver 实现

### 基础 Resolver

```go
package graph

import (
    "context"
    "errors"
    "example.com/graphql/graph/model"
)

// 模拟数据库
var users = []*model.User{
    {ID: "1", Name: "Alice", Email: "alice@example.com", Age: 25},
    {ID: "2", Name: "Bob", Email: "bob@example.com", Age: 30},
}

var posts = []*model.Post{
    {ID: "1", Title: "Hello", Content: "World", AuthorID: "1"},
}

// 实现 Query resolver
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    for _, u := range users {
        if u.ID == id {
            return u, nil
        }
    }
    return nil, errors.New("user not found")
}

func (r *queryResolver) Users(ctx context.Context, limit *int, offset *int) ([]*model.User, error) {
    l, o := 10, 0
    if limit != nil {
        l = *limit
    }
    if offset != nil {
        o = *offset
    }

    if o+l > len(users) {
        return users[o:], nil
    }
    return users[o : o+l], nil
}

func (r *queryResolver) SearchPosts(ctx context.Context, keyword string, limit *int) ([]*model.Post, error) {
    var result []*model.Post
    for _, p := range posts {
        if contains(p.Title, keyword) || contains(p.Content, keyword) {
            result = append(result, p)
        }
    }
    return result, nil
}

// 实现 Mutation resolver
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    user := &model.User{
        ID:    fmt.Sprintf("%d", len(users)+1),
        Name:  input.Name,
        Email: input.Email,
        Age:   input.Age,
    }
    users = append(users, user)
    return user, nil
}

func (r *mutationResolver) UpdateUser(ctx context.Context, id string, input model.UpdateUserInput) (*model.User, error) {
    for _, u := range users {
        if u.ID == id {
            if input.Name != nil {
                u.Name = *input.Name
            }
            if input.Email != nil {
                u.Email = *input.Email
            }
            if input.Age != nil {
                u.Age = *input.Age
            }
            return u, nil
        }
    }
    return nil, errors.New("user not found")
}

func (r *mutationResolver) DeleteUser(ctx context.Context, id string) (bool, error) {
    for i, u := range users {
        if u.ID == id {
            users = append(users[:i], users[i+1:]...)
            return true, nil
        }
    }
    return false, errors.New("user not found")
}

// 字段 Resolver (用于嵌套数据)
func (r *userResolver) Posts(ctx context.Context, obj *model.User) ([]*model.Post, error) {
    var result []*model.Post
    for _, p := range posts {
        if p.AuthorID == obj.ID {
            result = append(result, p)
        }
    }
    return result, nil
}

func (r *postResolver) Author(ctx context.Context, obj *model.Post) (*model.User, error) {
    for _, u := range users {
        if u.ID == obj.AuthorID {
            return u, nil
        }
    }
    return nil, errors.New("author not found")
}

// 辅助函数
func contains(s, substr string) bool {
    return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
```

### 生成代码

```bash
# 生成代码
gqlgen generate

# 运行服务器
go run server.go
```

---

## 服务器实现

### 完整服务器

```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/99designs/gqlgen/graphql/playground"
    "example.com/graphql/graph"
    "example.com/graphql/graph/generated"
)

const defaultPort = "8080"

func main() {
    port := os.Getenv("PORT")
    if port == "" {
        port = defaultPort
    }

    // 创建可执行的 GraphQL schema
    srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
        Resolvers: &graph.Resolver{},
    }))

    // GraphQL 端点
    http.Handle("/query", srv)

    // Playground (GraphQL IDE)
    http.Handle("/", playground.Handler("GraphQL playground", "/query"))

    log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

---

## 中间件与认证

### JWT 认证中间件

```go
package middleware

import (
    "context"
    "net/http"
    "strings"

    "github.com/golang-jwt/jwt/v5"
)

type ContextKey string

const UserContextKey ContextKey = "user"

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

func AuthMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if token == "" {
                http.Error(w, "missing authorization header", http.StatusUnauthorized)
                return
            }

            // 移除 "Bearer " 前缀
            token = strings.TrimPrefix(token, "Bearer ")

            claims := &Claims{}
            parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
                return []byte(secret), nil
            })

            if err != nil || !parsed.Valid {
                http.Error(w, "invalid token", http.StatusUnauthorized)
                return
            }

            // 将用户信息存入 context
            ctx := context.WithValue(r.Context(), UserContextKey, claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// 在 resolver 中获取用户
func (r *Resolver) getUser(ctx context.Context) (*Claims, error) {
    claims, ok := ctx.Value(UserContextKey).(*Claims)
    if !ok {
        return nil, errors.New("user not found in context")
    }
    return claims, nil
}
```

---

## Subscription 实现

### 订阅解析器

```go
package graph

import (
    "context"
    "time"

    "example.com/graphql/graph/model"
)

// 订阅通道
type Subscription struct {
    events chan *model.User
}

var subscriptions = make(map[string]*Subscription)

func (r *subscriptionResolver) UserCreated(ctx context.Context) (<-chan *model.User, error) {
    // 创建订阅通道
    sub := &Subscription{
        events: make(chan *model.User),
    }

    // 生成订阅 ID
    subID := generateSubID()
    subscriptions[subID] = sub

    // 当 context 取消时清理订阅
    go func() {
        <-ctx.Done()
        close(sub.events)
        delete(subscriptions, subID)
    }()

    return sub.events, nil
}

// 发布订阅事件 (在 Mutation 中调用)
func publishUserCreated(user *model.User) {
    for _, sub := range subscriptions {
        select {
        case sub.events <- user:
        default:
            // 跳过慢订阅者
        }
    }
}

func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    user := &model.User{
        ID:    fmt.Sprintf("%d", len(users)+1),
        Name:  input.Name,
        Email: input.Email,
    }
    users = append(users, user)

    // 发布订阅事件
    go publishUserCreated(user)

    return user, nil
}
```

---

## 错误处理

### 自定义错误

```go
import (
    "github.com/vektah/gqlparser/v2/gqlerror"
)

// 返回 GraphQL 错误
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
    // 验证输入
    if input.Email == "" {
        return nil, gqlerror.Errorf("email is required")
    }

    // 检查是否已存在
    for _, u := range users {
        if u.Email == input.Email {
            return nil, gqlerror.Errorf("email already exists")
        }
    }

    // ...
}

// 返回带错误码的错误
func (r *queryResolver) User(ctx context.Context, id string) (*model.User, error) {
    for _, u := range users {
        if u.ID == id {
            return u, nil
        }
    }
    return nil, &gqlerror.Error{
        Message: "User not found",
        Extensions: map[string]interface{}{
            "code": "NOT_FOUND",
        },
    }
}
```

---

## 查询示例

### GraphQL 查询

```graphql
# 查询用户
query {
  user(id: "1") {
    id
    name
    email
  }
}

# 查询用户及其文章
query {
  user(id: "1") {
    name
    posts {
      title
      content
    }
  }
}

# 带变量的查询
query GetUser($id: ID!) {
  user(id: $id) {
    name
    email
  }
}

# 创建用户 (Mutation)
mutation {
  createUser(input: {
    name: "Charlie"
    email: "charlie@example.com"
    age: 28
  }) {
    id
    name
    email
  }
}

# 订阅
subscription {
  userCreated {
    id
    name
    email
  }
}
```

---

## GraphQL 检查清单

```
[ ] 使用 gqlgen 生成类型安全代码
[ ] Schema 设计遵循 RESTful 原则
[ ] 实现适当的错误处理
[ ] 添加认证和授权中间件
[ ] 限制查询深度防止 DoS
[ ] 实现查询复杂度分析
[ ] 使用 DataLoader 解决 N+1 问题
[ ] 实现缓存策略
[ ] 添加速率限制
[ ] 监控查询性能
[ ] 使用 Playground 进行测试
[ ] 版本化 Schema (向后兼容)
```
