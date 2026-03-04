# 第四部分：标准库精讲

## 4.1 常用标准库

### fmt - 格式化 I/O

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // ========== Print 系列 ==========
    fmt.Print("Hello")           // 不换行
    fmt.Println("Hello")         // 自动换行，参数间加空格
    fmt.Printf("Name: %s\n", "Go")  // 格式化输出

    // ========== Sprintf 系列 (返回字符串) ==========
    s := fmt.Sprintf("%s %d", "hello", 42)
    fmt.Println(s)

    // ========== Fprint 系列 (输出到 io.Writer) ==========
    fmt.Fprint(os.Stderr, "错误信息")
    fmt.Fprintf(os.Stderr, "错误码：%d\n", 500)

    // ========== Scan 系列 (输入) ==========
    var name string
    var age int

    fmt.Print("请输入姓名：")
    fmt.Scan(&name)

    fmt.Print("请输入年龄：")
    fmt.Scan(&age)

    // ========== 常用格式化动词 ==========
    /*
    %v     默认格式
    %+v    带字段名 (结构体)
    %#v    Go 语法格式
    %T     类型

    %t     布尔 (true/false)

    %d     十进制整数
    %b     二进制
    %o     八进制
    %x     十六进制 (小写)
    %X     十六进制 (大写)

    %s     字符串
    %q     带引号的字符串
    %x     字符串的十六进制

    %f     浮点数 (默认 6 位小数)
    %.2f   浮点数 (2 位小数)
    %e     科学计数法

    %p     指针地址
    */
}
```

### strings - 字符串处理

```go
package main

import (
    "fmt"
    "strings"
)

func main() {
    s := "Hello, Go 世界！"

    // ========== 查找 ==========
    strings.Contains(s, "Go")       // true
    strings.ContainsAny(s, "abc")   // false
    strings.HasPrefix(s, "Hello")   // true
    strings.HasSuffix(s, "世界！")    // true
    strings.Index(s, "Go")          // 7
    strings.IndexAny(s, "世界")      // 10
    strings.LastIndex(s, "l")       // 9

    // ========== 大小写转换 ==========
    strings.ToLower(s)              // "hello, go 世界！"
    strings.ToUpper(s)              // "HELLO, GO 世界！"
    strings.Title(s)                // 已废弃，使用 cases 包

    // ========== 修剪 ==========
    strings.TrimSpace("  hello  ")         // "hello"
    strings.Trim("!!hello!!", "!")         // "hello"
    strings.TrimLeft("  hello  ", " ")     // "hello  "
    strings.TrimRight("  hello  ", " ")    // "  hello"
    strings.TrimPrefix("Hello, Go", "Hello, ")  // "Go"
    strings.TrimSuffix("Hello.go", ".go")     // "Hello"

    // ========== 分割 ==========
    strings.Split("a,b,c", ",")           // ["a", "b", "c"]
    strings.SplitN("a,b,c", ",", 2)       // ["a", "b,c"]
    strings.Fields("a  b\tc\n")           // ["a", "b", "c"] (空白字符分割)

    // ========== 连接 ==========
    strings.Join([]string{"a", "b", "c"}, ",")  // "a,b,c"

    // ========== 替换 ==========
    strings.Replace("hello world", "world", "Go", 1)  // "hello Go"
    strings.ReplaceAll("hello hello", "hello", "hi")  // "hi hi"

    // ========== 重复 ==========
    strings.Repeat("ab", 3)  // "ababab"

    // ========== 比较 ==========
    strings.EqualFold("Hello", "HELLO")  // true (忽略大小写)
    strings.Compare("a", "b")            // -1 (a < b)

    // ========== 统计 ==========
    strings.Count("hello world", "l")    // 3

    // ========== Reader ==========
    reader := strings.NewReader("hello")
    buf := make([]byte, 2)
    reader.Read(buf)
    fmt.Println(string(buf))  // "he"
}
```

### strconv - 类型转换

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    // ========== 字符串转整数 ==========
    i, _ := strconv.Atoi("123")           // 123
    i64, _ := strconv.ParseInt("123", 10, 64)  // 123, 十进制，64 位
    u64, _ := strconv.ParseUint("123", 10, 64) // 123

    // 不同进制
    strconv.ParseInt("ff", 16, 64)        // 255
    strconv.ParseInt("1010", 2, 64)       // 10
    strconv.ParseInt("77", 8, 64)         // 63

    // ========== 整数转字符串 ==========
    strconv.Itoa(123)                     // "123"
    strconv.FormatInt(123, 10)            // "123"
    strconv.FormatInt(255, 16)            // "ff"
    strconv.FormatUint(123, 10)           // "123"

    // ========== 浮点数转换 ==========
    f, _ := strconv.ParseFloat("3.14", 64)     // 3.14
    strconv.FormatFloat(3.14, 'f', 2, 64)      // "3.14"
    strconv.FormatFloat(3.14, 'e', 2, 64)      // "3.14e+00"

    // ========== 布尔转换 ==========
    b, _ := strconv.ParseBool("true")          // true
    b, _ = strconv.ParseBool("1")              // true
    b, _ = strconv.ParseBool("false")          // false

    strconv.FormatBool(true)                   // "true"

    // ========== Quote ==========
    strconv.Quote("hello\nworld")              // "\"hello\\nworld\""
    strconv.Unquote("\"hello\\nworld\"")       // "hello\nworld"
}
```

### time - 时间处理

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // ========== 当前时间 ==========
    now := time.Now()
    fmt.Println(now)

    // ========== 创建时间 ==========
    // 年 月 日 时 分 秒 纳秒 时区
    t := time.Date(2024, 1, 15, 10, 30, 0, 0, time.Local)

    // UTC 时间
    utc := time.Now().UTC()

    // ========== 时间组件 ==========
    t.Year()           // 2024
    t.Month()          // time.January
    t.Day()            // 15
    t.Hour()           // 10
    t.Minute()         // 30
    t.Second()         // 0
    t.Nanosecond()     // 0
    t.Weekday()        // time.Monday

    // ========== 时间格式化 ==========
    // Go 使用固定的参考时间：2006-01-02 15:04:05
    t.Format("2006-01-02 15:04:05")           // "2024-01-15 10:30:00"
    t.Format("2006/01/02")                    // "2024/01/15"
    t.Format("15:04:05")                      // "10:30:00"
    t.Format(time.RFC3339)                    // ISO 8601

    // ========== 解析时间 ==========
    time.Parse("2006-01-02", "2024-01-15")
    time.Parse(time.RFC3339, "2024-01-15T10:30:00Z")

    // ========== 时间运算 ==========
    t.Add(time.Hour * 2)                      // 加 2 小时
    t.AddDate(0, 1, 0)                        // 加 1 个月
    t.AddDate(1, 0, 0)                        // 加 1 年

    // ========== 时间比较 ==========
    t1 := time.Now()
    t2 := t1.Add(time.Hour)

    t1.Before(t2)     // true
    t1.After(t2)      // false
    t1.Equal(t2)      // false

    // ========== 时间差 ==========
    t2.Sub(t1)        // time.Duration
    t2.Sub(t1).Hours()    // 1
    t2.Sub(t1).Minutes()  // 60
    t2.Sub(t1).Seconds()  // 3600

    // ========== Timer 和 Ticker ==========
    // 一次性定时器
    timer := time.NewTimer(time.Second)
    <-timer.C
    fmt.Println("1 秒后")

    // 周期性定时器
    ticker := time.NewTicker(500 * time.Millisecond)
    done := make(chan bool)

    go func() {
        for range ticker.C {
            fmt.Println("每 500ms")
        }
    }()

    go func() {
        time.Sleep(2 * time.Second)
        done <- true
    }()

    <-done
    ticker.Stop()

    // ========== Sleep ==========
    time.Sleep(100 * time.Millisecond)

    // ========== 时区 ==========
    loc, _ := time.LoadLocation("Asia/Shanghai")
    t.In(loc)
}
```

### encoding/json - JSON 编解码

```go
package main

import (
    "encoding/json"
    "fmt"
    "strings"
)

type Person struct {
    Name    string   `json:"name"`
    Age     int      `json:"age"`
    Email   string   `json:"email,omitempty"`  // 空值时省略
    Hobbies []string `json:"hobbies"`
    Address Address  `json:"address"`
}

type Address struct {
    City  string `json:"city"`
    Zip   string `json:"zip"`
}

func main() {
    // ========== 编码 (Marshal) ==========
    p := Person{
        Name:    "Alice",
        Age:     25,
        Email:   "alice@example.com",
        Hobbies: []string{"reading", "coding"},
        Address: Address{City: "Beijing", Zip: "100000"},
    }

    // 编码为 []byte
    data, err := json.Marshal(p)
    if err != nil {
        fmt.Println("错误:", err)
    }
    fmt.Println(string(data))

    // 带缩进的编码
    pretty, _ := json.MarshalIndent(p, "", "  ")
    fmt.Println(string(pretty))

    // ========== 解码 (Unmarshal) ==========
    jsonStr := `{"name":"Bob","age":30,"hobbies":["gaming"]}`

    var p2 Person
    err = json.Unmarshal([]byte(jsonStr), &p2)
    if err != nil {
        fmt.Println("错误:", err)
    }
    fmt.Println(p2)

    // ========== 解码为 map ==========
    var m map[string]interface{}
    json.Unmarshal([]byte(jsonStr), &m)
    fmt.Println(m["name"])

    // ========== 解码 JSON 数组 ==========
    arrJSON := `[{"name":"A"},{"name":"B"}]`
    var people []Person
    json.Unmarshal([]byte(arrJSON), &people)

    // ========== 自定义 Marshal/Unmarshal ==========
    type MyTime struct {
        time.Time
    }

    func (mt MyTime) MarshalJSON() ([]byte, error) {
        return json.Marshal(mt.Format("2006-01-02"))
    }

    func (mt *MyTime) UnmarshalJSON(data []byte) error {
        t, err := time.Parse("2006-01-02", strings.Trim(string(data), "\""))
        if err != nil {
            return err
        }
        mt.Time = t
        return nil
    }

    // ========== JSON 解码器 (流式) ==========
    decoder := json.NewDecoder(strings.NewReader(jsonStr))
    var p3 Person
    decoder.Decode(&p3)

    // ========== JSON 编码器 (流式) ==========
    encoder := json.NewEncoder(os.Stdout)
    encoder.Encode(p)
}
```

### io 与 bufio - I/O 操作

```go
package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"
)

func main() {
    // ========== Reader 接口 ==========
    // type Reader interface { Read(p []byte) (n int, err error) }

    // ========== Writer 接口 ==========
    // type Writer interface { Write(p []byte) (n int, err error) }

    // ========== 文件读写 ==========
    file, _ := os.Open("file.txt")
    defer file.Close()

    buf := make([]byte, 1024)
    n, _ := file.Read(buf)
    fmt.Println(string(buf[:n]))

    // ========== 复制 (Copy) ==========
    src, _ := os.Open("src.txt")
    defer src.Close()

    dst, _ := os.Create("dst.txt")
    defer dst.Close()

    io.Copy(dst, src)

    // ========== 带缓冲的复制 ==========
    io.CopyBuffer(dst, src, make([]byte, 32*1024))

    // ========== 限制复制 ==========
    io.CopyN(dst, src, 100)  // 只复制 100 字节

    // ========== Bufio 包 ==========

    // BufioReader
    reader := bufio.NewReader(file)
    line, _ := reader.ReadString('\n')  // 读取到换行

    // BufioWriter
    writer := bufio.NewWriter(dst)
    writer.WriteString("hello")
    writer.Flush()  // 必须刷新

    // ========== bytes 包 ==========
    // bytes.Buffer - 可增长的字节缓冲区
    var buf2 bytes.Buffer
    buf2.WriteString("hello")
    buf2.WriteByte(' ')
    buf2.Write([]byte("world"))

    // bytes.Reader - 从字节切片创建 Reader
    reader2 := bytes.NewReader([]byte("hello"))

    // 常用函数
    bytes.Equal([]byte("a"), []byte("a"))      // true
    bytes.Contains([]byte("hello"), []byte("ll"))  // true
    bytes.ToUpper([]byte("hello"))             // "HELLO"
    bytes.Split([]byte("a,b,c"), []byte(","))  // [][]byte

    // ========== strings.Builder (推荐) ==========
    var sb strings.Builder
    sb.WriteString("hello")
    sb.WriteString(" ")
    sb.WriteString("world")
    result := sb.String()

    // ========== ReadAll ==========
    content, _ := io.ReadAll(file)

    // ========== MultiReader / MultiWriter ==========
    // 合并多个 Reader
    mr := io.MultiReader(strings.NewReader("a"), strings.NewReader("b"))

    // 写入多个 Writer
    mw := io.MultiWriter(os.Stdout, os.Stderr)
    mw.Write([]byte("hello"))
}
```

### os/os.File - 文件系统操作

```go
package main

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
)

func main() {
    // ========== 文件操作 ==========

    // 打开文件
    file, err := os.Open("file.txt")
    if err != nil {
        fmt.Println("打开失败:", err)
    }
    defer file.Close()

    // 创建文件 (不存在则创建，存在则截断)
    file2, _ := os.Create("new.txt")
    defer file2.Close()

    // 打开文件 (可指定模式)
    file3, _ := os.OpenFile("file.txt", os.O_RDWR|os.O_APPEND, 0644)
    defer file3.Close()

    // 写入
    file3.WriteString("hello\n")
    file3.Write([]byte("world\n"))

    // 读取
    content, _ := io.ReadAll(file)

    // ========== 文件信息 ==========
    info, _ := os.Stat("file.txt")
    fmt.Println(info.Name())     // 文件名
    fmt.Println(info.Size())     // 大小 (字节)
    fmt.Println(info.Mode())     // 权限
    fmt.Println(info.ModTime())  // 修改时间
    fmt.Println(info.IsDir())    // 是否目录

    // ========== 目录操作 ==========
    // 创建目录
    os.Mkdir("dir1", 0755)
    os.MkdirAll("a/b/c", 0755)  // 递归创建

    // 删除
    os.Remove("file.txt")        // 删除文件/空目录
    os.RemoveAll("a")            // 递归删除

    // 读取目录
    entries, _ := os.ReadDir(".")
    for _, e := range entries {
        fmt.Println(e.Name(), e.IsDir())
    }

    // ========== 路径操作 ==========
    // 获取当前目录
    wd, _ := os.Getwd()
    fmt.Println("工作目录:", wd)

    // 切换目录
    os.Chdir("/tmp")

    // ========== 环境变量 ==========
    os.Setenv("MY_VAR", "value")
    val := os.Getenv("MY_VAR")
    fmt.Println(val)

    // 获取所有环境变量
    envs := os.Environ()

    // ========== 文件权限 ==========
    os.Chmod("file.txt", 0755)
    os.Chown("file.txt", 1000, 1000)  // uid, gid
}
```

### filepath - 路径处理

```go
package main

import (
    "fmt"
    "path/filepath"
)

func main() {
    // ========== 路径拼接 ==========
    // 自动使用正确的路径分隔符
    path := filepath.Join("home", "user", "docs", "file.txt")
    fmt.Println(path)  // Unix: home/user/docs/file.txt

    // ========== 分隔符 ==========
    filepath.Separator   // '/'
    filepath.ListSeparator  // ':'

    // ========== 路径清理 ==========
    filepath.Clean("/home/../home/./user")  // "/home/user"

    // ========== 绝对路径 ==========
    abs, _ := filepath.Abs("file.txt")
    fmt.Println(abs)

    // ========== 相对路径 ==========
    rel, _ := filepath.Rel("/home/user", "/home/user/docs/file.txt")
    fmt.Println(rel)  // "docs/file.txt"

    // ========== 目录和文件名 ==========
    filepath.Dir("/home/user/file.txt")    // "/home/user"
    filepath.Base("/home/user/file.txt")   // "file.txt"
    filepath.Ext("/home/user/file.txt")    // ".txt"

    // ========== 文件名处理 ==========
    name := "file.tar.gz"
    filepath.Base(name)        // "file.tar.gz"
    filepath.Ext(name)         // ".gz"

    // 去除扩展名
    trimmed := name[:len(name)-len(filepath.Ext(name))]
    fmt.Println(trimmed)       // "file.tar"

    // ========== Glob 匹配 ==========
    // 匹配文件模式
    matches, _ := filepath.Glob("*.txt")
    fmt.Println(matches)

    matches2, _ := filepath.Glob("docs/*.md")
    fmt.Println(matches2)

    // ========== Walk 遍历目录 ==========
    filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        fmt.Println(path, info.Size())
        return nil
    })

    // WalkDir (Go 1.16+, 更高效)
    filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
        if err != nil {
            return err
        }
        fmt.Println(path, d.IsDir())
        return nil
    })
}
```

---

## 4.2 网络编程

### net 包基础

```go
package main

import (
    "fmt"
    "net"
)

func main() {
    // ========== IP 地址操作 ==========
    ip := net.ParseIP("192.168.1.1")
    fmt.Println(ip)

    // IPv4
    ip4 := ip.To4()
    fmt.Println(ip4)

    // ========== IP 掩码 ==========
    mask := net.CIDRMask(24, 32)  // 255.255.255.0

    // ========== IP 网络 ==========
    _, network, _ := net.ParseCIDR("192.168.1.0/24")
    fmt.Println(network)

    // 检查 IP 是否在网络中
    network.Contains(net.ParseIP("192.168.1.100"))  // true

    // ========== 端口服务 ==========
    net.LookupPort("tcp", "http")   // 80
    net.LookupPort("tcp", "https")  // 443

    // ========== DNS 查询 ==========
    // 查询 A 记录
    addrs, _ := net.LookupHost("google.com")
    fmt.Println(addrs)

    // 查询 IP 的域名
    names, _ := net.LookupAddr("8.8.8.8")
    fmt.Println(names)

    // 查询 MX 记录
    mx, _ := net.LookupMX("gmail.com")
    fmt.Println(mx)

    // 查询 NS 记录
    ns, _ := net.LookupNS("google.com")
    fmt.Println(ns)

    // 查询 TXT 记录
    txt, _ := net.LookupTXT("google.com")
    fmt.Println(txt)

    // ========== TCP 连接 ==========
    conn, err := net.Dial("tcp", "google.com:80")
    if err != nil {
        fmt.Println("连接失败:", err)
        return
    }
    defer conn.Close()

    // 发送数据
    conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))

    // 接收数据
    buf := make([]byte, 1024)
    n, _ := conn.Read(buf)
    fmt.Println(string(buf[:n]))

    // ========== UDP 连接 ==========
    udpConn, _ := net.Dial("udp", "8.8.8.8:53")
    defer udpConn.Close()

    // ========== 监听 TCP 端口 ==========
    listener, _ := net.Listen("tcp", ":8080")
    defer listener.Close()

    fmt.Println("监听端口 8080")

    for {
        conn, _ := listener.Accept()
        go handleConn(conn)
    }
}

func handleConn(conn net.Conn) {
    defer conn.Close()
    // 处理连接
}
```

### http 客户端

```go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

func main() {
    // ========== 简单 GET 请求 ==========
    resp, err := http.Get("https://api.example.com/data")
    if err != nil {
        fmt.Println("请求失败:", err)
        return
    }
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)
    fmt.Println(resp.StatusCode, string(body))

    // ========== 简单 POST 请求 ==========
    resp, _ = http.Post(
        "https://api.example.com/users",
        "application/json",
        bytes.NewBuffer([]byte(`{"name":"Alice"}`)),
    )
    defer resp.Body.Close()

    // ========== 创建请求 ==========
    req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)

    // 添加请求头
    req.Header.Set("Authorization", "Bearer token")
    req.Header.Set("Content-Type", "application/json")

    // 添加 Query 参数
    q := req.URL.Query()
    q.Add("page", "1")
    q.Add("limit", "10")
    req.URL.RawQuery = q.Encode()

    // ========== 自定义客户端 ==========
    client := &http.Client{
        Timeout: 30 * time.Second,
        // 可配置 Transport
    }

    resp, _ = client.Do(req)
    defer resp.Body.Close()

    // ========== 带超时的请求 ==========
    client2 := &http.Client{
        Timeout: 10 * time.Second,
    }

    resp, err = client2.Get("https://api.example.com/slow")
    if err != nil {
        fmt.Println("超时:", err)
    }

    // ========== POST JSON ==========
    type User struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }

    user := User{Name: "Alice", Age: 25}
    body, _ := json.Marshal(user)

    req, _ = http.NewRequest("POST", "https://api.example.com/users", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")

    resp, _ = client.Do(req)
    defer resp.Body.Close()

    // ========== 上传文件 ==========
    // 使用 multipart 表单
    // 见下面完整示例

    // ========== 下载文件 ==========
    resp, _ = http.Get("https://example.com/file.zip")
    defer resp.Body.Close()

    file, _ := os.Create("file.zip")
    defer file.Close()

    io.Copy(file, resp.Body)
}
```

### http 服务器

```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
)

// ========== 基础 HTTP 服务器 ==========
func basicServer() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
    })

    http.ListenAndServe(":8080", nil)
}

// ========== Handler 接口 ==========
// type Handler interface {
//     ServeHTTP(w ResponseWriter, r *Request)
// }

type MyHandler struct{}

func (h *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Custom handler")
}

func customHandler() {
    http.Handle("/custom", &MyHandler{})
    http.ListenAndServe(":8080", nil)
}

// ========== 路由处理 ==========
func routerHandler() {
    http.HandleFunc("/api/users", handleUsers)
    http.HandleFunc("/api/users/", handleUserByID)
    http.HandleFunc("/health", handleHealth)

    http.ListenAndServe(":8080", nil)
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        // 获取用户列表
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode([]string{"Alice", "Bob"})
    case http.MethodPost:
        // 创建用户
        w.WriteHeader(http.StatusCreated)
        fmt.Fprint(w, "User created")
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleUserByID(w http.ResponseWriter, r *http.Request) {
    // /api/users/123
    id := r.URL.Path[len("/api/users/"):]
    fmt.Fprintf(w, "User ID: %s", id)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ========== 请求处理 ==========
func requestDetails(w http.ResponseWriter, r *http.Request) {
    // 请求方法
    method := r.Method

    // 请求 URL
    url := r.URL.String()
    path := r.URL.Path
    query := r.URL.RawQuery

    // 请求头
    contentType := r.Header.Get("Content-Type")
    auth := r.Header.Get("Authorization")

    // Query 参数
    page := r.URL.Query().Get("page")
    allParams := r.URL.Query()

    // 读取请求体
    var body []byte
    if r.Body != nil {
        body, _ = io.ReadAll(r.Body)
        defer r.Body.Close()
    }

    // 解析表单
    r.ParseForm()
    formValue := r.FormValue("field")

    // 获取客户端 IP
    clientIP := r.RemoteAddr

    fmt.Fprintf(w, "Method: %s, Path: %s, Query: %s", method, path, query)
}

// ========== 响应处理 ==========
func responseExample(w http.ResponseWriter, r *http.Request) {
    // 设置响应头
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Custom-Header", "value")

    // 设置状态码
    w.WriteHeader(http.StatusOK)

    // 写入响应体
    json.NewEncoder(w).Encode(map[string]string{"message": "success"})
}

// ========== 静态文件服务 ==========
func staticFileServer() {
    // 服务当前目录
    http.Handle("/", http.FileServer(http.Dir(".")))

    // 服务特定路径
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))

    http.ListenAndServe(":8080", nil)
}

// ========== 中间件 ==========
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        fmt.Printf("%s %s %s\n", r.Method, r.URL.Path, time.Since(start))
    })
}

func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token != "secret" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func withMiddleware() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/", apiHandler)

    // 应用中间件
    handler := loggingMiddleware(authMiddleware(mux))
    http.ListenAndServe(":8080", handler)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "API response")
}

// ========== Graceful Shutdown ==========
func gracefulShutdown() {
    server := &http.Server{
        Addr:    ":8080",
        Handler: http.DefaultServeMux,
    }

    // 在 goroutine 中启动服务器
    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Println("Server error:", err)
        }
    }()

    // 等待中断信号
    // quit := make(chan os.Signal, 1)
    // signal.Notify(quit, os.Interrupt)
    // <-quit

    // 优雅关闭
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        fmt.Println("Shutdown error:", err)
    }
}

func main() {
    fmt.Println("启动服务器...")
    routerHandler()
}
```

---

## 4.3 数据库操作

### database/sql 接口

```go
package main

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // ========== 连接数据库 ==========
    db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/dbname")
    if err != nil {
        fmt.Println("连接失败:", err)
        return
    }
    defer db.Close()

    // 验证连接
    if err := db.Ping(); err != nil {
        fmt.Println("Ping 失败:", err)
        return
    }

    fmt.Println("连接成功")
}
```

### 增删改查操作

```go
package main

import (
    "database/sql"
    "fmt"
    "time"
)

type User struct {
    ID        int
    Name      string
    Email     string
    CreatedAt time.Time
}

// ========== 插入 ==========
func insertUser(db *sql.DB) (int64, error) {
    result, err := db.Exec(
        "INSERT INTO users (name, email, created_at) VALUES (?, ?, ?)",
        "Alice", "alice@example.com", time.Now(),
    )
    if err != nil {
        return 0, err
    }

    id, _ := result.LastInsertId()
    return id, nil
}

// ========== 查询单行 ==========
func getUser(db *sql.DB, id int) (*User, error) {
    var user User
    err := db.QueryRow(
        "SELECT id, name, email, created_at FROM users WHERE id = ?",
        id,
    ).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)

    if err == sql.ErrNoRows {
        return nil, nil  // 未找到
    }
    if err != nil {
        return nil, err
    }

    return &user, nil
}

// ========== 查询多行 ==========
func getAllUsers(db *sql.DB) ([]*User, error) {
    rows, err := db.Query("SELECT id, name, email, created_at FROM users")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []*User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
            return nil, err
        }
        users = append(users, &user)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}

// ========== 更新 ==========
func updateUser(db *sql.DB, id int, email string) error {
    result, err := db.Exec(
        "UPDATE users SET email = ? WHERE id = ?",
        email, id,
    )
    if err != nil {
        return err
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("用户不存在")
    }

    return nil
}

// ========== 删除 ==========
func deleteUser(db *sql.DB, id int) error {
    result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
    if err != nil {
        return err
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return fmt.Errorf("用户不存在")
    }

    return nil
}

// ========== 批量插入 (事务) ==========
func batchInsert(db *sql.DB, users []User) error {
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    stmt, err := tx.Prepare("INSERT INTO users (name, email) VALUES (?, ?)")
    if err != nil {
        return err
    }
    defer stmt.Close()

    for _, u := range users {
        if _, err := stmt.Exec(u.Name, u.Email); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

### 连接池配置

```go
package main

import "database/sql"

func configureConnectionPool(db *sql.DB) {
    // 最大空闲连接数
    db.SetMaxIdleConns(10)

    // 最大打开连接数
    db.SetMaxOpenConns(100)

    // 连接最大生命周期
    db.SetConnMaxLifetime(time.Hour)

    // 连接最大空闲时间 (Go 1.15+)
    db.SetConnMaxIdleTime(time.Minute)
}
```

### 事务处理

```go
package main

import (
    "database/sql"
    "fmt"
)

func transferMoney(db *sql.DB, fromID, toID int, amount float64) error {
    // 开启事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }

    // 确保事务回滚或提交
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    // 扣款
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance - ? WHERE id = ?",
        amount, fromID,
    )
    if err != nil {
        tx.Rollback()
        return err
    }

    // 收款
    _, err = tx.Exec(
        "UPDATE accounts SET balance = balance + ? WHERE id = ?",
        amount, toID,
    )
    if err != nil {
        tx.Rollback()
        return err
    }

    // 提交事务
    if err := tx.Commit(); err != nil {
        return err
    }

    return nil
}

// ========== 事务选项 ==========
func transactionWithOptions(db *sql.DB) error {
    // 设置隔离级别
    tx, err := db.BeginTx(context.Background(), &sql.TxOptions{
        Isolation: sql.LevelReadCommitted,
        ReadOnly:  false,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // ... 操作 ...

    return tx.Commit()
}
```

### ORM 简介

```go
// ========== GORM 示例 ==========
/*
import (
    "gorm.io/gorm"
    "gorm.io/driver/mysql"
)

// 定义模型
type User struct {
    gorm.Model
    Name  string
    Email string `gorm:"uniqueIndex"`
    Posts []Post
}

type Post struct {
    gorm.Model
    Title   string
    Content string
    UserID  uint
}

// 连接数据库
func connectGORM() (*gorm.DB, error) {
    dsn := "user:pass@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// 自动迁移
db.AutoMigrate(&User{}, &Post{})

// 创建
db.Create(&User{Name: "Alice", Email: "alice@example.com"})

// 查询
var user User
db.First(&user, 1)
db.Where("name = ?", "Alice").First(&user)
db.Find(&users)

// 更新
db.Model(&user).Update("Name", "Bob")
db.Model(&user).Updates(User{Name: "Charlie"})

// 删除
db.Delete(&user)

// 关联查询
db.Preload("Posts").First(&user, 1)
*/

// ========== sqlx 示例 ==========
/*
import "github.com/jmoiron/sqlx"

// sqlx 是 database/sql 的扩展

// 连接
db, _ := sqlx.Connect("mysql", dsn)

// 结构化查询
var users []User
db.Select(&users, "SELECT * FROM users")

// 命名查询
db.NamedExec("INSERT INTO users (name, email) VALUES (:name, :email)", user)

// Get (单行)
var user User
db.Get(&user, "SELECT * FROM users WHERE id = ?", 1)
*/
```

---

## 4.4 测试与基准测试

### testing 包

```go
package main

import "testing"

// ========== 测试函数 ==========
// 测试文件以 _test.go 结尾
// 测试函数以 Test 开头

func TestAdd(t *testing.T) {
    // 测试用例
    result := add(2, 3)
    expected := 5

    if result != expected {
        t.Errorf("add(2, 3) = %d; expected %d", result, expected)
    }
}

// ========== 表格驱动测试 ==========
func TestAddTable(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"正数", 2, 3, 5},
        {"负数", -1, -1, -2},
        {"零", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("add(%d, %d) = %d; expected %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}

// ========== 辅助函数 ==========
func add(a, b int) int {
    return a + b
}

// ========== 测试辅助方法 ==========
func TestHelper(t *testing.T) {
    result := add(1, 2)

    if result != 3 {
        t.Helper()  // 标记为辅助函数，错误指向调用处
        t.Errorf("add(1, 2) = %d; expected 3", result)
    }
}

// ========== 并行测试 ==========
func TestParallel(t *testing.T) {
    t.Parallel()  // 标记为可并行执行
    // 测试代码
}

// ========== Setup 和 Teardown ==========
func TestWithSetup(t *testing.T) {
    // Setup
    setup()
    defer teardown()

    // 测试代码
}

func setup()   { /* 准备 */ }
func teardown() { /* 清理 */ }
```

### 基准测试

```go
package main

import "testing"

// ========== 基准测试函数 ==========
// 以 Benchmark 开头

func BenchmarkAdd(b *testing.B) {
    // b.N 会自动调整以达到稳定的测量
    for i := 0; i < b.N; i++ {
        add(2, 3)
    }
}

// ========== 基准测试选项 ==========
func BenchmarkAddOptimized(b *testing.B) {
    b.ResetTimer()  // 重置计时器 (排除准备时间)

    for i := 0; i < b.N; i++ {
        add(2, 3)
    }

    b.ReportAllocs()  // 报告内存分配
}

// ========== 并行基准测试 ==========
func BenchmarkAddParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            add(2, 3)
        }
    })
}

// ========== 运行基准测试 ==========
// go test -bench=.
// go test -bench=. -benchmem  # 包含内存分配
// go test -bench=. -cpuprofile=cpu.out  # CPU 分析
```

### 示例测试

```go
package main

import (
    "fmt"
)

// ========== 示例测试 ==========
// 以 Example 开头
// 输出注释用于验证

func ExampleAdd() {
    result := Add(2, 3)
    fmt.Println(result)
    // Output: 5
}

func ExampleGreet() {
    name := "World"
    result := Greet(name)
    fmt.Println(result)
    // Output: Hello, World!
}

func Add(a, b int) int {
    return a + b
}

func Greet(name string) string {
    return "Hello, " + name + "!"
}
```

### Test Main

```go
package main

import (
    "os"
    "testing"
)

// ========== TestMain ==========
// 用于测试前后的全局设置

func TestMain(m *testing.M) {
    // Setup (测试前)
    setupTestDatabase()
    setupTestServer()

    // 运行测试
    code := m.Run()

    // Teardown (测试后)
    teardownTestDatabase()
    teardownTestServer()

    // 退出
    os.Exit(code)
}

func setupTestDatabase()   { /* ... */ }
func teardownTestDatabase() { /* ... */ }
func setupTestServer()      { /* ... */ }
func teardownTestServer()   { /* ... */ }
```

### 测试覆盖率

```bash
# 查看覆盖率
go test -cover

# 生成覆盖率报告
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# 查看每个函数的覆盖率
go test -coverprofile=coverage.out
go tool cover -func=coverage.out

# 生成带覆盖率标注的源码
go tool cover -html=coverage.out -o coverage.html
```

---

## 4.5 日志与调试

### log 包基础

```go
package main

import (
    "log"
    "os"
)

func main() {
    // ========== 标准日志 ==========
    log.Println("普通日志")
    log.Printf("格式化日志：%s", "info")

    // ========== 日志级别 ==========
    // 标准库没有级别，需要自定义

    // ========== 输出到文件 ==========
    file, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    defer file.Close()

    log.SetOutput(file)
    log.Println("写入文件")

    // ========== 自定义前缀 ==========
    log.SetPrefix("[APP] ")

    // ========== 添加日期时间 ==========
    log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
```

### 结构化日志

```go
// ========== 使用 zap ==========
/*
import "go.uber.org/zap"

logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("用户登录",
    zap.String("username", "alice"),
    zap.Int("user_id", 123),
)

// 快速使用
zap.L().Info("message")
*/

// ========== 使用 zerolog ==========
/*
import "github.com/rs/zerolog"

logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

logger.Info().
    Str("username", "alice").
    Int("user_id", 123).
    Msg("用户登录")
*/
```

### pprof 性能分析

```go
package main

import (
    "net/http"
    _ "net/http/pprof"
    "runtime"
    "runtime/pprof"
    "os"
)

// ========== HTTP pprof ==========
func main() {
    // 访问 /debug/pprof/ 查看
    // 使用 go tool pprof http://localhost:6060/debug/pprof/profile
    http.ListenAndServe("localhost:6060", nil)
}

// ========== CPU 分析 ==========
func cpuProfile() {
    f, _ := os.Create("cpu.prof")
    defer f.Close()

    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // ... 运行代码 ...
}

// ========== 内存分析 ==========
func memProfile() {
    f, _ := os.Create("mem.prof")
    defer f.Close()

    runtime.GC()
    pprof.WriteHeapProfile(f)
}

// ========== 阻塞分析 ==========
func blockProfile() {
    runtime.SetBlockProfileRate(1)
    f, _ := os.Create("block.prof")
    defer f.Close()
    pprof.Lookup("block").WriteTo(f, 0)
}

// ========== 使用分析文件 ==========
// go tool pprof cpu.prof
// go tool pprof mem.prof
// go tool pprof -http=:8080 cpu.prof
```

---

## 第四部分完

接下来可以继续学习第五部分：Web 开发
