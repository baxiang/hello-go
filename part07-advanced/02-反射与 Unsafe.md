# 7.2 反射与 Unsafe

## reflect 包基础

```go
package main

import (
    "fmt"
    "reflect"
)

type Person struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
}

func main() {
    p := Person{Name: "Alice", Age: 25}

    // 获取类型信息
    t := reflect.TypeOf(p)
    fmt.Println("类型:", t)           // main.Person
    fmt.Println("种类:", t.Kind())    // struct
    fmt.Println("字段数:", t.NumField())  // 2

    // 遍历字段
    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        fmt.Printf("字段：%s (%s) tag=%s\n",
            field.Name, field.Type, field.Tag)
    }

    // 获取值信息
    v := reflect.ValueOf(p)
    for i := 0; i < v.NumField(); i++ {
        fmt.Printf("%s = %v\n", v.Type().Field(i).Name, v.Field(i).Interface())
    }

    // 修改值 (需要指针)
    vp := reflect.ValueOf(&p).Elem()
    vp.FieldByName("Name").SetString("Bob")
    fmt.Println(p)  // {Bob 25}

    // 创建实例
    t2 := reflect.TypeOf(Person{})
    v2 := reflect.New(t2)  // 返回指针
    newInstance := v2.Interface().(*Person)
}
```

---

## 反射的应用场景

```go
// 1. 结构体验证
func validate(v interface{}) error {
    val := reflect.ValueOf(v)
    typ := val.Type()

    for i := 0; i < val.NumField(); i++ {
        field := val.Field(i)
        fieldType := typ.Field(i)

        validateTag := fieldType.Tag.Get("validate")
        if validateTag == "required" && field.IsZero() {
            return fmt.Errorf("%s 是必填项", fieldType.Name)
        }
    }
    return nil
}

// 2. 深度拷贝
func deepCopy(dst, src interface{}) error {
    dstVal := reflect.ValueOf(dst)
    srcVal := reflect.ValueOf(src)

    if dstVal.Kind() != reflect.Ptr || dstVal.IsNil() {
        return fmt.Errorf("dst 必须是非空指针")
    }

    dstVal.Elem().Set(srcVal)
    return nil
}

// 3. 动态调用
func invokeHandler(handler interface{}, input string) string {
    hv := reflect.ValueOf(handler)
    args := []reflect.Value{reflect.ValueOf(input)}
    results := hv.Call(args)
    return results[0].String()
}
```

---

## unsafe 包

```go
package main

import (
    "fmt"
    "unsafe"
)

func main() {
    // unsafe.Pointer
    x := 42
    p := &x
    ptr := unsafe.Pointer(p)

    // 转换为其他类型指针
    intPtr := (*int)(ptr)
    fmt.Println(*intPtr)  // 42

    // uintptr
    addr := uintptr(ptr)
    fmt.Printf("地址：0x%x\n", addr)

    // 访问结构体字段
    type Person struct {
        Name string
        Age  int
    }

    p2 := Person{Name: "Alice", Age: 25}
    p2Ptr := unsafe.Pointer(&p2)

    // Name 字段偏移量 = 0
    namePtr := (*string)(p2Ptr)
    fmt.Println(*namePtr)  // Alice

    // 字节转换 (零拷贝)
    s := "hello"
    b := unsafe.Slice(
        (*byte)(unsafe.Pointer(unsafe.StringData(s))),
        len(s),
    )
    fmt.Println(string(b))  // hello
}

// 注意：unsafe 破坏了类型安全，应谨慎使用
```
