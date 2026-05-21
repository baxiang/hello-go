// Package biz 提供用户业务逻辑
package biz

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"services/pkg/token"
)

// 业务错误
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrInvalidToken      = errors.New("无效的令牌")
)

// User 用户实体
type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Username  string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"type:varchar(128);index" json:"email"`
	Phone     string    `gorm:"type:varchar(32)" json:"phone"`
	Nickname  string    `gorm:"type:varchar(64)" json:"nickname"`
	Avatar    string    `gorm:"type:varchar(256)" json:"avatar"`
	Password  string    `gorm:"type:varchar(128);not null" json:"-"`
	Status    int32     `gorm:"default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 自定义表名
func (User) TableName() string {
	return "users"
}

// UserRepo 用户仓库接口
type UserRepo interface {
	Create(ctx context.Context, u *User) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, page, pageSize int, keyword string) ([]*User, int64, error)
}

// UserUseCase 用户用例
type UserUseCase struct {
	repo  UserRepo
	token *token.Manager
	log   *zap.Logger
}

// NewUserUseCase 创建用户用例
func NewUserUseCase(repo UserRepo, tokenMgr *token.Manager, log *zap.Logger) *UserUseCase {
	return &UserUseCase{
		repo:  repo,
		token: tokenMgr,
		log:   log,
	}
}

// Create 创建用户
func (uc *UserUseCase) Create(ctx context.Context, u *User) (*User, error) {
	// 检查用户名是否已存在
	existing, err := uc.repo.FindByUsername(ctx, u.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// 加密密码
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.Password = string(hashed)
	if u.Status == 0 {
		u.Status = 1
	}

	if err := uc.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	uc.log.Info("创建用户成功", zap.Int64("id", u.ID), zap.String("username", u.Username))
	return u, nil
}

// Get 获取用户
func (uc *UserUseCase) Get(ctx context.Context, id int64) (*User, error) {
	u, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// Update 更新用户
func (uc *UserUseCase) Update(ctx context.Context, u *User) (*User, error) {
	existing, err := uc.repo.FindByID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	existing.Email = u.Email
	existing.Phone = u.Phone
	existing.Nickname = u.Nickname
	existing.Avatar = u.Avatar
	if u.Status != 0 {
		existing.Status = u.Status
	}

	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

// Delete 删除用户
func (uc *UserUseCase) Delete(ctx context.Context, id int64) error {
	if _, err := uc.repo.FindByID(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return uc.repo.Delete(ctx, id)
}

// List 用户列表
func (uc *UserUseCase) List(ctx context.Context, page, pageSize int, keyword string) ([]*User, int64, error) {
	return uc.repo.List(ctx, page, pageSize, keyword)
}

// Login 登录
func (uc *UserUseCase) Login(ctx context.Context, username, password string) (*User, string, error) {
	u, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	tok, err := uc.token.Generate(u.ID, u.Username)
	if err != nil {
		return nil, "", err
	}
	uc.log.Info("用户登录成功", zap.Int64("id", u.ID))
	return u, tok, nil
}

// ValidateToken 验证 Token
func (uc *UserUseCase) ValidateToken(ctx context.Context, tokStr string) (*User, error) {
	claims, err := uc.token.Validate(tokStr)
	if err != nil {
		return nil, ErrInvalidToken
	}
	u, err := uc.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}
