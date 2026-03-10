package database

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// DB 数据库客户端
type DB struct {
	*gorm.DB
	log *zap.Logger
}

// NewDB 创建数据库连接
func NewDB(cfg *Config, log *zap.Logger) (*DB, error) {
	dsn := cfg.Source

	gormLogger := logger.New(
		log.Named("gorm"),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:      gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		NowFunc: func() time.Time {
			return time.Now().In(time.Local)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层连接失败: %w", err)
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLife) * time.Second)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("测试数据库连接失败: %w", err)
	}

	log.Info("数据库连接成功",
		zap.String("driver", cfg.Driver),
		zap.Int("max_open", cfg.MaxOpen),
		zap.Int("max_idle", cfg.MaxIdle),
	)

	return &DB{DB: db, log: log}, nil
}

// Close 关闭连接
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Transaction 事务
func (d *DB) Transaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(fn)
}

// AutoMigrate 自动迁移
func (d *DB) AutoMigrate(models ...interface{}) error {
	return d.DB.AutoMigrate(models...)
}

// Config 数据库配置
type Config struct {
	Driver   string
	Source   string
	MaxOpen  int
	MaxIdle  int
	MaxLife  int
}

// NewConfig 创建配置
func NewConfig(driver, source string, maxOpen, maxIdle, maxLife int) *Config {
	return &Config{
		Driver:  driver,
		Source:  source,
		MaxOpen: maxOpen,
		MaxIdle: maxIdle,
		MaxLife: maxLife,
	}
}