// Package database 提供 MySQL 数据库连接
package database

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"services/pkg/config"
)

// DB 数据库客户端
type DB struct {
	*gorm.DB
	log *zap.Logger
}

// New 创建数据库连接
func New(cfg *config.DatabaseConfig, log *zap.Logger) (*DB, error) {
	gormLogger := logger.New(
		zapLogWriter{log: log.Named("gorm")},
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(cfg.Source), &gorm.Config{
		Logger: gormLogger,
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

	sqlDB.SetMaxOpenConns(cfg.MaxOpen)
	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.MaxLife) * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping 数据库失败: %w", err)
	}

	log.Info("数据库连接成功",
		zap.String("driver", cfg.Driver),
		zap.Int("max_open", cfg.MaxOpen),
	)
	return &DB{DB: db, log: log}, nil
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// zapLogWriter 适配 gorm logger Writer 接口
type zapLogWriter struct {
	log *zap.Logger
}

// Printf 实现 gorm logger.Writer 接口
func (w zapLogWriter) Printf(format string, args ...interface{}) {
	w.log.Sugar().Infof(format, args...)
}
