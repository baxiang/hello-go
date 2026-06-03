// Package service 提供用户服务的 gRPC 接口实现
package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "services/api/user/v1"
	"services/user-service/internal/biz"
)

// UserService 用户服务
type UserService struct {
	v1.UnimplementedUserServiceServer
	uc *biz.UserUseCase
}

// NewUserService 创建用户服务
func NewUserService(uc *biz.UserUseCase) *UserService {
	return &UserService{uc: uc}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.User, error) {
	u, err := s.uc.Create(ctx, &biz.User{
		Username: req.GetUsername(),
		Password: req.GetPassword(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Nickname: req.GetNickname(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(u), nil
}

// GetUser 获取用户
func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.User, error) {
	u, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(u), nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.User, error) {
	u, err := s.uc.Update(ctx, &biz.User{
		ID:       req.GetId(),
		Email:    req.GetEmail(),
		Phone:    req.GetPhone(),
		Nickname: req.GetNickname(),
		Avatar:   req.GetAvatar(),
		Status:   req.GetStatus(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(u), nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.GetId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

// ListUser 用户列表
func (s *UserService) ListUser(ctx context.Context, req *v1.ListUserRequest) (*v1.ListUserReply, error) {
	users, total, err := s.uc.List(ctx, int(req.GetPage()), int(req.GetPageSize()), req.GetKeyword())
	if err != nil {
		return nil, toGRPCError(err)
	}
	out := make([]*v1.User, len(users))
	for i, u := range users {
		out[i] = toProto(u)
	}
	return &v1.ListUserReply{
		Users:    out,
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// Login 登录
func (s *UserService) Login(ctx context.Context, req *v1.LoginRequest) (*v1.LoginReply, error) {
	u, tok, err := s.uc.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &v1.LoginReply{Token: tok, User: toProto(u)}, nil
}

// Logout 登出（演示版：实际应将 token 加入黑名单）
func (s *UserService) Logout(ctx context.Context, req *v1.LogoutRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}

// ValidateToken 验证 Token
func (s *UserService) ValidateToken(ctx context.Context, req *v1.ValidateTokenRequest) (*v1.User, error) {
	u, err := s.uc.ValidateToken(ctx, req.GetToken())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(u), nil
}

// toProto biz.User → proto User
func toProto(u *biz.User) *v1.User {
	return &v1.User{
		Id:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Phone:     u.Phone,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		Status:    u.Status,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

// toGRPCError 将业务错误转为 gRPC 状态错误
func toGRPCError(err error) error {
	switch {
	case errors.Is(err, biz.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, biz.ErrInvalidPassword), errors.Is(err, biz.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
