# AGENTS.md

This file provides guidance to agentic coding agents operating in this Go language learning documentation repository.

## Repository Overview

This is a **Go language learning documentation repository** containing comprehensive Chinese technical guides covering Go basics through cloud-native development. Contains markdown documentation (10 parts, 50+ files) and sample Go code in `part08-projects/kratos/`.

## Build/Test/Lint Commands

### Go Commands
```bash
# Run all tests in a module
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a single test file
go test -v -run TestFunctionName ./path/to/package
go test -v -run TestUserService ./internal/service

# Run a specific test case (subtest)
go test -v -run "TestUserService/成功获取用户" ./internal/service

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmark tests
go test -bench=. ./...
go test -bench=. -benchmem ./...

# Run fuzz tests
go test -fuzz=FuzzParseInt ./...

# Format code
go fmt ./...
goimports -w .

# Lint code
go vet ./...
golangci-lint run
golangci-lint run ./...
golangci-lint run --fix

# Build
go build ./...
go build -o app ./cmd/myapp

# Tidy dependencies
go mod tidy
go mod verify
```

### Sample Project (part08-projects/kratos)
```bash
cd part08-projects/kratos

# Run tests for kratos project
go test ./...

# Run specific package tests
go test -v ./internal/service
go test -v ./internal/biz

# Build services
go build -o user-service ./cmd/user-service
go build -o order-service ./cmd/order-service
```

## Code Style Guidelines

### Imports Ordering
Imports are organized in three groups, separated by blank lines:
1. Standard library imports
2. Third-party imports
3. Local project imports

```go
import (
    "context"
    "errors"
    "time"

    "github.com/go-kratos/kratos/v2/log"
    "gorm.io/gorm"

    "kratos/internal/data"
    "kratos/internal/repo"
)
```

### Naming Conventions
- **Package names**: Short, lowercase, single word (e.g., `handler`, `service`, `biz`)
- **Variables**: CamelCase (e.g., `userID`, `userService`)
- **Constants**: CamelCase or PascalCase (e.g., `MaxRetry`, `defaultTimeout`)
- **Functions/Methods**: PascalCase for public, camelCase for private
- **Interfaces**: PascalCase with verb suffix (e.g., `Reader`, `UserRepository`)
- **Structs**: PascalCase (e.g., `User`, `UserUseCase`)
- **Acronyms**: Keep consistent (e.g., `HTTPServer`, `HTTPClient`, not `HttpServer`)

### Comments
- Use Chinese comments in code examples (this is a Chinese learning resource)
- Package comments start with `// Package <name> ...`
- Function comments start with function name: `// GetUser 根据 ID 获取用户`
- Explain "why" not "what" for complex logic

```go
// Package handler 提供 HTTP 处理器
package handler

// GetUser 根据 ID 获取用户
func GetUser(id int) (*User, error) {
    // ...
}

// initLogger 初始化日志
func initLogger() *zap.Logger {
    // ...
}
```

### Error Handling
- Always check errors explicitly
- Error messages should be lowercase, no trailing period
- Use `errors.New()` for simple errors
- Define sentinel errors at package level
- Wrap errors with context using `fmt.Errorf()` or errors wrapping

```go
var (
    ErrUserNotFound      = errors.New("用户不存在")
    ErrUserAlreadyExists = errors.New("用户已存在")
    ErrInvalidPassword   = errors.New("密码错误")
)

if err != nil {
    return fmt.Errorf("failed to get user: %w", err)
}
```

### Struct Definition
- Use struct tags for JSON, mapstructure, etc.
- Group related fields together
- Use sensible zero values

```go
type User struct {
    ID        int64     `json:"id" mapstructure:"id"`
    Username  string    `json:"username" mapstructure:"username"`
    Email     string    `json:"email" mapstructure:"email"`
    Password  string    `json:"-"`
    Status    int32     `json:"status" mapstructure:"status"`
    CreatedAt time.Time `json:"created_at" mapstructure:"created_at"`
    UpdatedAt time.Time `json:"updated_at" mapstructure:"updated_at"`
}
```

### Function Design
- Accept interfaces, return structs
- Keep functions small and focused
- Use guard clauses for early returns
- Avoid deep nesting

```go
func (uc *UserUseCase) Get(ctx context.Context, id int64) (*User, error) {
    user, err := uc.userRepo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, err
    }
    return user, nil
}
```

### Testing Patterns
- Use table-driven tests
- Clear test naming
- Mock external dependencies with testify/mock or gomock
- Use testify assertions

```go
func TestUserService_GetUser(t *testing.T) {
    tests := []struct {
        name        string
        userID      string
        mockSetup   func(*MockUserRepository)
        expectUser  *User
        expectError error
    }{
        {
            name:   "成功获取用户",
            userID: "123",
            mockSetup: func(m *MockUserRepository) {
                m.On("GetByID", "123").Return(&User{ID: "123", Name: "Alice"}, nil)
            },
            expectUser:  &User{ID: "123", Name: "Alice"},
            expectError: nil,
        },
        // ...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(MockUserRepository)
            tt.mockSetup(mockRepo)

            service := &UserService{repo: mockRepo}
            user, err := service.GetUser(tt.userID)

            // assertions...
            mockRepo.AssertExpectations(t)
        })
    }
}
```

### Project Structure (Standard Go Layout)
```
project/
├── cmd/                    # Application entrypoints
│   └── myapp/main.go
├── internal/               # Private packages (not importable externally)
│   ├── config/
│   ├── service/
│   ├── biz/               # Business logic
│   ├── repo/              # Repository interface
│   └── data/              # Data layer implementation
├── pkg/                    # Public packages (importable)
├── api/                    # API definitions (proto, swagger)
├── configs/                # Configuration files
├── go.mod
├── go.sum
└── Makefile
```

### Dependencies
- **kratos**: Microservice framework (github.com/go-kratos/kratos/v2)
- **viper**: Configuration management (github.com/spf13/viper)
- **zap**: Structured logging (go.uber.org/zap)
- **gorm**: ORM (gorm.io/gorm)
- **testify**: Testing utilities (github.com/stretchr/testify)
- **wire**: Dependency injection (github.com/google/wire)

### Code Quality Tools
- **golangci-lint**: Integrated linting tool
- **go vet**: Static analysis
- **staticcheck**: Advanced static analysis
- **goimports**: Import management

### Common Patterns Used
- **Clean Architecture**: service -> biz -> repo -> data layers
- **Dependency Injection**: Constructor functions with explicit dependencies
- **Functional Options**: Configuration pattern for complex constructors
- **Error Wrapping**: Preserving error context with `%w`
- **Context Propagation**: Always pass `context.Context` as first parameter

### Formatting Rules
- Use `gofmt` and `goimports` for formatting
- No trailing whitespace
- Tabs for indentation (Go standard)
- Align struct fields when appropriate

### Documentation Files
- Markdown files use Chinese content
- File naming: `NN-标题.md` format (e.g., `01-环境搭建与工具链.md`)
- Code examples are complete, runnable Go code with package and imports
- Include checklists and comparison tables where applicable

## Important Notes
- This repository primarily contains documentation in markdown format
- Sample Go code is located in `part08-projects/kratos/`
- Comments in Go code examples are in Chinese
- Follow Effective Go and Uber Go Style Guide principles
- Ensure all code examples are syntactically correct and runnable