package biz

import (
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"kratos/internal/data"
	"kratos/internal/repo"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserAlreadyExists = errors.New("用户已存在")
	ErrInvalidPassword   = errors.New("密码错误")
	ErrInvalidToken      = errors.New("无效的令牌")
)

// User 用户业务实体
type User struct {
	ID        int64     `json:"id"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Nickname   string    `json:"nickname"`
	Avatar     string    `json:"avatar"`
	Password   string    `json:"-"`
	Status     int32     `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UserUseCase 用户用例
type UserUseCase struct {
	userRepo repo.UserRepo
	log      *log.Helper
}

// NewUserUseCase 创建用户用例
func NewUserUseCase(userRepo repo.UserRepo, logger log.Logger) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		log:      log.NewHelper(logger),
	}
}

// Create 创建用户
func (uc *UserUseCase) Create(ctx context.Context, u *User) (*User, error) {
	// 检查用户名是否已存在
	existing, err := uc.userRepo.FindByUsername(ctx, u.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrUserAlreadyExists
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u.Password = string(hashedPassword)
	u.Status = 1

	// 创建用户
	if err := uc.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	uc.log.Info("创建用户成功", log.Any("user", u))
	return u, nil
}

// Get 获取用户
func (uc *UserUseCase) Get(ctx context.Context, id int64) (*User, error) {
	user, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// Update 更新用户
func (uc *UserUseCase) Update(ctx context.Context, u *User) (*User, error) {
	existing, err := uc.userRepo.FindByID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// 更新字段
	existing.Email = u.Email
	existing.Phone = u.Phone
	existing.Nickname = u.Nickname
	existing.Avatar = u.Avatar
	existing.Status = u.Status

	if err := uc.userRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	uc.log.Info("更新用户成功", log.Any("user", existing))
	return existing, nil
}

// Delete 删除用户
func (uc *UserUseCase) Delete(ctx context.Context, id int64) error {
	existing, err := uc.userRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	if err := uc.userRepo.Delete(ctx, existing.ID); err != nil {
		return err
	}

	uc.log.Info("删除用户成功", log.Any("id", id))
	return nil
}

// List 用户列表
func (uc *UserUseCase) List(ctx context.Context, page, pageSize int, keyword string) ([]*User, int64, error) {
	return uc.userRepo.List(ctx, page, pageSize, keyword)
}

// Login 登录
func (uc *UserUseCase) Login(ctx context.Context, username, password string) (*User, string, error) {
	user, err := uc.userRepo.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrUserNotFound
		}
		return nil, "", err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidPassword
	}

	// 生成 token (简化版，实际应使用 JWT)
	token := data.GenerateToken(user.ID, user.Username)

	uc.log.Info("用户登录成功", log.Any("user", user))
	return user, token, nil
}

// Logout 登出
func (uc *UserUseCase) Logout(ctx context.Context, token string) error {
	// 简化版，实际应将 token 加入黑名单
	uc.log.Info("用户登出", log.Any("token", token))
	return nil
}

// ValidateToken 验证 Token
func (uc *UserUseCase) ValidateToken(ctx context.Context, token string) (*User, error) {
	claims, err := data.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	user, err := uc.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}