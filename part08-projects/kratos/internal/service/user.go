package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos/api/user/v1"
	"kratos/internal/biz"
)

// UserService 用户服务
type UserService struct {
	v1.UnimplementedUserServiceServer

	uc *biz.UserUseCase
}

// NewUserService 创建用户服务
func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{
		uc: uc,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.User, error) {
	user := &biz.User{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
	}

	result, err := s.uc.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.toProto(result), nil
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.User, error) {
	user, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return s.toProto(user), nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.User, error) {
	user := &biz.User{
		ID:       req.Id,
		Email:    req.Email,
		Phone:    req.Phone,
		Nickname: req.Nickname,
		Avatar:   req.Avatar,
		Status:   req.Status,
	}

	result, err := s.uc.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.toProto(result), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*emptypb.Empty, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ListUser 用户列表
func (s *UserService) ListUser(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserReply, error) {
	users, total, err := s.uc.List(ctx, int(req.Page), int(req.PageSize), req.Keyword)
	if err != nil {
		return nil, err
	}

	protoUsers := make([]*v1.User, len(users))
	for i, user := range users {
		protoUsers[i] = s.toProto(user)
	}

	return &v1.ListUserReply{
		Users:    protoUsers,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// Login 登录
func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	user, token, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &v1.LoginReply{
		Token: token,
		User:  s.toProto(user),
	}, nil
}

// Logout 登出
func (s *UserService) Logout(ctx context.Context, req *v1.LogoutRequest) (*emptypb.Empty, error) {
	err := s.uc.Logout(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// toProto 转换为 Protobuf 消息
func (s *UserService) toProto(user *biz.User) *v1.User {
	return &v1.User{
		Id:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Phone:     user.Phone,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}