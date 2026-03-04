# 第一部分：Go 语言基础

## 1.1 环境搭建与工具链

### Go 安装与版本管理

#### macOS 安装
```bash
# 使用 Homebrew 安装
brew install go

# 验证安装
go version
```

#### Linux 安装
```bash
# 下载并安装
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPATH=$HOME/go' >> ~/.bashrc
source ~/.bashrc
```

#### Windows 安装
从 https://go.dev/dl/ 下载安装程序，按向导完成安装。

#### 版本管理工具 (goenv/asdf)

```bash
# 使用 goenv (类似 pyenv)
git clone https://github.com/go-nv/goenv.git ~/.goenv
echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bashrc
echo 'export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bashrc
echo 'eval "$(goenv init -)"' >> ~/.bashrc
source ~/.bashrc

# 安装指定版本
goenv install 1.21.0
goenv install 1.20.0

# 设置全局版本
goenv global 1.21.0

# 设置项目版本
goenv local 1.20.0
```

### GOPATH 与 Go Modules

#### GOPATH (旧模式)
```bash
# GOPATH 结构
GOPATH/
├── src/          # 源代码目录
├── pkg/          # 编译后的包
└── bin/          # 可执行文件
```

#### Go Modules (推荐，Go 1.11+)
```bash
# 初始化模块
go mod init <module-name>

# go.mod 文件结构
module example.com/myapp

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
)

# 常用命令
go mod tidy      # 整理依赖
go mod download  # 下载依赖
go mod vendor    # 将依赖复制到 vendor 目录
go mod graph     # 查看依赖图
go mod why       # 查看为什么需要某个依赖
go mod verify    # 验证依赖完整性
```

### 常用开发工具

#### VS Code + Go 插件
1. 安装 VS Code
2. 安装 Go 插件 (Go Team at Google)
3. 配置 settings.json:
```json
{
    "go.formatTool": "goimports",
    "go.lintTool": "golangci-lint",
    "go.useLanguageServer": true,
    "gopls": {
        "ui.semanticTokens": true,
        "ui.completion.usePlaceholders": true
    }
}
```

#### GoLand (JetBrains)
- 开箱即用的 Go IDE
- 内置调试器、测试工具
- 支持代码重构、数据库工具

### go 命令详解

```bash
# 运行 Go 程序
go run main.go

# 编译程序
go build -o app main.go

# 安装到 GOPATH/bin
go install

# 运行测试
go test ./...
go test -v ./...      # 详细输出
go test -race ./...   # 竞态检测

# 格式化代码
go fmt ./...
goimports -w .        # 更智能的格式化

# 代码检查
go vet ./...

# 查看包信息
go list -f '{{.Imports}}' ./...

# 生成文档
go doc fmt.Println

# 生成性能分析
go test -bench=. -cpuprofile=cpu.out
```

### 项目结构规范

```
project/
├── cmd/                    # 可执行应用入口
│   └── myapp/
│       └── main.go
├── internal/               # 私有包 (不可被外部导入)
│   ├── config/
│   ├── handler/
│   └── service/
├── pkg/                    # 公共包 (可被外部导入)
│   └── utils/
├── api/                    # API 定义 (proto, swagger)
├── configs/                # 配置文件
├── scripts/                # 脚本文件
├── test/                   # 测试文件
├── go.mod                  # 模块定义
├── go.sum                  # 依赖校验
├── Makefile                # 构建脚本
└── README.md               # 项目说明
```

---

## 1.2 语法基础

### 变量与常量声明

```go
package main

import "fmt"

// 包级变量
var globalVar int = 10
var globalVar2 = 20  // 类型推断
var name, age = "go", 18  // 多变量声明

// 短变量声明只能在函数内使用
// var short = "error"  // 错误：必须在函数内

// 常量
const Pi = 3.14159
const (
    StatusOK = 200
    StatusNotFound = 404
    StatusServerError = 500
)

// iota 枚举
const (
    Sunday = iota  // 0
    Monday         // 1
    Tuesday        // 2
    Wednesday      // 3
    Thursday       // 4
    Friday         // 5
    Saturday       // 6
)

// iota 表达式
const (
    _ = iota          // 0 (跳过)
    KB = 1 << (10*iota)  // 1 << 10 = 1024
    MB                 // 1 << 20 = 1048576
    GB                 // 1 << 30
)

func main() {
    // 短变量声明 (最常用)
    x := 42
    name := "Go"
    active := true

    // 多变量声明
    var a, b int = 1, 2
    c, d := 3, 4

    // 交换值
    a, b = b, a

    // 空白标识符 (忽略返回值)
    _, val := getValue()

    fmt.Println(x, name, active, a, b, c, d, val)
}

func getValue() (string, int) {
    return "result", 100
}
```

### 基本数据类型

```go
package main

import (
    "fmt"
    "math"
    "unicode/utf8"
)

func main() {
    // ========== 整数类型 ==========
    var i8 int8 = 127      // -128 ~ 127
    var i16 int16 = 32767  // -32768 ~ 32767
    var i32 int32 = 2147483647
    var i64 int64 = math.MaxInt64
    var i int = 100        // 根据平台 32 或 64 位

    var u8 uint8 = 255     // 0 ~ 255 (alias: byte)
    var u32 uint32 = 4294967295
    var u64 uint64 = 18446744073709551615

    // ========== 浮点类型 ==========
    var f32 float32 = 3.14
    var f64 float64 = 3.141592653589793
    var c64 complex64 = 1 + 2i
    var c128 complex128 = 1 + 2i

    // ========== 布尔类型 ==========
    var isTrue bool = true
    var isFalse bool = false

    // 布尔运算
    fmt.Println(true && false)  // false
    fmt.Println(true || false)  // true
    fmt.Println(!true)          // false

    // ========== 字符串类型 ==========
    var str string = "Hello, Go!"

    // 字符串是不可变的字节序列
    // str[0] = 'h'  // 编译错误！

    // 字符串长度 (字节数)
    fmt.Println(len(str))  // 12

    // 中文字符数
    fmt.Println(utf8.RuneCountInString("你好 Go"))  // 4

    // 子串
    sub := str[0:5]  // "Hello"

    // 原始字符串 (保留换行和特殊字符)
    raw := `这是原始字符串
可以跨越多行
不需要转义 " 和 \`

    // 格式化字符串
    name := "Go"
    version := 1.21
    msg := fmt.Sprintf("%s version %.2f", name, version)

    // ========== 类型别名 ==========
    type (
        Integer int
        Text string
        Predicate func(int) bool
    )

    var num Integer = 42
    var text Text = "hello"

    fmt.Println(num, text, msg, sub, raw)
}
```

### 运算符与表达式

```go
package main

import "fmt"

func main() {
    // ========== 算术运算符 ==========
    a, b := 17, 5

    fmt.Println(a + b)  // 加：22
    fmt.Println(a - b)  // 减：12
    fmt.Println(a * b)  // 乘：85
    fmt.Println(a / b)  // 除：3 (整数除法)
    fmt.Println(a % b)  // 取模：2
    fmt.Println(a / 3.0) // 浮点除法：5.666...

    // 自增自减 (只能作为语句)
    a++
    b--
    // ++a  // 错误！Go 没有前缀形式
    // a += 1  // 正确，等价于 a++

    // ========== 关系运算符 ==========
    x, y := 10, 20
    fmt.Println(x == y)  // false
    fmt.Println(x != y)  // true
    fmt.Println(x > y)   // false
    fmt.Println(x < y)   // true
    fmt.Println(x >= 10) // true
    fmt.Println(x <= 9)  // false

    // ========== 逻辑运算符 ==========
    p, q := true, false
    fmt.Println(p && q)  // AND: false
    fmt.Println(p || q)  // OR: true
    fmt.Println(!p)      // NOT: false
    fmt.Println(p && q)  // 短路求值

    // ========== 位运算符 ==========
    m, n := 12, 5  // 1100, 0101

    fmt.Println(m & n)   // AND: 4 (0100)
    fmt.Println(m | n)   // OR: 13 (1101)
    fmt.Println(m ^ n)   // XOR: 9 (1001)
    fmt.Println(m &^ n)  // 位清除：8 (1000)
    fmt.Println(m << 1)  // 左移：24 (11000)
    fmt.Println(m >> 2)  // 右移：3 (11)

    // ========== 赋值运算符 ==========
    val := 10
    val += 5  // val = val + 5
    val -= 3  // val = val - 3
    val *= 2  // val = val * 2
    val /= 4  // val = val / 4
    val %= 3  // val = val % 3
    val &= 7  // val = val & 7
    val |= 8  // val = val | 8
    val ^= 1  // val = val ^ 1

    // ========== 其他运算符 ==========
    ptr := &val  // 取地址
    refVal := *ptr  // 解引用

    // 下标运算符
    arr := []int{1, 2, 3}
    first := arr[0]

    // 类型转换
    var f float64 = 3.9
    i := int(f)  // 3 (截断，不是四舍五入)

    fmt.Println("Result:", i, refVal, first)
}
```

### 格式化输入输出

```go
package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    // ========== Print 系列 ==========
    name := "Go"
    age := 13

    fmt.Print("Hello ", name, "!\n")
    fmt.Println("Hello", name, "!")  // 自动添加空格和换行
    fmt.Printf("Name: %s, Age: %d\n", name, age)

    // ========== 格式化动词 ==========
    // 布尔
    fmt.Printf("%t\n", true)          // true

    // 整数
    fmt.Printf("%d\n", 42)           // 十进制：42
    fmt.Printf("%b\n", 42)           // 二进制：101010
    fmt.Printf("%o\n", 42)           // 八进制：52
    fmt.Printf("%x\n", 42)           // 十六进制：2a
    fmt.Printf("%X\n", 42)           // 大写十六进制：2A
    fmt.Printf("%c\n", 65)           // 字符：A

    // 浮点数
    fmt.Printf("%f\n", 3.14159)      // 3.141590
    fmt.Printf("%.2f\n", 3.14159)    // 3.14
    fmt.Printf("%e\n", 123456.0)     // 1.234560e+05
    fmt.Printf("%g\n", 123456.0)     // 123456 (自动选择)

    // 字符串
    fmt.Printf("%s\n", "hello")      // hello
    fmt.Printf("%q\n", "hello")      // "hello" (带引号)
    fmt.Printf("%x\n", "hello")      // 68656c6c6f
    fmt.Printf("%p\n", &name)        // 指针地址

    // 接口/任意类型
    fmt.Printf("%v\n", age)          // 默认格式
    fmt.Printf("%+v\n", age)         // 带字段名 (结构体)
    fmt.Printf("%#v\n", age)         // Go 语法格式
    fmt.Printf("%T\n", age)          // 类型名

    // 宽度和精度
    fmt.Printf("%10d\n", 42)         // 右对齐，宽度 10
    fmt.Printf("%010d\n", 42)        // 补零：0000000042
    fmt.Printf("%-10d\n", 42)        // 左对齐

    // ========== Scan 系列 (输入) ==========
    var input string
    var number int

    // 从标准输入读取
    fmt.Print("Enter your name: ")
    fmt.Scan(&input)

    fmt.Print("Enter your age: ")
    fmt.Scan(&number)

    // 读取整行
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Enter a sentence: ")
    line, _ := reader.ReadString('\n')
    line = strings.TrimSpace(line)

    // 格式化扫描
    var a, b int
    fmt.Println("Enter two numbers (space separated):")
    fmt.Scanf("%d %d", &a, &b)

    // 从字符串扫描
    var x, y int
    fmt.Sscanf("10 20", "%d %d", &x, &y)
    fmt.Println(x, y)  // 10 20

    fmt.Printf("Name: %s, Age: %d, Line: %s\n", input, number, line)
}
```

### 注释规范

```go
// ========== 单行注释 ==========
// 这是单行注释
// 用于解释代码逻辑

x := 42  // 行尾注释：说明 x 的含义

// ========== 多行注释 ==========
/*
 * 这是多行注释
 * 适用于大段说明
 * 但不如单行注释常用
 */

// ========== 文档注释规范 ==========

// Package main 是应用程序入口包
package main

// FormatUser 格式化用户信息
// 参数:
//   - name: 用户姓名
//   - age: 用户年龄
// 返回:
//   - string: 格式化后的用户信息
// 示例:
//   FormatUser("Alice", 25) // 返回 "Alice (25)"
func FormatUser(name string, age int) string {
    return fmt.Sprintf("%s (%d)", name, age)
}

// Config 应用配置结构
// 用于存储应用程序的配置信息
type Config struct {
    // Name 应用名称
    Name string

    // Version 应用版本号
    Version string

    // Debug 是否启用调试模式
    Debug bool

    // Port 服务端口
    Port int `json:"port" yaml:"port"`
}

// ========== 注释最佳实践 ==========

// 好的注释：解释为什么，而不是做什么
// 使用哈希算法验证数据完整性
if checksum != expectedChecksum {
    return errors.New("data corrupted")
}

// 避免冗余注释
x := x + 1  // x 加 1  <- 不好的注释

// 待办事项标记
// TODO(baxiang): 优化这个算法的性能
// FIXME: 修复并发竞态条件
// HACK: 临时解决方案，需要重构
// XXX: 危险！需要仔细检查
// NOTE: 重要说明
```

---

## 1.3 流程控制

### if/else 条件语句

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // ========== 基本 if 语句 ==========
    age := 20

    if age >= 18 {
        fmt.Println("成年人")
    }

    // ========== if-else ==========
    if age >= 60 {
        fmt.Println("老年人")
    } else if age >= 18 {
        fmt.Println("成年人")
    } else if age >= 12 {
        fmt.Println("青少年")
    } else {
        fmt.Println("儿童")
    }

    // ========== if 初始化语句 ==========
    // 可以在 if 前执行一个简短语句
    if hour := time.Now().Hour(); hour < 12 {
        fmt.Println("上午好")
    } else if hour < 18 {
        fmt.Println("下午好")
    } else {
        fmt.Println("晚上好")
    }
    // hour 在这里不可访问

    // ========== if 没有三目运算符 ==========
    // Go 没有 ternary operator: condition ? a : b
    // 使用 if-else 或辅助函数

    // ========== 常见模式 ==========

    // 1. 错误检查
    result, err := doSomething()
    if err != nil {
        fmt.Println("出错了:", err)
        return
    }

    // 2. 提前返回 (Guard Clause)
    if err := validate(); err != nil {
        return err
    }
    // 主逻辑

    // 3. 类型断言检查
    // if val, ok := iface.(Type); ok { ... }
}

func doSomething() (string, error) {
    return "result", nil
}

func validate() error {
    return nil
}
```

### switch 语句

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    // ========== 基本 switch ==========
    day := time.Now().Weekday()

    switch day {
    case time.Monday:
        fmt.Println("周一，新的开始")
    case time.Tuesday, time.Wednesday:
        fmt.Println("周二或周三")
    case time.Thursday, time.Friday:
        fmt.Println("周四或周五，快周末了")
    default:
        fmt.Println("周末！")
    }

    // ========== switch 带初始化 ==========
    switch os := runtime.GOOS; os {
    case "darwin":
        fmt.Println("macOS")
    case "linux":
        fmt.Println("Linux")
    case "windows":
        fmt.Println("Windows")
    default:
        fmt.Printf("其他：%s\n", os)
    }

    // ========== 无条件 switch (类似 if-else) ==========
    num := 75

    switch {
    case num < 0:
        fmt.Println("负数")
    case num < 50:
        fmt.Println("0-49")
    case num < 100:
        fmt.Println("50-99")
    default:
        fmt.Println("100 以上")
    }

    // ========== fallthrough (贯穿) ==========
    // Go 的 switch 默认有 break，需要 fallthrough 才能贯穿
    score := 85

    switch {
    case score >= 90:
        fmt.Println("优秀")
        // fallthrough  // 取消注释可贯穿
    case score >= 80:
        fmt.Println("良好")
        fallthrough
    case score >= 60:
        fmt.Println("及格")
        fallthrough
    default:
        fmt.Println("评定完成")
    }
    // 输出：良好、及格、评定完成

    // ========== switch 返回类型 ==========
    switch t := interface{}(time.Now()); t.(type) {
    default:
        fmt.Printf(" unexpected type %T\n", t)
    }
}
```

### Type Switch

```go
package main

import "fmt"

func main() {
    // ========== Type Switch ==========
    var values []interface{} = []interface{}{
        "hello", 42, 3.14, true, []int{1, 2, 3},
    }

    for _, v := range values {
        // type switch
        switch val := v.(type) {
        case int:
            fmt.Printf("整数：%d\n", val)
        case string:
            fmt.Printf("字符串：%s\n", val)
        case float64:
            fmt.Printf("浮点数：%.2f\n", val)
        case bool:
            fmt.Printf("布尔：%t\n", val)
        case []int:
            fmt.Printf("整数切片：%v\n", val)
        default:
            fmt.Printf("未知类型：%T - %v\n", val, val)
        }
    }

    // ========== 类型检查 (不获取值) ==========
    var x interface{} = "test"

    switch x.(type) {
    case string:
        fmt.Println("x 是字符串")
    case int:
        fmt.Println("x 是整数")
    default:
        fmt.Println("x 是其他类型")
    }

    // ========== 多类型合并 ==========
    var y interface{} = int32(100)

    switch y.(type) {
    case int, int8, int16, int32, int64:
        fmt.Println("有符号整数")
    case uint, uint8, uint16, uint32, uint64:
        fmt.Println("无符号整数")
    case float32, float64:
        fmt.Println("浮点数")
    default:
        fmt.Println("其他")
    }
}
```

### for 循环

```go
package main

import "fmt"

func main() {
    // ========== 传统 for 循环 ==========
    sum := 0
    for i := 0; i < 10; i++ {
        sum += i
    }
    fmt.Println("0-9 的和:", sum)  // 45

    // ========== while 风格 ==========
    count := 0
    for count < 5 {
        fmt.Println(count)
        count++
    }

    // ========== 无限循环 ==========
    step := 0
    for {
        step++
        if step >= 3 {
            break
        }
    }
    fmt.Println("无限循环结束，step =", step)

    // ========== range 遍历 ==========

    // 遍历数组/切片
    nums := []int{10, 20, 30, 40}
    for index, value := range nums {
        fmt.Printf("nums[%d] = %d\n", index, value)
    }

    // 只需要索引
    for i := range nums {
        fmt.Println("索引:", i)
    }

    // 只需要值 (忽略索引)
    for _, v := range nums {
        fmt.Println("值:", v)
    }

    // 遍历 Map
    m := map[string]int{"a": 1, "b": 2, "c": 3}
    for key, val := range m {
        fmt.Printf("m[%s] = %d\n", key, val)
    }

    // 遍历字符串 (rune)
    str := "你好 Go"
    for i, r := range str {
        fmt.Printf("%d: %c\n", i, r)
    }

    // 遍历 Channel
    ch := make(chan int, 3)
    ch <- 1
    ch <- 2
    ch <- 3
    close(ch)

    for v := range ch {
        fmt.Println("收到:", v)
    }
}
```

### break, continue, goto

```go
package main

import "fmt"

func main() {
    // ========== break ==========
    // 跳出当前循环或 switch

    for i := 0; i < 10; i++ {
        if i == 5 {
            break  // 跳出循环
        }
        fmt.Print(i, " ")  // 0 1 2 3 4
    }
    fmt.Println()

    // ========== 标签 break (跳出多层循环) ==========
Outer:
    for i := 0; i < 3; i++ {
        for j := 0; j < 3; j++ {
            if i+j == 3 {
                break Outer  // 直接跳出外层循环
            }
            fmt.Printf("(%d,%d) ", i, j)
        }
    }
    fmt.Println()
    // 输出：(0,0) (0,1) (0,2) (1,0) (1,1)

    // ========== continue ==========
    // 跳过本次迭代，进入下一次

    for i := 0; i < 5; i++ {
        if i == 2 {
            continue  // 跳过 2
        }
        fmt.Print(i, " ")  // 0 1 3 4
    }
    fmt.Println()

    // ========== goto ==========
    // 跳转到标签位置 (慎用)

    k := 0
    goto End
    fmt.Println("这行不会执行")

End:
    fmt.Println("跳转到这里，k =", k)

    // ========== 标签 continue ==========
    // Go 不支持标签 continue，这是设计决定
}
```

### defer 语句

```go
package main

import "fmt"

func main() {
    // ========== defer 基础 ==========
    // defer 延迟函数执行，直到当前函数返回

    fmt.Println("开始")
    defer fmt.Println("defer 1")  // 最后执行
    defer fmt.Println("defer 2")  // 倒数第二执行
    fmt.Println("结束")

    // 输出顺序：开始 -> 结束 -> defer 2 -> defer 1

    // ========== defer 参数立即求值 ==========
    x := 1
    defer fmt.Println("x =", x)  // x = 1 (立即求值)
    x = 2
    fmt.Println("x =", x)        // x = 2

    // ========== defer 在循环中的使用 ==========
    // 常见错误：defer 在循环内累积
    for i := 0; i < 3; i++ {
        defer fmt.Println("defer i =", i)
    }
    // 输出：defer i = 2, defer i = 1, defer i = 0 (LIFO)

    // ========== defer 常见用途 ==========

    // 1. 资源清理
    fmt.Println("\n资源清理示例:")
    cleanup()

    // 2. 解锁
    fmt.Println("\n解锁示例:")
    // withLock()

    // 3. 记录执行时间
    fmt.Println("\n性能分析示例:")
    profile()()

    // 4. recover 捕获 panic
    fmt.Println("\n异常处理示例:")
    safeFunc()
}

func cleanup() {
    // 模拟资源操作
    fmt.Println("获取资源")
    defer fmt.Println("释放资源")
    fmt.Println("使用资源")
}

func profile() func() {
    start := fmt.Sprintf("开始时间：%d", 0)
    fmt.Println(start)
    return func() {
        fmt.Println("结束时间:", 100)
    }
}

func safeFunc() {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("捕获 panic:", r)
        }
    }()

    panic("出错了!")
}
```

---

## 1.4 数组与切片

### 数组

```go
package main

import "fmt"

func main() {
    // ========== 数组声明 ==========

    // 指定长度
    var arr1 [5]int  // [0 0 0 0 0]

    // 声明并初始化
    arr2 := [5]int{1, 2, 3, 4, 5}
    arr3 := [5]int{1, 2, 3}      // [1 2 3 0 0]
    arr4 := [...]int{1, 2, 3, 4, 5}  // 编译器推断长度
    arr5 := [5]int{2: 10, 4: 20}     // [0 0 10 0 20]

    // ========== 数组访问 ==========
    arr2[0] = 100
    first := arr2[0]
    last := arr2[len(arr2)-1]

    fmt.Println(arr2)
    fmt.Println("长度:", len(arr2))
    fmt.Println("第一个:", first, "最后一个:", last)

    // ========== 数组遍历 ==========
    for i, v := range arr3 {
        fmt.Printf("arr3[%d] = %d\n", i, v)
    }

    // ========== 数组是值类型 ==========
    a := [3]int{1, 2, 3}
    b := a  // 复制整个数组
    b[0] = 100

    fmt.Println("a:", a)  // [1 2 3]
    fmt.Println("b:", b)  // [100 2 3]

    // 数组作为参数 (会复制)
    modifyArray(a)
    fmt.Println("modifyArray 后:", a)  // 不变

    // 传递指针
    modifyArrayPtr(&a)
    fmt.Println("modifyArrayPtr 后:", a)  // [100 2 3]

    // ========== 多维数组 ==========
    matrix := [2][3]int{
        {1, 2, 3},
        {4, 5, 6},
    }
    fmt.Println(matrix[0][1])  // 2
}

func modifyArray(arr [3]int) {
    arr[0] = 100
}

func modifyArrayPtr(arr *[3]int) {
    (*arr)[0] = 100
}
```

### 切片

```go
package main

import "fmt"

func main() {
    // ========== 切片创建 ==========

    // 从数组创建
    arr := [5]int{1, 2, 3, 4, 5}
    slice1 := arr[1:4]  // [2 3 4]
    slice2 := arr[:3]   // [1 2 3]
    slice3 := arr[2:]   // [3 4 5]
    slice4 := arr[:]    // [1 2 3 4 5]

    // 直接创建切片
    slice5 := []int{1, 2, 3, 4, 5}

    // make 创建
    slice6 := make([]int, 5)        // 长度 5，容量 5
    slice7 := make([]int, 3, 5)     // 长度 3，容量 5
    slice8 := make([]int, 0, 10)    // 长度 0，容量 10

    // ========== 切片属性 ==========
    s := []int{1, 2, 3, 4, 5}
    fmt.Println("长度:", len(s))   // 5
    fmt.Println("容量:", cap(s))   // 5

    // 切片是引用类型
    sub := s[1:4]  // [2 3 4]
    sub[0] = 200
    fmt.Println(s)    // [1 200 3 4 5] - 原切片也被修改
    fmt.Println(sub)  // [200 3 4]

    // ========== 切片重新切片 ==========
    s2 := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
    s2 = s2[2:8]     // [2 3 4 5 6 7]
    s2 = s2[1:4]     // [3 4 5]
    // 注意：容量是从原始切片开始位置计算的
}
```

### 切片的扩容机制

```go
package main

import "fmt"

func main() {
    // ========== append ==========
    s := []int{1, 2, 3}
    s = append(s, 4)          // [1 2 3 4]
    s = append(s, 5, 6, 7)    // [1 2 3 4 5 6 7]

    // 追加另一个切片
    s2 := []int{8, 9, 10}
    s = append(s, s2...)      // [1 2 3 4 5 6 7 8 9 10]

    // ========== 扩容机制 ==========
    // Go 1.18+ 扩容策略:
    // 1. 如果容量小于 256，新容量 = 旧容量 * 2
    // 2. 如果容量 >= 256，新容量 = 旧容量 * 1.25 (大约)
    // 3. 如果还不够，继续增长直到满足需求

    var empty []int
    fmt.Printf("len=%d, cap=%d\n", len(empty), cap(empty))

    for i := 0; i < 20; i++ {
        empty = append(empty, i)
        fmt.Printf("len=%d, cap=%d\n", len(empty), cap(empty))
    }

    // ========== 切片共享底层数组的陷阱 ==========
    a := make([]int, 5, 10)
    b := append(a, 100)  // b 与 a 共享底层数组
    a[0] = 999
    fmt.Println("b[0]:", b[0])  // 可能被影响

    // 避免方法：使用 copy 创建独立切片
    c := make([]int, len(a))
    copy(c, a)
}
```

### copy 和 append

```go
package main

import "fmt"

func main() {
    // ========== copy 函数 ==========
    src := []int{1, 2, 3, 4, 5}

    // 完整复制
    dst1 := make([]int, len(src))
    copy(dst1, src)
    dst1[0] = 100
    fmt.Println("src:", src)   // 不变
    fmt.Println("dst1:", dst1) // [100 2 3 4 5]

    // 部分复制 (copy 返回实际复制的元素数)
    dst2 := make([]int, 3)
    n := copy(dst2, src)
    fmt.Println("复制数量:", n, "dst2:", dst2)  // 3, [1 2 3]

    // 复制超过目标长度
    dst3 := make([]int, 10)
    copy(dst3, src)
    fmt.Println("dst3:", dst3)  // [1 2 3 4 5 0 0 0 0 0]

    // ========== 原地删除元素 ==========
    arr := []int{1, 2, 3, 4, 5}

    // 删除索引 2 的元素
    i := 2
    arr = append(arr[:i], arr[i+1:]...)
    fmt.Println("删除后:", arr)  // [1 2 4 5]

    // ========== 原地插入元素 ==========
    arr2 := []int{1, 2, 4, 5}
    i = 2
    arr2 = append(arr2[:i], append([]int{3}, arr2[i:]...)...)
    fmt.Println("插入后:", arr2)  // [1 2 3 4 5]

    // ========== 切片拼接 ==========
    s1 := []int{1, 2, 3}
    s2 := []int{4, 5, 6}
    s3 := append(s1, s2...)
    fmt.Println("拼接:", s3)  // [1 2 3 4 5 6]
}
```

### 多维数组与切片

```go
package main

import "fmt"

func main() {
    // ========== 二维切片 ==========

    // 方式 1: 字面量
    matrix1 := [][]int{
        {1, 2, 3},
        {4, 5, 6},
        {7, 8, 9},
    }

    // 方式 2: make 创建
    rows, cols := 3, 4
    matrix2 := make([][]int, rows)
    for i := range matrix2 {
        matrix2[i] = make([]int, cols)
    }

    // 赋值
    matrix2[0][0] = 1
    matrix2[1][2] = 100

    // ========== 遍历二维切片 ==========
    for i, row := range matrix1 {
        for j, val := range row {
            fmt.Printf("matrix[%d][%d] = %d\n", i, j, val)
        }
    }

    // ========== 不规则切片 (每行长度不同) ==========
    jagged := [][]int{
        {1},
        {2, 3},
        {4, 5, 6},
        {7, 8, 9, 10},
    }

    // ========== 扁平化多维切片 ==========
    // 一维模拟二维 (性能更好)
    flat := make([]int, rows*cols)
    // 访问 (i, j): flat[i*cols + j]
    flat[0*4+2] = 100  // 第 0 行第 2 列
    fmt.Println("扁平化访问:", flat)
}
```

---

## 1.5 Map

### Map 的创建与初始化

```go
package main

import "fmt"

func main() {
    // ========== Map 声明 ==========

    // 方式 1: make
    m1 := make(map[string]int)

    // 方式 2: 字面量
    m2 := map[string]int{
        "a": 1,
        "b": 2,
    }

    // 方式 3: 空 map
    m3 := map[string]int{}

    // 方式 4: nil map (不能直接赋值)
    var m4 map[string]int
    // m4["a"] = 1  // panic!

    // ========== 指定初始容量 ==========
    // 减少扩容开销
    m5 := make(map[string]int, 100)
}
```

### 增删改查操作

```go
package main

import "fmt"

func main() {
    // ========== 增/改 ==========
    scores := make(map[string]int)

    scores["Alice"] = 95
    scores["Bob"] = 87
    scores["Alice"] = 98  // 修改

    // ========== 查 ==========
    aliceScore := scores["Alice"]
    fmt.Println("Alice:", aliceScore)

    // 访问不存在的键
    fmt.Println("Carol:", scores["Carol"])  // 0 (零值)

    // 检查键是否存在
    val, ok := scores["Alice"]
    if ok {
        fmt.Println("存在:", val)
    }

    // 简洁写法
    if val, ok := scores["Bob"]; ok {
        fmt.Println("Bob 的分数:", val)
    }

    // ========== 删除 ==========
    delete(scores, "Bob")

    // 删除不存在的键 (不会报错)
    delete(scores, "Unknown")

    // ========== 获取长度 ==========
    fmt.Println("Map 大小:", len(scores))

    // ========== Map 是引用类型 ==========
    m := map[string]int{"a": 1}
    modifyMap(m)
    fmt.Println(m)  // a 被修改
}

func modifyMap(m map[string]int) {
    m["a"] = 100
    m["b"] = 200
}
```

### Map 的遍历

```go
package main

import (
    "fmt"
    "sort"
)

func main() {
    m := map[string]int{
        "a": 1,
        "b": 2,
        "c": 3,
    }

    // ========== 遍历 ==========
    // 注意：Map 遍历顺序是不确定的
    for key, val := range m {
        fmt.Printf("%s: %d\n", key, val)
    }

    // 只遍历键
    for key := range m {
        fmt.Println(key)
    }

    // ========== 有序遍历 ==========
    // 提取键并排序
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, m[k])
    }

    // ========== 遍历中删除 ==========
    for k := range m {
        if k == "b" {
            delete(m, k)
            // 删除后不要再使用 k
        }
    }
}
```

### Map 与切片的组合使用

```go
package main

import "fmt"

func main() {
    // ========== Map 的 value 是切片 ==========
    m1 := map[string][]int{
        "scores": {1, 2, 3},
        "ages":   {20, 30},
    }
    m1["scores"] = append(m1["scores"], 4)

    // ========== 切片的元素是 Map ==========
    m2 := map[string]int{"a": 1}
    slice := []map[string]int{m2, {"b": 2}}
    slice[0]["c"] = 3

    // ========== 统计词频 ==========
    words := []string{"a", "b", "a", "c", "b", "a"}
    freq := make(map[string]int)

    for _, word := range words {
        freq[word]++
    }
    fmt.Println(freq)  // map[a:3 b:2 c:1]

    // ========== 分组 ==========
    type Person struct {
        Name string
        Age  int
    }

    people := []Person{
        {"A", 20}, {"B", 25}, {"C", 20},
    }

    byAge := make(map[int][]Person)
    for _, p := range people {
        byAge[p.Age] = append(byAge[p.Age], p)
    }
    fmt.Println(byAge)
}
```

### sync.Map (并发安全)

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    // sync.Map 用于并发场景
    var m sync.Map

    // ========== 基本操作 ==========
    m.Store("a", 1)
    m.Store("b", 2)

    // 读取
    val, ok := m.Load("a")
    if ok {
        fmt.Println("a =", val)
    }

    // 读取或删除
    val, ok = m.LoadAndDelete("b")

    // 如果不存在则存储
    val, loaded := m.LoadOrStore("c", 3)

    // 遍历
    m.Range(func(key, value interface{}) bool {
        fmt.Printf("%v: %v\n", key, value)
        return true  // 继续遍历
    })

    // 删除
    m.Delete("a")
}
```

---

## 1.6 函数

### 函数定义与调用

```go
package main

import "fmt"

// ========== 基本函数 ==========
func add(a int, b int) int {
    return a + b
}

// ========== 简洁参数类型 ==========
// 相同类型可以合并
func multiply(a, b int) int {
    return a * b
}

// ========== 无返回值 ==========
func printSum(a, b int) {
    fmt.Println(a + b)
}

// ========== 无参数 ==========
func sayHello() {
    fmt.Println("Hello!")
}

func main() {
    result := add(3, 5)
    fmt.Println(result)  // 8

    printSum(10, 20)
    sayHello()
}
```

### 多返回值

```go
package main

import (
    "errors"
    "fmt"
)

// ========== 返回两个值 ==========
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("除数不能为 0")
    }
    return a / b, nil
}

// ========== 返回多个值 ==========
func getUserInfo(id int) (name string, age int, err error) {
    if id < 0 {
        return "", 0, errors.New("无效 ID")
    }
    return "Alice", 25, nil
}

// ========== 命名返回值 ==========
// 命名返回值会自动初始化为零值
func calculate(a, b int) (sum, product int) {
    sum = a + b
    product = a * b
    // 裸返回，返回命名变量
    return
}

// ========== 处理多返回值 ==========
func main() {
    // 接收所有返回值
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("错误:", err)
        return
    }
    fmt.Println(result)

    // 忽略某个返回值
    val, _ := getUserInfo(1)

    // 一次性接收
    s, p := calculate(3, 4)
    fmt.Println("和:", s, "积:", p)
}
```

### 可变参数

```go
package main

import "fmt"

// ========== 可变参数函数 ==========
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// ========== 混合参数 ==========
func printf(format string, args ...interface{}) {
    fmt.Printf(format, args...)
}

// ========== 可变参数必须在最后 ==========
// func wrong(a ...int, b string) {}  // 错误!

func main() {
    fmt.Println(sum())           // 0
    fmt.Println(sum(1))          // 1
    fmt.Println(sum(1, 2, 3))    // 6
    fmt.Println(sum(1, 2, 3, 4, 5))  // 15

    // 切片展开
    nums := []int{1, 2, 3, 4}
    fmt.Println(sum(nums...))    // 10

    // 部分展开
    more := []int{5, 6}
    fmt.Println(sum(append(nums, more...)...))
}
```

### 匿名函数

```go
package main

import "fmt"

func main() {
    // ========== 匿名函数赋值给变量 ==========
    add := func(a, b int) int {
        return a + b
    }
    fmt.Println(add(3, 5))  // 8

    // ========== 立即执行 ==========
    func() {
        fmt.Println("立即执行")
    }()

    func(msg string) {
        fmt.Println(msg)
    }("带参数")

    // ========== 作为函数参数 ==========
    numbers := []int{1, 2, 3, 4, 5}

    // 回调函数
    process(numbers, func(n int) int {
        return n * 2
    })

    // ========== 作为返回值 ==========
    multiplier := makeMultiplier(3)
    fmt.Println(multiplier(5))  // 15
}

func process(nums []int, transform func(int) int) {
    for i, n := range nums {
        nums[i] = transform(n)
    }
    fmt.Println(nums)
}

func makeMultiplier(factor int) func(int) int {
    return func(n int) int {
        return n * factor
    }
}
```

### 闭包与作用域

```go
package main

import "fmt"

func main() {
    // ========== 闭包基础 ==========
    // 闭包捕获外部变量

    counter := func() func() int {
        count := 0
        return func() int {
            count++
            return count
        }
    }()

    fmt.Println(counter())  // 1
    fmt.Println(counter())  // 2
    fmt.Println(counter())  // 3

    // ========== 闭包捕获引用 ==========
    x := 10
    f := func() {
        fmt.Println(x)
    }

    x = 20
    f()  // 输出 20 (捕获的是引用)

    // ========== 闭包陷阱 ==========
    // 循环中闭包捕获循环变量

    // 错误示例
    var funcs []func()
    for i := 0; i < 3; i++ {
        funcs = append(funcs, func() {
            fmt.Println(i)  // 都输出 3
        })
    }

    // 正确示例
    var funcs2 []func()
    for i := 0; i < 3; i++ {
        i := i  // 创建新变量
        funcs2 = append(funcs2, func() {
            fmt.Println(i)  // 输出 0, 1, 2
        })
    }

    // ========== 实用闭包示例 ==========

    // 1. 缓存
    cache := make(map[string]string)
    getOrCompute := func(key string) string {
        if val, ok := cache[key]; ok {
            return val
        }
        val := "computed:" + key
        cache[key] = val
        return val
    }

    // 2. 配置函数
    type Config struct {
        Host string
        Port int
    }

    config := &Config{}
    configure := func(fn func(*Config)) {
        fn(config)
    }

    configure(func(c *Config) {
        c.Host = "localhost"
        c.Port = 8080
    })

    fmt.Println(config)
}
```

---

## 1.7 指针

### 指针基础概念

```go
package main

import "fmt"

func main() {
    // ========== 指针声明 ==========
    var p *int  // 指向 int 的指针，值为 nil

    x := 42
    p = &x  // p 指向 x

    fmt.Println("x 的值:", x)      // 42
    fmt.Println("x 的地址:", &x)   // 0x...
    fmt.Println("p 的值:", p)      // 0x... (同 x 的地址)
    fmt.Println("p 指向的值:", *p)  // 42

    // ========== 指针运算 ==========
    *p = 100  // 修改 x 的值
    fmt.Println("x:", x)  // 100

    // ========== 指针的指针 ==========
    pp := &p
    fmt.Println("pp:", pp)
    fmt.Println("*pp:", *pp)   // p 的值
    fmt.Println("**pp:", **pp) // x 的值
}
```

### new 和 make 的区别

```go
package main

import "fmt"

func main() {
    // ========== new ==========
    // new(T): 分配内存，返回 *T，值为零值
    // 适用于任何类型

    p1 := new(int)      // *int, 值为 0
    p2 := new(string)   // *string, 值为 ""
    p3 := new([]int)    // *[]int, 值为 nil

    fmt.Println(*p1)  // 0
    fmt.Println(*p2)  // 空字符串
    fmt.Println(*p3)  // nil

    // ========== make ==========
    // make(T, args): 仅用于 slice, map, channel
    // 返回 T (不是指针)
    // 进行初始化

    s := make([]int, 5)    // []int, 长度 5
    m := make(map[string]int)  // map[string]int
    ch := make(chan int, 1)    // chan int

    fmt.Println(s)  // [0 0 0 0 0]
    fmt.Println(m)  // map[]

    // ========== 对比 ==========
    var slice1 *[]int = new([]int)  // 指向 nil 切片的指针
    var slice2 []int = make([]int, 5)  // 初始化的切片

    slice1 = &slice2  // 通常这样用
}
```

### 指针与函数参数传递

```go
package main

import "fmt"

func main() {
    // ========== 值传递 ==========
    x := 10
    modifyValue(x)
    fmt.Println("值传递后:", x)  // 10 (不变)

    // ========== 指针传递 ==========
    modifyPointer(&x)
    fmt.Println("指针传递后:", x)  // 100 (改变)

    // ========== 切片/Map/Channel ==========
    // 这些类型内部包含指针，所以是引用语义
    s := []int{1, 2, 3}
    modifySlice(s)
    fmt.Println("切片修改后:", s)  // [100 2 3]

    m := map[string]int{"a": 1}
    modifyMap(m)
    fmt.Println("Map 修改后:", m)  // map[a:100]
}

func modifyValue(x int) {
    x = 100
}

func modifyPointer(p *int) {
    *p = 100
}

func modifySlice(s []int) {
    s[0] = 100
}

func modifyMap(m map[string]int) {
    m["a"] = 100
}
```

### 指针的安全性

```go
package main

import "fmt"

func main() {
    // ========== Go 指针的安全特性 ==========
    // 1. 不能进行指针算术运算
    // p := &x
    // p++  // 编译错误

    // 2. 指针不能转换为其他类型指针 (需要 unsafe)
    // var p *int
    // var q *float64 = p  // 编译错误

    // 3. 空指针检查
    var p *int
    if p == nil {
        fmt.Println("p 是空指针")
    }

    // 4. 野指针 (Go 的 GC 会处理)
    func() {
        x := 10
        p = &x  // x 会逃逸到堆上
    }()
    // p 仍然有效，不会成为野指针
    fmt.Println(*p)  // 10

    // ========== 安全使用指针的准则 ==========
    // 1. 只在必要时使用指针
    // 2. 始终检查 nil
    // 3. 避免返回局部变量地址 (虽然 Go 会处理)
    // 4. 注意并发访问
}
```

---

## 第一部分完

接下来可以继续学习第二部分：核心特性（结构体、接口、错误处理等）
