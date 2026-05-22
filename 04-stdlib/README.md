# Part 04: 标准库精讲

本部分深入讲解Go语言标准库的核心包，帮助开发者掌握标准库的使用方法和最佳实践。

## 学习目标

完成本部分学习后，你将能够：

- 熟练使用常用标准库包
- 掌握文件IO操作
- 理解网络编程基础
- 编写单元测试和基准测试
- 处理序列化和反序列化

## 重要说明

**本部分专注于Go官方标准库**，不包括第三方库。

第三方库内容：
- 配置管理（Viper）→ Part 06 工程实践
- ORM框架（GORM）→ Part 05 Web开发
- Web框架（Gin/Echo）→ Part 05 Web开发

## 章节结构

### 01-常用标准库

**学习时间**: 2-3天

**核心内容**:
- fmt - 格式化I/O
- strings - 字符串处理
- strconv - 类型转换
- sort - 排序算法
- math - 数学函数

**实践要求**:
- [ ] 使用fmt进行格式化
- [ ] 掌握strings包函数
- [ ] 实现自定义排序

### 02-网络编程

**学习时间**: 2-3天

**核心内容**:
- net包基础
- TCP/UDP编程
- http客户端（net/http）
- http服务器
- URL解析（net/url）

**实践要求**:
- [ ] 实现TCP客户端/服务器
- [ ] 编写HTTP客户端
- [ ] 构建简单HTTP服务器

### 03-数据库操作

**学习时间**: 1-2天

**核心内容**:
- database/sql接口
- SQL驱动使用（mysql、postgres）
- 连接池配置
- 事务处理
- 预编译语句

**实践要求**:
- [ ] 使用database/sql连接数据库
- [ ] 执行查询和更新
- [ ] 处理事务

### 04-测试与基准测试

**学习时间**: 2-3天

**核心内容**:
- testing包
- 单元测试编写
- 表格驱动测试
- 测试覆盖率
- 基准测试（Benchmark）
- 示例测试（Example）

**实践要求**:
- [ ] 编写单元测试
- [ ] 实现表格驱动测试
- [ ] 编写基准测试
- [ ] 生成覆盖率报告

### 05-日志与调试

**学习时间**: 1-2天

**核心内容**:
- log包基础
- 结构化日志
- 日志级别
- pprof性能分析
- Delve调试器

**实践要求**:
- [ ] 使用log包记录日志
- [ ] 使用pprof分析性能
- [ ] 使用Delve调试程序

### 06-文件与IO操作

**学习时间**: 2-3天

**核心内容**:
- os包文件系统操作
- io.Reader与io.Writer
- io.Copy与io.Pipe
- bufio带缓冲IO
- 文件权限与模式
- 路径处理（filepath）

**实践要求**:
- [ ] 读写文件
- [ ] 实现io.Reader/Writer
- [ ] 处理文件路径

### 07-正则表达式

**学习时间**: 1-2天

**核心内容**:
- regexp包基础
- 正则表达式语法
- 匹配和查找
- 替换和分割

**实践要求**:
- [ ] 编写正则表达式
- [ ] 实现文本匹配
- [ ] 处理文本替换

### 08-序列化

**学习时间**: 2-3天

**核心内容**:
- JSON编解码（encoding/json）
- JSON标签与选项
- 自定义JSON marshal/unmarshal
- XML编解码（encoding/xml）
- Gob序列化

**实践要求**:
- [ ] 编解码JSON
- [ ] 自定义序列化
- [ ] 处理XML数据

## 前置知识

- 完成**Part 01: Go语言基础**
- 完成**Part 02: 核心特性**
- 理解接口和错误处理

## 核心概念

### 标准库 vs 第三方库

**标准库**：
- 由Go官方维护
- 包含在Go安装包中
- 质量高、稳定
- 示例：`fmt`、`net/http`、`database/sql`

**第三方库**：
- 社区开发维护
- 需要单独安装
- 功能更丰富
- 示例：`gin`、`gorm`、`viper`

### IO操作模式

```go
// io.Reader接口
type Reader interface {
    Read(p []byte) (n int, err error)
}

// io.Writer接口
type Writer interface {
    Write(p []byte) (n int, err error)
}

// 使用示例
func copyFile(dst, src string) error {
    r, err := os.Open(src)
    if err != nil {
        return err
    }
    defer r.Close()
    
    w, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer w.Close()
    
    _, err = io.Copy(w, r)
    return err
}
```

### 测试驱动开发

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 1, 2, 3},
        {"negative", -1, -2, -3},
        {"zero", 0, 0, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", 
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}

func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(1, 2)
    }
}
```

## 检查清单

完成本部分后，验证你的掌握程度：

- [ ] 熟练使用fmt、strings、strconv
- [ ] 实现TCP/HTTP网络编程
- [ ] 使用database/sql操作数据库
- [ ] 编写单元测试和基准测试
- [ ] 处理文件IO操作
- [ ] 使用正则表达式
- [ ] 编解码JSON/XML

## 学习建议

1. **优先掌握标准库**：标准库是基础，第三方库往往基于标准库构建
2. **阅读源码**：标准库源码是最好的学习材料
3. **实践为主**：每个包都要写代码验证
4. **查看文档**：使用`go doc`查看包文档

## 常用命令

```bash
# 查看包文档
go doc fmt
go doc fmt.Printf

# 运行测试
go test ./...

# 运行基准测试
go test -bench=.

# 查看覆盖率
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# 性能分析
go test -cpuprofile=cpu.out
go tool pprof cpu.out
```

## 延伸阅读

- [Go标准库文档](https://golang.org/pkg/)
- [Go by Example](https://gobyexample.com/)
- [Go标准库源码](https://github.com/golang/go/tree/master/src)

## 下一步学习

完成Part 04后，建议继续学习：

- **Part 05: Web开发** - Web框架、中间件、微服务
- **Part 06: 工程实践** - 项目结构、配置管理、测试策略

---

## 总结

Part 04专注Go标准库，帮助开发者打下坚实基础：

**重点内容**:
1. ✅ 常用标准库包
2. ✅ 文件IO和网络编程
3. ✅ 测试框架
4. ✅ 数据序列化

**学习时间**: 约2-3周（每天2-3小时）

**关键点**: 标准库是Go的基础，务必熟练掌握！