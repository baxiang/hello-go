// Package data 提供用户数据访问层
package data

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"services/pkg/database"
	"services/user-service/internal/biz"
)

// userRepo 用户仓库实现
type userRepo struct {
	db  *database.DB
	log *zap.Logger
}

// NewUserRepo 创建用户仓库
func NewUserRepo(db *database.DB, log *zap.Logger) biz.UserRepo {
	return &userRepo{db: db, log: log}
}

// AutoMigrate 自动迁移用户表
func AutoMigrate(db *database.DB) error {
	return db.AutoMigrate(&biz.User{})
}

func (r *userRepo) Create(ctx context.Context, u *biz.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *userRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	var u biz.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var u biz.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) Update(ctx context.Context, u *biz.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&biz.User{}, id).Error
}

func (r *userRepo) List(ctx context.Context, page, pageSize int, keyword string) ([]*biz.User, int64, error) {
	var users []*biz.User
	var total int64

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	query := r.db.WithContext(ctx).Model(&biz.User{})
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?", like, like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// 静态接口检查
var _ biz.UserRepo = (*userRepo)(nil)

// 让 gorm 导入不报"unused"
var _ = gorm.ErrRecordNotFound
