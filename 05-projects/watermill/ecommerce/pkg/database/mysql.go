// Package database 提供 GORM + MySQL 连接管理
package database

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 数据库连接封装
type DB struct {
	*gorm.DB
	log *zap.Logger
}

// New 创建数据库连接
func New(dsn string, log *zap.Logger) (*DB, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Info("数据库连接成功")
	return &DB{DB: db, log: log}, nil
}

// Close 关闭数据库连接
func (d *DB) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	d.log.Info("数据库连接已关闭")
	return sqlDB.Close()
}
