# 1.5 Map

## 创建与初始化

```go
package main

import "fmt"

func main() {
    // make 创建
    m1 := make(map[string]int)

    // 字面量
    m2 := map[string]int{"a": 1, "b": 2}

    // 空 map
    m3 := map[string]int{}

    // 指定初始容量
    m4 := make(map[string]int, 100)

    // nil map (不能直接赋值)
    var m5 map[string]int
}
```

---

## 增删改查

```go
package main

import "fmt"

func main() {
    scores := make(map[string]int)

    // 增/改
    scores["Alice"] = 95
    scores["Bob"] = 87
    scores["Alice"] = 98  // 修改

    // 查
    fmt.Println(scores["Alice"])  // 98
    fmt.Println(scores["Carol"])  // 0 (零值)

    // 检查键是否存在
    if val, ok := scores["Bob"]; ok {
        fmt.Println("Bob:", val)
    }

    // 删除
    delete(scores, "Bob")

    // 获取长度
    fmt.Println(len(scores))
}
```

---

## 遍历

```go
package main

import (
    "fmt"
    "sort"
)

func main() {
    m := map[string]int{"a": 1, "b": 2, "c": 3}

    // 遍历 (无序)
    for k, v := range m {
        fmt.Printf("%s: %d\n", k, v)
    }

    // 有序遍历
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, m[k])
    }
}
```

---

## Map 与切片组合

```go
package main

import "fmt"

func main() {
    // Map 的 value 是切片
    m1 := map[string][]int{
        "scores": {1, 2, 3},
    }
    m1["scores"] = append(m1["scores"], 4)

    // 统计词频
    words := []string{"a", "b", "a", "c"}
    freq := make(map[string]int)
    for _, w := range words {
        freq[w]++
    }
    fmt.Println(freq)  // map[a:2 b:1 c:1]

    // 分组
    type Person struct {
        Name string
        Age  int
    }
    byAge := make(map[int][]Person)
    // byAge[20] = append(byAge[20], person)
}
```

---

## sync.Map (并发安全)

```go
package main

import (
    "fmt"
    "sync"
)

func main() {
    var m sync.Map

    m.Store("a", 1)
    m.Store("b", 2)

    val, ok := m.Load("a")
    fmt.Println(val, ok)

    m.Delete("a")

    m.Range(func(key, value interface{}) bool {
        fmt.Printf("%v: %v\n", key, value)
        return true
    })
}
```
