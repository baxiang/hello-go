# GORM 数据库操作

## 简介

GORM 是 Go 的 ORM 库，特点：
- 全功能 ORM (关联、事务、钩子等)
- 自动迁移
- 支持 MySQL、PostgreSQL、SQLite、SQL Server
- 灵活的插件系统

## 安装

```bash
go get -u gorm.io/gorm
go get -u gorm.io/driver/mysql    # MySQL
go get -u gorm.io/driver/postgres # PostgreSQL
go get -u gorm.io/driver/sqlite   # SQLite
```

## 快速开始

### 连接数据库

```go
package main

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func main() {
    dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        // 禁用外键约束
        DisableForeignKeyConstraintWhenMigrating: true,
        // 日志级别
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        panic(err)
    }

    // 获取底层 *sql.DB
    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
}
```

### 定义模型

```go
package model

import (
    "time"
    "gorm.io/gorm"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey"`
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`  // 软删除

    Name     string `gorm:"size:255;not null"`
    Email    string `gorm:"size:255;uniqueIndex"`
    Age      int    `gorm:"comment:用户年龄"`
    Birthday *time.Time

    // 关联
    Profile  *Profile
    Orders   []Order
    Roles    []Role `gorm:"many2many:user_roles;"`
}

// Profile 用户资料
type Profile struct {
    ID     uint   `gorm:"primaryKey"`
    UserID uint   `gorm:"uniqueIndex;not null"`
    Bio    string `gorm:"type:text"`
    Avatar string
}

// Order 订单
type Order struct {
    ID        uint   `gorm:"primaryKey"`
    UserID    uint   `gorm:"index"`
    OrderNo   string `gorm:"size:64;uniqueIndex"`
    Amount    int64  `gorm:"comment:金额 (分)"`
    Status    int    `gorm:"default:0"`
    User      User   `gorm:"foreignKey:UserID"`
    Items     []OrderItem
}

// OrderItem 订单项
type OrderItem struct {
    ID      uint   `gorm:"primaryKey"`
    OrderID uint   `gorm:"index"`
    Product string
    Quantity int
    Price   int64
}

// Role 角色
type Role struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"size:64;uniqueIndex"`
    Users []User `gorm:"many2many:user_roles;"`
}
```

### 自动迁移

```go
func migrate(db *gorm.DB) error {
    // 自动迁移所有模型
    return db.AutoMigrate(
        &User{},
        &Profile{},
        &Order{},
        &OrderItem{},
        &Role{},
    )
}

// 自定义迁移
func migrateWithCustom(db *gorm.DB) error {
    err := db.AutoMigrate(&User{})
    if err != nil {
        return err
    }

    // 创建索引
    db.Exec("CREATE INDEX idx_users_email ON users(email)")

    // 添加外键
    db.Exec("ALTER TABLE profiles ADD CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id)")

    return nil
}
```

## CRUD 操作

### 创建

```go
// 单条创建
user := User{Name: "Alice", Email: "alice@example.com", Age: 25}
result := db.Create(&user)
fmt.Println(user.ID)  // 自动生成的 ID

// 批量创建
users := []User{
    {Name: "Bob", Email: "bob@example.com"},
    {Name: "Charlie", Email: "charlie@example.com"},
}
db.Create(&users)

// 带关联创建
user := User{
    Name:  "David",
    Email: "david@example.com",
    Profile: &Profile{
        Bio: "Developer",
    },
    Orders: []Order{
        {OrderNo: "ORD001", Amount: 10000},
    },
}
db.Create(&user)  // 级联创建关联数据

// 使用 Map 创建
db.Model(&User{}).Create(map[string]interface{}{
    "Name":  "Eve",
    "Email": "eve@example.com",
})

// 选择字段创建
db.Select("Name", "Email").Create(&User{
    Name:  "Frank",
    Email: "frank@example.com",
    Age:   30,  // 会被忽略
})

// 忽略字段创建
db.Omit("Age").Create(&user)

// 创建或更新 (Upsert)
user := User{ID: 1, Name: "Updated"}
db.Clauses(clause.OnConflict{
    UpdateAll: true,  // 更新所有字段
}).Create(&user)
```

### 查询

```go
var user User
var users []User

// 根据主键查询
db.First(&user, 1)           // WHERE id = 1
db.Take(&user, "id = ?", 1)  // 第一条

db.Find(&users, []int{1, 2, 3})  // WHERE id IN (1,2,3)

// 条件查询
db.Where("name = ?", "Alice").First(&user)
db.Where("name = ? AND age > ?", "Alice", 20).Find(&users)
db.Where("age IN ?", []int{20, 30}).Find(&users)
db.Where("name LIKE ?", "%al%").Find(&users)

// Map 条件
db.Where(map[string]interface{}{"name": "Alice", "age": 25}).Find(&users)

// 结构体条件 (仅非零值)
db.Where(&User{Name: "Alice"}).Find(&users)

// OR 条件
db.Where("role = ?", "admin").Or("role = ?", "super_admin").Find(&users)

// NOT 条件
db.Not("name = ?", "Alice").Find(&users)
db.Where(db.Not("id IN ?", []int{1, 2, 3})).Find(&users)

// 选择字段
db.Select("name", "email").Find(&users)
db.Select([]string{"name", "email"}).Find(&users)

// 排除字段
db.Omit("password").Find(&users)

// 排序
db.Order("created_at DESC").Find(&users)
db.Order("age ASC, name DESC").Find(&users)

// 限制和偏移
db.Limit(10).Offset(20).Find(&users)  // LIMIT 10 OFFSET 20

// 计数
var count int64
db.Model(&User{}).Where("age > ?", 20).Count(&count)

// 聚合
db.Model(&Order{}).Select("SUM(amount)").Scan(&total)
db.Model(&Order{}).Select("AVG(amount)").Scan(&avg)
db.Model(&Order{}).Select("MAX(amount)").Scan(&max)

// 分组
db.Model(&Order{}).Select("user_id, SUM(amount)").Group("user_id").Find(&results)

// Having
db.Model(&Order{}).Select("user_id, COUNT(*) as cnt").
    Group("user_id").Having("COUNT(*) > ?", 5).Find(&results)

// 原生 SQL
type Result struct {
    ID   uint
    Name string
}
db.Raw("SELECT id, name FROM users WHERE age > ?", 20).Scan(&results)
db.Exec("UPDATE users SET age = ? WHERE id = ?", 30, 1)
```

### 更新

```go
var user User
db.First(&user, 1)

// 更新单个字段
db.Model(&user).Update("Name", "New Name")

// 更新多个字段
db.Model(&user).Updates(map[string]interface{}{
    "Name":  "New Name",
    "Email": "new@example.com",
})

// 使用结构体更新
db.Model(&user).Updates(User{Name: "New Name"})

// 选择字段更新
db.Model(&user).Select("Name", "Age").Updates(User{
    Name:  "New",
    Age:   30,
    Email: "unchanged@example.com",  // 被忽略
})

// 忽略字段更新
db.Model(&user).Omit("Email").Updates(user)

// 更新所有字段 (包括零值)
db.Model(&user).Select(clause.Associations).Updates(&user)

// 条件更新
db.Model(&User{}).Where("age < ?", 18).Update("status", "minor")

// 批量更新
db.Model(&User{}).Where("id IN ?", []int{1, 2, 3}).Update("status", "active")
```

### 删除

```go
var user User
db.First(&user, 1)

// 软删除 (需要 DeletedAt 字段)
db.Delete(&user)
db.Where("id = ?", 1).Delete(&User{})

// 硬删除 (需要 Remove Soft Delete)
db.Unscoped().Delete(&user)
db.Unscoped().Where("id = ?", 1).Delete(&User{})

// 批量删除
db.Where("age < ?", 18).Delete(&User{})

// 根据主键删除
db.Delete(&User{}, 1)
db.Delete(&User{}, []int{1, 2, 3})
```

## 关联查询

### 一对一

```go
// 定义
type User struct {
    ID      uint
    Name    string
    Profile *Profile  // has one
}

type Profile struct {
    ID     uint
    UserID uint
    Bio    string
}

// 查询
var user User
db.Preload("Profile").First(&user, 1)

// 条件预加载
db.Preload("Profile", "bio LIKE ?", "%developer%").First(&user)

// 嵌套预加载
db.Preload("Profile").Preload("Profile.Address").First(&user)

// 所有关联
db.Preload(clause.Associations).Find(&users)
```

### 一对多

```go
// 定义
type User struct {
    ID     uint
    Name   string
    Orders []Order  // has many
}

type Order struct {
    ID     uint
    UserID uint
    User   User   `gorm:"foreignKey:UserID"`
    Amount int64
}

// 查询
var user User
db.Preload("Orders").First(&user, 1)

// 条件预加载
db.Preload("Orders", "amount > ?", 10000).First(&user)

// 限制预加载数量
db.Preload(clause.Associations).Find(&users)
```

### 多对多

```go
// 定义
type User struct {
    ID    uint
    Name  string
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    ID    uint
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}

// 查询
var user User
db.Preload("Roles").First(&user, 1)

// 条件预加载
db.Preload("Roles", "name IN ?", []string{"admin", "user"}).First(&user)

// 关联查询
var roles []Role
db.Model(&user).Association("Roles").Find(&roles)

// 关联添加
db.Model(&user).Association("Roles").Append(&Role{ID: 1})

// 关联替换
db.Model(&user).Association("Roles").Replace([]Role{role1, role2})

// 关联删除
db.Model(&user).Association("Roles").Delete(&role1)

// 关联清空
db.Model(&user).Association("Roles").Clear()

// 关联计数
db.Model(&user).Association("Roles").Count()
```

## 事务

```go
// 基础事务
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

var user User
if err := tx.First(&user, 1).Error; err != nil {
    tx.Rollback()
    return
}

if err := tx.Model(&user).Update("age", 30).Error; err != nil {
    tx.Rollback()
    return
}

if err := tx.Create(&Order{UserID: user.ID, Amount: 100}).Error; err != nil {
    tx.Rollback()
    return
}

// 提交事务
if err := tx.Commit().Error; err != nil {
    return
}

// 事务闭包
err := db.Transaction(func(tx *gorm.DB) error {
    var user User
    if err := tx.First(&user, 1).Error; err != nil {
        return err
    }

    if err := tx.Model(&user).Update("age", 30).Error; err != nil {
        return err
    }

    return tx.Create(&Order{UserID: user.ID}).Error
})

// 带保存点的事务
db.Transaction(func(tx1 *gorm.DB) error {
    tx1.Create(&User{Name: "Alice"})

    // 保存点
    tx1.SavePoint("before_order")

    if err := tx1.Create(&Order{}).Error; err != nil {
        // 回滚到保存点
        tx1.RollbackTo("before_order")
    }

    return nil
})
```

## 钩子 (Hooks)

```go
type User struct {
    ID   uint
    Name string
}

// 创建前
func (u *User) BeforeCreate(tx *gorm.DB) error {
    // 加密密码
    hash, _ := bcrypt.GenerateFromPassword([]byte(u.Password), 10)
    u.Password = string(hash)

    // 设置默认值
    if u.Status == "" {
        u.Status = "active"
    }

    return nil
}

// 创建后
func (u *User) AfterCreate(tx *gorm.DB) error {
    // 发送欢迎邮件
    go sendWelcomeEmail(u.Email)
    return nil
}

// 更新前
func (u *User) BeforeUpdate(tx *gorm.DB) error {
    u.UpdatedAt = time.Now()
    return nil
}

// 删除前
func (u *User) BeforeDelete(tx *gorm.DB) error {
    // 检查是否可以删除
    var count int64
    tx.Model(&Order{}).Where("user_id = ?", u.ID).Count(&count)
    if count > 0 {
        return errors.New("用户有未完成的订单")
    }
    return nil
}

// 查找后
func (u *User) AfterFind(tx *gorm.DB) error {
    // 计算字段
    u.FullName = u.FirstName + " " + u.LastName
    return nil
}
```

## 高级查询

### 子查询

```go
// WHERE in 子查询
db.Where("amount > (?)", db.Table("orders").Select("AVG(amount)")).Find(&users)

// 条件子查询
db.Where("id IN (?)", db.Table("orders").Select("user_id").Where("amount > ?", 100)).Find(&users)

// 带别名的子查询
db.Select("u.*").Table("users as u").
    Where("u.id IN (?)", db.Table("orders").Select("user_id")).Find(&results)
```

### Join 查询

```go
type Result struct {
    UserName  string
    OrderNo   string
    Amount    int64
}

db.Table("users").
    Select("users.name as user_name, orders.order_no, orders.amount").
    Joins("LEFT JOIN orders ON orders.user_id = users.id").
    Where("orders.amount > ?", 100).
    Scan(&results)

// 使用条件 Join
db.Joins("JOIN orders ON orders.user_id = users.id AND orders.status = ?", 1).Find(&users)

// Join 预加载
db.Joins("Profile").Find(&users)
```

### 自定义查询

```go
// Scopes 定义可复用的查询条件
func AmountGreaterThan100(db *gorm.DB) *gorm.DB {
    return db.Where("amount > ?", 100)
}

func StatusActive(db *gorm.DB) *gorm.DB {
    return db.Where("status = ?", "active")
}

// 使用
db.Scopes(AmountGreaterThan100, StatusActive).Find(&orders)

// 带参数的 Scope
func WithName(name string) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        return db.Where("name LIKE ?", "%"+name+"%")
    }
}

db.Scopes(WithName("alice")).Find(&users)
```

## 性能优化

```go
// N+1 问题 - 使用 Preload
// 不好
var users []User
db.Find(&users)
for i := range users {
    db.Find(&users[i].Orders)  // N 次查询
}

// 好
db.Preload("Orders").Find(&users)  // 2 次查询

// 选择需要的字段
db.Select("id", "name", "email").Find(&users)

// 批量操作
users := []User{{Name: "A"}, {Name: "B"}, {Name: "C"}}
db.Create(&users)  // 一次插入

// 使用 FindInBatches 处理大数据
var users []User
db.FindInBatches(&users, 100, func(tx *gorm.DB, batch int) error {
    for _, u := range users {
        // 处理
    }
    return nil
})

// 禁用钩子提升性能
db.Session(&gorm.Session{SkipHooks: true}).Create(&user)

// 准备语句
db = db.Session(&gorm.Session{PrepareStmt: true})

// 缓存 (使用插件或自定义)
```

## GORM 检查清单

```
[ ] 使用指针接收者定义钩子
[ ] 为软删除添加 DeletedAt 字段
[ ] 使用 Preload 避免 N+1
[ ] 事务使用闭包方式
[ ] 批量操作减少数据库交互
[ ] 为常用查询添加索引
[ ] 使用 Select 只查询需要的字段
[ ] 使用 Scopes 复用查询条件
[ ] 设置合理的连接池大小
[ ] 使用 FindInBatches 处理大数据
[ ] 错误处理检查 gorm.ErrRecordNotFound
```
