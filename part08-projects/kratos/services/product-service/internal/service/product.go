// Package service 提供商品 gRPC 服务实现
package service

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "services/api/product/v1"
	"services/product-service/internal/biz"
)

// ProductService 商品服务
type ProductService struct {
	v1.UnimplementedProductServiceServer
	uc *biz.ProductUseCase
}

// NewProductService 创建商品服务
func NewProductService(uc *biz.ProductUseCase) *ProductService {
	return &ProductService{uc: uc}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.Product, error) {
	p, err := s.uc.Create(ctx, &biz.Product{
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Category:    req.GetCategory(),
		Price:       req.GetPrice(),
		Stock:       req.GetStock(),
		ImageURL:    req.GetImageUrl(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(p), nil
}

// GetProduct 获取商品
func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.Product, error) {
	p, err := s.uc.Get(ctx, req.GetId())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(p), nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.Product, error) {
	p, err := s.uc.Update(ctx, &biz.Product{
		ID:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Category:    req.GetCategory(),
		Price:       req.GetPrice(),
		Stock:       req.GetStock(),
		ImageURL:    req.GetImageUrl(),
		Status:      req.GetStatus(),
	})
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProto(p), nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.uc.Delete(ctx, req.GetId()); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

// ListProduct 商品列表
func (s *ProductService) ListProduct(ctx context.Context, req *v1.ListProductRequest) (*v1.ListProductReply, error) {
	products, total, err := s.uc.List(ctx, int(req.GetPage()), int(req.GetPageSize()), req.GetCategory(), req.GetKeyword())
	if err != nil {
		return nil, toGRPCError(err)
	}
	out := make([]*v1.Product, len(products))
	for i, p := range products {
		out[i] = toProto(p)
	}
	return &v1.ListProductReply{
		Products: out,
		Total:    int32(total),
		Page:     req.GetPage(),
		PageSize: req.GetPageSize(),
	}, nil
}

// DeductStock 扣减库存
func (s *ProductService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockReply, error) {
	p, err := s.uc.DeductStock(ctx, req.GetProductId(), req.GetQuantity())
	if err != nil {
		if errors.Is(err, biz.ErrStockNotEnough) {
			return &v1.DeductStockReply{Success: false, Message: err.Error()}, nil
		}
		return nil, toGRPCError(err)
	}
	return &v1.DeductStockReply{Success: true, Message: "ok", RemainingStock: p.Stock}, nil
}

// RestoreStock 恢复库存
func (s *ProductService) RestoreStock(ctx context.Context, req *v1.RestoreStockRequest) (*v1.DeductStockReply, error) {
	p, err := s.uc.RestoreStock(ctx, req.GetProductId(), req.GetQuantity())
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &v1.DeductStockReply{Success: true, Message: "ok", RemainingStock: p.Stock}, nil
}

// toProto biz.Product → proto Product
func toProto(p *biz.Product) *v1.Product {
	return &v1.Product{
		Id:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Stock:       p.Stock,
		ImageUrl:    p.ImageURL,
		Status:      p.Status,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, biz.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, biz.ErrStockNotEnough):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
