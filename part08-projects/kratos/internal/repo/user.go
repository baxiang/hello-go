package repo

import (
	"context"

	"gorm.io/gorm"

	"kratos/internal/biz"
)

// UserRepo 用户仓库接口
type UserRepo interface {
	Create(ctx context.Context, user *biz.User) error
	FindByID(ctx context.Context, id int64) (*biz.User, error)
	FindByUsername(ctx context.Context, username string) (*biz.User, error)
	FindByEmail(ctx context.Context, email string) (*biz.User, error)
	Update(ctx context.Context, user *biz.User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, keyword string) ([]*biz.User, int64, error)
}

// userRepo 用户仓库实现
type userRepo struct {
	data *data.Data
	log  interface{}
}

// NewUserRepo 创建用户仓库
func NewUserRepo(data *data.Data, logger interface{}) UserRepo {
	return &userRepo{
		data: data,
		log:  logger,
	}
}

// Create 创建用户
func (r *userRepo) Create(ctx context.Context, user *biz.User) error {
	return r.data.DB.WithContext(ctx).Create(user).Error
}

// FindByID 根据 ID 查找
func (r *userRepo) FindByID(ctx context.Context, id int64) (*biz.User, error) {
	var user biz.User
	err := r.data.DB.WithContext(ctx).First(&user, id).Error
	return &user, err
}

// FindByUsername 根据用户名查找
func (r *userRepo) FindByUsername(ctx context.Context, username string) (*biz.User, error) {
	var user biz.User
	err := r.data.DB.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

// FindByEmail 根据邮箱查找
func (r *userRepo) FindByEmail(ctx context.Context, email string) (*biz.User, error) {
	var user biz.User
	err := r.data.DB.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return &user, err
}

// Update 更新用户
func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	return r.data.DB.WithContext(ctx).Save(user).Error
}

// Delete 删除用户
func (r *userRepo) Delete(ctx context.Context, id int64) error {
	return r.data.DB.WithContext(ctx).Delete(&biz.User{}, id).Error
}

// List 用户列表
func (r *userRepo) List(ctx context.Context, page, pageSize int, keyword string) ([]*biz.User, int64, error) {
	var users []*biz.User
	var total int64

	query := r.data.DB.WithContext(ctx).Model(&biz.User{})

	// 关键词搜索
	if keyword != "" {
		query = query.Where("username LIKE ? OR email LIKE ? OR nickname LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// 确保实现接口
var _ UserRepo = (*userRepo)(nil)