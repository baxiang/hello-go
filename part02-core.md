# 第二部分：核心特性

## 2.1 结构体与方法

### 结构体定义与实例化

```go
package main

import "fmt"

// ========== 结构体定义 ==========
type Person struct {
    Name string
    Age  int
    City string
}

// ========== 匿名字段 ==========
type Employee struct {
    Person  // 匿名嵌入
    Company string
    Salary  float64
}

// ========== 结构体实例化 ==========
func main() {
    // 方式 1: 字面量
    p1 := Person{"Alice", 25, "Beijing"}
    p2 := Person{Name: "Bob", Age: 30}  // City 为空值

    // 方式 2: new
    p3 := new(Person)  // *Person
    p3.Name = "Carol"
    p3.Age = 28

    // 方式 3: 取地址
    p4 := &Person{Name: "David", Age: 35}

    fmt.Println(p1)
    fmt.Println(p2)
    fmt.Println(*p3)
    fmt.Println(*p4)

    // ========== 字段访问 ==========
    // 指针可以自动解引用
    fmt.Println(p4.Name)  // 等价于 (*p4).Name
    p4.Age = 36

    // ========== 匿名嵌入的使用 ==========
    emp := Employee{
        Person:  Person{Name: "Eve", Age: 28, City: "Shanghai"},
        Company: "ACME",
        Salary:  50000,
    }

    // 可以直接访问嵌入类型的字段
    fmt.Println(emp.Name)     // Eve
    fmt.Println(emp.Company)  // ACME

    // ========== 结构体比较 ==========
    // 可比较的结构体可以用 == 比较
    p5 := Person{"Alice", 25, "Beijing"}
    fmt.Println(p1 == p5)  // true

    // 包含切片、Map、函数的结构体不可比较
}
```

### 匿名字段与结构体嵌入

```go
package main

import "fmt"

// ========== 嵌入多个类型 ==========
type Address struct {
    Street string
    City   string
    Zip    string
}

type Contact struct {
    Email string
    Phone string
}

type User struct {
    Name    string
    Address  // 匿名嵌入
    Contact  // 匿名嵌入
}

// ========== 提升规则 ==========
// 嵌入类型的字段会提升到外层结构体

// ========== 字段冲突 ==========
type A struct {
    X int
}

type B struct {
    X int
}

type C struct {
    A
    B
}

func main() {
    user := User{
        Name: "Alice",
        Address: Address{
            Street: "Main St",
            City:   "Beijing",
        },
        Contact: Contact{
            Email: "alice@example.com",
        },
    }

    // 可以直接访问提升的字段
    fmt.Println(user.City)   // Beijing
    fmt.Println(user.Email)  // alice@example.com

    // ========== 字段冲突处理 ==========
    c := C{}
    // c.X = 1  // 编译错误：ambiguous selector

    // 必须明确指定
    c.A.X = 1
    c.B.X = 2

    fmt.Println(c.A.X, c.B.X)  // 1 2

    // ========== 嵌入指针类型 ==========
    type Config struct {
        Debug bool
    }

    type Server struct {
        *Config  // 嵌入指针
        Host     string
    }

    s := Server{
        Config: &Config{Debug: true},
        Host:   "localhost",
    }

    fmt.Println(s.Debug)  // true

    // 注意：指针可能为 nil
    s2 := Server{Host: "example.com"}
    // fmt.Println(s2.Debug)  // panic: nil pointer dereference
    if s2.Config != nil {
        fmt.Println(s2.Debug)
    }

    // ========== 嵌入接口 ==========
    type Reader interface {
        Read(p []byte) (int, error)
    }

    type Writer interface {
        Write(p []byte) (int, error)
    }

    type ReadWriter struct {
        Reader  // 嵌入接口
        Writer  // 嵌入接口
    }
    // ReadWriter 实现了 Reader 和 Writer 的所有方法
}
```

### 结构体标签 (struct tags)

```go
package main

import (
    "encoding/json"
    "fmt"
    "reflect"
)

// ========== 结构体标签定义 ==========
type User struct {
    ID       int     `json:"id" db:"user_id"`
    Name     string  `json:"name" db:"user_name" validate:"required"`
    Email    string  `json:"email,omitempty" db:"email"`
    Password string  `json:"-" db:"password"`  // - 表示忽略
    Age      int     `json:"age" db:"age" default:"0"`
}

func main() {
    // ========== 标签的用途 ==========
    // 1. JSON 序列化/反序列化
    user := User{
        ID:       1,
        Name:     "Alice",
        Email:    "alice@example.com",
        Password: "secret",
    }

    data, _ := json.Marshal(user)
    fmt.Println(string(data))
    // {"id":1,"name":"Alice","email":"alice@example.com","age":0}

    // 2. 反射读取标签
    t := reflect.TypeOf(user)
    field, _ := t.FieldByName("Name")

    jsonTag := field.Tag.Get("json")
    dbTag := field.Tag.Get("db")
    validateTag := field.Tag.Get("validate")

    fmt.Println("json:", jsonTag)           // name
    fmt.Println("db:", dbTag)               // user_name
    fmt.Println("validate:", validateTag)   // required

    // ========== 解析标签 ==========
    // 自定义解析
    tags := field.Tag
    for _, tag := range []string{"json", "db", "validate"} {
        fmt.Printf("%s: %s\n", tag, tags.Get(tag))
    }

    // ========== 标签语法解析 ==========
    // `key:"value" key2:"value2"`
    // 使用空格分隔不同的 key
    // 值必须用双引号包围

    // ========== 常见用途 ==========
    // 1. JSON: `json:"name,omitempty"`
    // 2. 数据库：`db:"column_name"`
    // 3. 表单绑定：`form:"username"`
    // 4. 验证：`validate:"required,min=3,max=20"`
    // 5. 环境变量：`env:"APP_PORT" envDefault:"8080"`
}
```

### 方法定义

```go
package main

import "fmt"

// ========== 方法定义 ==========
type Counter struct {
    value int
}

// 值接收者方法
func (c Counter) Value() int {
    return c.value
}

// 指针接收者方法
func (c *Counter) Increment() {
    c.value++
}

func (c *Counter) Add(n int) {
    c.value += n
}

// ========== 接收者类型 ==========
type Number int

func (n Number) Double() Number {
    return n * 2
}

func (n *Number) Set(v int) {
    *n = Number(v)
}

func main() {
    // ========== 方法调用 ==========
    c := Counter{value: 0}

    fmt.Println(c.Value())  // 0

    c.Increment()
    fmt.Println(c.Value())  // 1

    c.Add(10)
    fmt.Println(c.Value())  // 11

    // ========== 值接收者 vs 指针接收者 ==========

    // 值接收者
    // - 不能修改接收者
    // - 值和指针都可以调用

    // 指针接收者
    // - 可以修改接收者
    // - 值和指针都可以调用 (Go 自动取地址)

    c2 := Counter{value: 5}
    pc := &c2
    c2.Increment()  // Go 自动转换为 (&c2).Increment()
    pc.Increment()  // 直接调用

    // ========== 接收者选择指南 ==========
    // 使用指针接收者:
    // 1. 需要修改接收者
    // 2. 接收者是大结构体 (避免复制)
    // 3. 为了一致性

    // 使用值接收者:
    // 1. 不需要修改接收者
    // 2. 接收者是小结构体或基本类型
    // 3. 类型是不可变的

    // ========== 方法作为值 ==========
    n := Number(5)
    doubleFn := n.Double  // 方法值
    fmt.Println(doubleFn())  // 10

    // ========== 不能定义方法的情况 ==========
    // 1. 非本地类型
    // func (i int) Double() int { return i * 2 }  // 错误！

    // 2. 指向非本地类型的指针
    // func (s *string) Upper() string { ... }  // 错误！

    // 解决方案：定义新类型
    type MyString string
    func (s MyString) Upper() string {
        return string(s) + " [UPPER]"
    }
}
```

### Stringer 接口实现

```go
package main

import (
    "fmt"
    "strings"
)

// ========== Stringer 接口 ==========
// type Stringer interface {
//     String() string
// }

type Person struct {
    Name string
    Age  int
}

// 实现 String 方法
func (p Person) String() string {
    return fmt.Sprintf("%s(%d)", p.Name, p.Age)
}

// ========== 切片/指针的 String 方法 ==========
type Items []int

func (i Items) String() string {
    var sb strings.Builder
    sb.WriteString("[")
    for idx, v := range i {
        if idx > 0 {
            sb.WriteString(", ")
        }
        sb.WriteString(fmt.Sprintf("%d", v))
    }
    sb.WriteString("]")
    return sb.String()
}

func main() {
    p := Person{Name: "Alice", Age: 25}
    fmt.Println(p)  // Alice(25) - 自动调用 String()

    items := Items{1, 2, 3, 4, 5}
    fmt.Println(items)  // [1, 2, 3, 4, 5]

    // ========== fmt 包中的其他相关接口 ==========
    // - GoStringer: GoString() string (用于 %#v)
    // - Formatter: Format(s State, v rune)
}

// ========== GoStringer 示例 ==========
type Secret struct {
    Value string
}

func (s Secret) String() string {
    return "[REDACTED]"
}

func (s Secret) GoString() string {
    return fmt.Sprintf("Secret{Value: %q}", s.Value)
}

// 使用:
// s := Secret{Value: "password"}
// fmt.Println(s)    // [REDACTED]  (%v)
// fmt.Printf("%#v\n", s)  // Secret{Value: "password"}
```

---

## 2.2 接口

### 接口定义与实现

```go
package main

import "fmt"

// ========== 接口定义 ==========
type Speaker interface {
    Speak() string
}

// ========== 隐式实现 ==========
// Go 的接口是隐式实现的，不需要显式声明

type Dog struct {
    Name string
}

func (d Dog) Speak() string {
    return "Woof!"
}

type Cat struct {
    Name string
}

func (c Cat) Speak() string {
    return "Meow!"
}

// ========== 接口作为参数 ==========
func MakeSpeak(s Speaker) {
    fmt.Println(s.Speak())
}

// ========== 接口作为返回值 ==========
func NewPet(name string) Speaker {
    return Dog{Name: name}
}

func main() {
    dog := Dog{Name: "Buddy"}
    cat := Cat{Name: "Kitty"}

    // 接口变量
    var s Speaker
    s = dog
    fmt.Println(s.Speak())  // Woof!

    s = cat
    fmt.Println(s.Speak())  // Meow!

    // 接口切片
    pets := []Speaker{dog, cat}
    for _, pet := range pets {
        pet.Speak()
    }

    // ========== 空接口 ==========
    // interface{} 可以存储任何类型的值
    var anything interface{}
    anything = 42
    anything = "hello"
    anything = Dog{Name: "Max"}

    // ========== 接口Nil检查 ==========
    var nilSpeaker Speaker
    fmt.Println(nilSpeaker == nil)  // true

    // 注意：接口包含类型和值
    // 只有类型和值都为 nil 时，接口才是 nil
}
```

### 空接口 interface{}

```go
package main

import "fmt"

func main() {
    // ========== 空接口存储任意类型 ==========
    var x interface{}

    x = 42
    x = "hello"
    x = true
    x = map[string]int{"a": 1}
    x = []int{1, 2, 3}

    // ========== 空接口切片 ==========
    items := []interface{}{1, "hello", true, 3.14}

    for i, item := range items {
        fmt.Printf("%d: %v (type: %T)\n", i, item, item)
    }

    // ========== 通用函数 ==========
    // 接受任意类型的参数
    PrintAnything(42)
    PrintAnything("hello")
    PrintAnything([]int{1, 2, 3})

    // ========== 泛型出现前的常用模式 ==========
    // 现在推荐使用泛型
}

func PrintAnything(x interface{}) {
    fmt.Printf("Value: %v, Type: %T\n", x, x)
}

// ========== 注意 ==========
// 空接口会丢失类型信息
// 需要类型断言或类型开关来获取实际类型
```

### 类型断言与类型开关

```go
package main

import "fmt"

func main() {
    // ========== 类型断言 ==========
    var x interface{} = "hello"

    // 语法: value, ok := x.(T)
    s, ok := x.(string)
    fmt.Println(s, ok)  // hello true

    // 断言失败，ok 为 false
    i, ok := x.(int)
    fmt.Println(i, ok)  // 0 false

    // 不安全断言 (可能 panic)
    // s2 := x.(int)  // panic!

    // ========== 类型开关 ==========
    var values []interface{} = []interface{}{
        "hello", 42, 3.14, true, []int{1, 2, 3},
    }

    for _, v := range values {
        switch val := v.(type) {
        case string:
            fmt.Printf("字符串: %s\n", val)
        case int:
            fmt.Printf("整数: %d\n", val)
        case float64:
            fmt.Printf("浮点数: %.2f\n", val)
        case bool:
            fmt.Printf("布尔: %t\n", val)
        case []int:
            fmt.Printf("整数切片: %v\n", val)
        default:
            fmt.Printf("未知类型：%T - %v\n", val, val)
        }
    }

    // ========== 实际应用场景 ==========
    // 1. 解析 JSON (map[string]interface{})
    // 2. 处理接口返回值
    // 3. 实现通用容器 (泛型出现前)
}
```

### 接口嵌套

```go
package main

import "fmt"

// ========== 接口嵌套定义 ==========
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 接口可以嵌套
type ReadWriter interface {
    Reader
    Writer
}

type Closer interface {
    Close() error
}

// 多个接口组合
type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// ========== 标准库中的例子 ==========
// io.ReadWriter = Reader + Writer
// io.ReadWriteCloser = Reader + Writer + Closer

func main() {
    // ========== 接口实现 ==========
    // 实现 ReadWriteCloser 需要实现所有方法
}

// ========== 接口设计原则 ==========
// 1. 接口要小 (单一职责)
// 2. 接受接口，返回结构体
// 3. 不要为了接口而接口
// 4. 接口应该描述行为，而不是数据
```

### 接口设计原则

```go
package main

import (
    "fmt"
    "io"
    "os"
)

// ========== 原则 1: 小接口 ==========
// 好接口：方法少，职责单一
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// 坏接口：大而全
type BadStorage interface {
    Read() string
    Write(data string) error
    Delete() error
    Update(data string) error
    List() []string
    // ... 太多方法
}

// ========== 原则 2: 接受接口，返回结构体 ==========
func ProcessReader(r io.Reader) error {
    // 处理读取逻辑
    return nil
}

func NewFileReader(filename string) (*os.File, error) {
    return os.Open(filename)
}

// ========== 原则 3: 接口应该在消费者处定义 ==========
// 调用方定义需要的接口，而不是让实现方定义大接口

type Stringer interface {
    String() string
}

func PrintStrings(items []Stringer) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// ========== 原则 4: 接口组合 ==========
// 将小接口组合成需要的功能
type Config interface {
    io.Reader
    io.Writer
    io.Closer
}

// ========== 原则 5: 避免过早抽象 ==========
// 先有实现，再有接口
// 当有多个实现需要时，再提取接口
```

---

## 2.3 错误处理

### error 接口

```go
package main

import (
    "errors"
    "fmt"
)

// error 接口定义
// type error interface {
//     Error() string
// }

// ========== 创建错误 ==========
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("除数不能为零")
    }
    return a / b, nil
}

// ========== 错误检查 ==========
func main() {
    result, err := divide(10, 0)
    if err != nil {
        fmt.Println("错误:", err)
        return
    }
    fmt.Println(result)

    // ========== 错误处理模式 ==========

    // 1. 直接返回
    if err != nil {
        return err
    }

    // 2. 包装错误
    if err != nil {
        return fmt.Errorf("处理数据失败：%w", err)
    }

    // 3. 记录日志后继续
    if err != nil {
        logError(err)
        // 继续执行
    }

    // 4. 忽略错误 (明确标注)
    result, _ := divide(10, 2)  // 明确表示忽略错误
    fmt.Println(result)
}

func logError(err error) {
    fmt.Println("记录错误:", err)
}
```

### 自定义错误类型

```go
package main

import (
    "errors"
    "fmt"
    "time"
)

// ========== 自定义错误类型 ==========
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("验证错误 [%s]: %s", e.Field, e.Message)
}

// ========== 带堆栈的错误 ==========
type AppError struct {
    Code    string
    Message string
    Err     error
    Time    time.Time
}

func (e *AppError) Error() string {
    return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// ========== 错误类型判断 ==========
func validateAge(age int) error {
    if age < 0 {
        return &ValidationError{
            Field:   "age",
            Message: "年龄不能为负数",
        }
    }
    if age > 150 {
        return &ValidationError{
            Field:   "age",
            Message: "年龄不能超过 150",
        }
    }
    return nil
}

func main() {
    err := validateAge(-1)
    if err != nil {
        // 类型断言
        if ve, ok := err.(*ValidationError); ok {
            fmt.Printf("字段：%s, 消息：%s\n", ve.Field, ve.Message)
        }
    }

    // ========== 哨兵错误 ==========
    var ErrNotFound = errors.New("未找到")
    var ErrUnauthorized = errors.New("未授权")

    // ========== 错误包装 ==========
    dbErr := errors.New("数据库连接失败")
    appErr := &AppError{
        Code:    "DB_ERROR",
        Message: "无法连接数据库",
        Err:     dbErr,
        Time:    time.Now(),
    }

    fmt.Println(appErr)
}
```

### errors 包 (Is, As, Join)

```go
package main

import (
    "errors"
    "fmt"
)

// ========== 哨兵错误 ==========
var ErrNotFound = errors.New("记录未找到")
var ErrInvalidInput = errors.New("无效输入")

func findUser(id int) error {
    if id < 0 {
        return ErrInvalidInput
    }
    return ErrNotFound
}

func main() {
    // ========== errors.Is ==========
    // 检查错误链中是否包含某个错误

    err := findUser(-1)
    if errors.Is(err, ErrInvalidInput) {
        fmt.Println("输入无效")
    }

    // ========== errors.As ==========
    // 从错误链中提取特定类型的错误

    type ValidationError struct {
        Field string
    }

    func (e *ValidationError) Error() string {
        return "验证错误"
    }

    var ve *ValidationError
    if errors.As(err, &ve) {
        fmt.Println("字段:", ve.Field)
    }

    // ========== errors.Join ==========
    // 合并多个错误 (Go 1.20+)

    err1 := errors.New("错误 1")
    err2 := errors.New("错误 2")
    err3 := errors.Join(err1, err2)

    fmt.Println(err3)  // 错误 1\n错误 2

    // 检查合并的错误
    if errors.Is(err3, err1) {
        fmt.Println("包含错误 1")
    }

    // ========== 错误包装 ==========
    // %w 包装错误
    baseErr := errors.New("原始错误")
    wrappedErr := fmt.Errorf("包装：%w", baseErr)

    fmt.Println(wrappedErr)  // 包装：原始错误
    fmt.Println(errors.Unwrap(wrappedErr))  // 原始错误

    // 多层包装
    err4 := fmt.Errorf("第二层：%w", wrappedErr)
    fmt.Println(errors.Is(err4, baseErr))  // true
}
```

### panic 与 recover

```go
package main

import "fmt"

func main() {
    // ========== panic 基础 ==========
    // panic 会停止当前函数执行，开始栈展开

    // ========== recover 捕获 panic ==========
    result := safeDivide(10, 0)
    fmt.Println("结果:", result)

    // ========== defer + recover ==========
    safeFunc()

    // ========== panic 使用场景 ==========
    // 1. 不可恢复的错误
    // 2. 程序启动时配置错误
    // 3. 不应该发生的情况 (防御性编程)
}

func safeDivide(a, b int) (result int) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("捕获 panic:", r)
            result = 0  // 设置默认值
        }
    }()

    if b == 0 {
        panic("除数为零")
    }

    return a / b
}

func safeFunc() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("从 panic 恢复:", r)
        }
    }()

    panic("测试 panic")
}

// ========== panic/recover 最佳实践 ==========
// 1. 不要在普通错误处理中使用
// 2. 只在 goroutine 顶层使用 recover
// 3. 用于保护服务不崩溃
// 4. 中间件中统一处理 panic
```

### 错误处理最佳实践

```go
package main

import (
    "errors"
    "fmt"
)

// ========== 1. 总是检查错误 ==========
func badExample() {
    // 错误：忽略错误
    // os.WriteFile("data.txt", []byte("hello"), 0644)
}

func goodExample() error {
    // 正确：检查错误
    err := fmt.Errorf("test")
    if err != nil {
        return err
    }
    return nil
}

// ========== 2. 不要忽略错误 ==========
func process() {
    result, err := mightFail()

    // 如果确实要忽略，使用 _
    _ = result

    // 明确处理
    if err != nil {
        log.Println(err)
    }
}

// ========== 3. 错误包装提供上下文 ==========
func loadData(id string) error {
    data, err := fetchData(id)
    if err != nil {
        return fmt.Errorf("加载数据 %s 失败：%w", id, err)
    }
    _ = data
    return nil
}

// ========== 4. 提前返回 (Guard Clause) ==========
func processUser(id int) error {
    if id <= 0 {
        return errors.New("无效的用户 ID")
    }

    // 主逻辑
    return nil
}

// ========== 5. 不要在错误信息中暴露敏感信息 ==========
// 错误：return fmt.Errorf("密码 %s 不正确", password)
// 正确：return errors.New("认证失败")

// ========== 6. 统一错误处理 ==========
// 在应用层统一处理错误日志、格式化等

func mightFail() (string, error) {
    return "", errors.New("失败了")
}

var log struct{ Println func(...interface{}) }

// ========== 7. 使用错误类型区分错误 ==========
type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s 未找到：%s", e.Resource, e.ID)
}

func handleErr(err error) {
    var notFound *NotFoundError
    if errors.As(err, &notFound) {
        // 处理 404
    } else if errors.Is(err, errors.ErrUnsupported) {
        // 处理不支持的操作
    }
}
```

---

## 2.4 包管理

### Go Modules 详解

```go
// go.mod 文件结构
/*
module example.com/myapp

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/stretchr/testify v1.8.4
)

require (
    github.com/bytedance/sonic v1.9.1 // indirect
    github.com/gabriel-vasile/mimetype v1.4.2 // indirect
)

replace github.com/old/package => github.com/new/package v1.0.0

exclude github.com/bad/package v1.0.0

retract v1.0.0
*/

// ========== 常用命令 ==========
/*
# 初始化
go mod init <module-name>

# 整理依赖
go mod tidy

# 下载依赖
go mod download

# 查看依赖
go list -m all

# 查看为什么需要某个依赖
go mod why github.com/gin-gonic/gin

# 升级依赖
go get github.com/gin-gonic/gin@latest
go get github.com/gin-gonic/gin@v1.9.0

# 降级依赖
go get github.com/gin-gonic/gin@v1.8.0

# 移除依赖
go mod edit -droprequire=github.com/some/package

# 查看依赖图
go mod graph

# 验证依赖
go mod verify

# 清理缓存
go clean -modcache
*/
```

### go.mod 与 go.sum 解析

```go
// ========== go.mod 详解 ==========
/*
module github.com/username/myproject  // 模块路径

go 1.21  // Go 版本

require (
    // 直接依赖
    github.com/gin-gonic/gin v1.9.1

    // 间接依赖 (当前模块未直接使用，但依赖的依赖需要)
    golang.org/x/sys v0.8.0 // indirect
)

// 替换依赖 (用于本地测试或 fork)
replace github.com/gin-gonic/gin => ./forks/gin

// 或者替换为特定版本
replace github.com/old/pkg => github.com/new/pkg v1.0.0

// 排除特定版本
exclude github.com/problematic/pkg v1.0.0

// 撤回已发布的版本
retract v1.0.0
retract [v1.0.0, v1.1.0]
*/

// ========== go.sum 详解 ==========
/*
go.sum 文件包含所有依赖的哈希值
用于验证下载依赖的完整性

格式:
github.com/gin-gonic/gin v1.9.1 h1:xxx...
github.com/gin-gonic/gin v1.9.1/go.mod h1:yyy...

- 第一行是模块内容的哈希
- 第二行是 go.mod 文件的哈希

go.sum 应该提交到版本控制
*/
```

### 依赖版本管理

```go
// ========== 版本选择 ==========
/*
# 语义化版本
v1.2.3  // 主版本。次版本。修订版本

# 伪版本 (用于非 tag 提交)
v0.0.0-20230101120000-abcdef123456

# 获取特定版本
go get github.com/pkg@v1.0.0
go get github.com/pkg@latest      // 最新版
go get github.com/pkg@main         // main 分支
go get github.com/pkg@master       // master 分支
go get github.com/pkg@branch-name  // 特定分支
*/

// ========== 版本升级策略 ==========
/*
# 小版本升级 (bug 修复)
go get github.com/pkg@v1.0.1  // v1.0.0 -> v1.0.1

# 次版本升级 (新功能，向后兼容)
go get github.com/pkg@v1.1.0  // v1.0.0 -> v1.1.0

# 主版本升级 (可能有破坏性变更)
go get github.com/pkg@v2.0.0  // 需要修改导入路径
*/

// ========== 主版本 v2+ ==========
/*
// v1 版本
import "github.com/pkg"

// v2 版本需要修改导入路径
import "github.com/pkg/v2"

// go.mod 也需要声明
module github.com/pkg/v2
*/
```

### 私有仓库配置

```bash
# ========== GOPRIVATE 配置 ==========
# 告诉 Go 某些模块是私有的，不使用公共代理

# 临时设置
export GOPRIVATE=github.com/mycompany/*

# 永久配置
go env -w GOPRIVATE=github.com/mycompany/*

# ========== Git 配置 ==========
# 配置 Git 使用 SSH
git config --global url."git@github.com:".insteadOf "https://github.com/"

# ========== .netrc 配置 (自动认证) ==========
# ~/.netrc 文件
# machine github.com login username password token

# ========== SSH 密钥 ==========
# 确保 SSH 密钥已添加到 SSH agent
ssh-add ~/.ssh/id_rsa

# ========== 使用替代 ==========
# go.mod 中配置
# replace github.com/mycompany/private => git.example.com/private v1.0.0
```

### 发布自己的包

```go
// ========== 发布步骤 ==========
/*
1. 创建 GitHub 仓库
   - 仓库名即为模块名的一部分

2. 初始化 go.mod
   go mod init github.com/username/mypackage

3. 编写代码和文档
   - 添加 README.md
   - 添加 LICENSE
   - 编写示例代码

4. 打标签发布
   git tag v1.0.0
   git push origin v1.0.0

5. 其他人可以使用
   go get github.com/username/mypackage@v1.0.0
*/

// ========== 包的最佳实践 ==========
/*
1. 有意义的包名
   - 包名简短且有意义
   - 避免与标准库冲突

2. 提供文档
   - 包文档注释
   - 导出标识符的文档
   - 示例代码 (Example 函数)

3. 版本管理
   - 使用语义化版本
   - 遵循向后兼容原则

4. 测试
   - 提供单元测试
   - 保证测试覆盖率
*/
```

---

## 2.5 泛型 (Go 1.18+)

### 泛型语法基础

```go
package main

import "fmt"

// ========== 泛型函数 ==========
// [T any] 定义类型参数

// 打印任意类型的切片
func PrintSlice[T any](s []T) {
    for _, v := range s {
        fmt.Print(v, " ")
    }
    fmt.Println()
}

// 多类型参数
func Pair[T, U any](a T, b U) (T, U) {
    return a, b
}

func main() {
    // 类型推断 (Go 1.18+)
    PrintSlice([]int{1, 2, 3})
    PrintSlice([]string{"a", "b", "c"})

    // 显式指定类型
    PrintSlice[int]([]int{4, 5, 6})

    p := Pair(1, "hello")
    fmt.Println(p)  // (1, hello)
}
```

### 类型参数与类型约束

```go
package main

import "fmt"

// ========== 类型约束 ==========
// 约束定义类型参数必须实现的方法

// 自定义约束
type Number interface {
    ~int | ~float64
}

type Stringer interface {
    String() string
}

// 带方法的约束
type Constraint interface {
    ~int | ~string
    Stringer
}

// ========== 使用约束 ==========
func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}

func PrintStringers[T Stringer](items []T) {
    for _, item := range items {
        fmt.Println(item.String())
    }
}

// ========== 标准库约束 ==========
// constraints 包 (golang.org/x/exp/constraints)
/*
type Signed interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Unsigned interface {
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

type Integer interface {
    Signed | Unsigned
}

type Float interface {
    ~float32 | ~float64
}

type Ordered interface {
    Integer | Float | ~string
}
*/

func main() {
    ints := []int{1, 2, 3, 4, 5}
    fmt.Println("总和:", Sum(ints))

    floats := []float64{1.1, 2.2, 3.3}
    fmt.Println("总和:", Sum(floats))
}

// ========== 底层类型 (~) ==========
// ~int 表示底层类型是 int 的所有类型
type MyInt int

func UseMyInt() {
    var m MyInt = 10
    // Sum([]MyInt{m})  // 可以使用，因为 ~int
}
```

### 泛型函数与泛型类型

```go
package main

import "fmt"

// ========== 泛型类型 ==========
type Stack[T any] struct {
    items []T
}

func NewStack[T any]() *Stack[T] {
    return &Stack[T]{}
}

func (s *Stack[T]) Push(v T) {
    s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    idx := len(s.items) - 1
    v := s.items[idx]
    s.items = s.items[:idx]
    return v, true
}

func (s *Stack[T]) IsEmpty() bool {
    return len(s.items) == 0
}

// ========== 泛型接口 ==========
type Container[T any] interface {
    Add(T)
    Get() T
    IsEmpty() bool
}

// ========== 泛型方法 ==========
// Go 不支持泛型方法，只有泛型类型的方法
// 类型已经泛型化，方法不需要再泛型化

func main() {
    // 整数栈
    intStack := NewStack[int]()
    intStack.Push(1)
    intStack.Push(2)
    val, ok := intStack.Pop()
    fmt.Println(val, ok)  // 2 true

    // 字符串栈
    strStack := NewStack[string]()
    strStack.Push("hello")
    strStack.Push("world")
    s, _ := strStack.Pop()
    fmt.Println(s)  // world

    // ========== 类型推断 ==========
    // Go 可以从函数参数推断类型参数
    PrintPair(1, 2)      // T 推断为 int
    PrintPair("a", "b")  // T 推断为 string
}

func PrintPair[T any](a, b T) {
    fmt.Println(a, b)
}
```

### 标准库中的泛型

```go
package main

import (
    "cmp"
    "fmt"
    "slices"
)

func main() {
    // ========== slices 包 (Go 1.21+) ==========
    nums := []int{3, 1, 4, 1, 5, 9, 2, 6}

    // 排序
    slices.Sort(nums)
    fmt.Println(nums)  // [1 1 2 3 4 5 6 9]

    // 检查是否有序
    fmt.Println(slices.IsSorted(nums))  // true

    // 查找
    idx := slices.Index(nums, 4)
    fmt.Println(idx)  // 4

    // 二分查找
    idx2 := slices.BinarySearch(nums, 5)
    fmt.Println(idx2)  // 5

    // 删除
    nums = slices.Delete(nums, 0, 2)
    fmt.Println(nums)  // [2 3 4 5 6 9]

    // 替换
    slices.Replace(nums, 1, 2, 100, 200)
    fmt.Println(nums)  // [2 100 200 5 6 9]

    // ========== maps 包 (Go 1.21+) ==========
    m := map[string]int{"a": 1, "b": 2}

    keys := slices.Sorted(maps.Keys(m))
    fmt.Println(keys)  // [a b]

    // ========== 自定义泛型工具 ==========
    // 使用 cmp 包进行比较
    type Person struct {
        Name string
        Age  int
    }

    people := []Person{
        {"Alice", 30},
        {"Bob", 25},
        {"Charlie", 35},
    }

    // 按年龄排序
    slices.SortFunc(people, func(a, b Person) int {
        return cmp.Compare(a.Age, b.Age)
    })

    fmt.Println(people)
}
```

### 泛型使用场景与限制

```go
package main

import "fmt"

// ========== 适用场景 ==========

// 1. 数据结构 (容器)
type Queue[T any] struct {
    items []T
}

func (q *Queue[T]) Enqueue(v T) {
    q.items = append(q.items, v)
}

func (q *Queue[T]) Dequeue() (T, bool) {
    if len(q.items) == 0 {
        var zero T
        return zero, false
    }
    v := q.items[0]
    q.items = q.items[1:]
    return v, true
}

// 2. 算法函数
func Filter[T any](slice []T, pred func(T) bool) []T {
    var result []T
    for _, v := range slice {
        if pred(v) {
            result = append(result, v)
        }
    }
    return result
}

func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

// 3. 类型安全的包装
type Result[T any] struct {
    Value T
    Error error
}

// ========== 限制 ==========
/*
1. 不能有方法重载
   泛型类型的方法不能再次泛型化

2. 不能使用类型参数作为其他类型的底层类型
   type MySlice[T any] []T  // 可以
   type IntSlice int        // 可以
   type GenericIntSlice[T any] int  // 不行

3. 不能访问未定义在约束中的方法
   func Call[T any](t T) {
       t.SomeMethod()  // 错误！除非约束中有
   }

4. 不能有循环依赖
   类型参数不能有循环引用

5. 不能使用反射获取类型参数名称
   运行时会擦除类型信息
*/

// ========== 何时使用泛型 ==========
/*
使用泛型:
1. 需要编写适用于多种类型的通用代码
2. 类型安全很重要
3. 性能要求高 (避免 interface{} 的装箱拆箱)

不使用泛型:
1. 只有一种类型使用
2. 类型不确定且复杂
3. 代码可读性会降低
*/

func main() {
    // 使用示例
    nums := []int{1, 2, 3, 4, 5, 6}

    // 过滤偶数
    evens := Filter(nums, func(n int) bool {
        return n%2 == 0
    })
    fmt.Println(evens)  // [2 4 6]

    // Map 操作
    doubled := Map(nums, func(n int) int {
        return n * 2
    })
    fmt.Println(doubled)  // [2 4 6 8 10 12]

    // 队列
    q := &Queue[string]{}
    q.Enqueue("first")
    q.Enqueue("second")
    val, _ := q.Dequeue()
    fmt.Println(val)  // first
}
```

---

## 第二部分完

接下来可以继续学习第三部分：并发编程（Goroutine、Channel、同步原语等）
