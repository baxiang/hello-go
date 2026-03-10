package service

import (
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "kratos/api/product/v1"
	"kratos/internal/biz"
)

// ProductService 商品服务
type ProductService struct {
	v1.UnimplementedProductServiceServer

	uc *biz.ProductUseCase
}

// NewProductService 创建商品服务
func NewProductService(uc *biz.ProductUseCase) *ProductService {
	return &ProductService{
		uc: uc,
	}
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.Product, error) {
	product := &biz.Product{
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageUrl,
		Status:      1,
	}

	result, err := s.uc.Create(ctx, product)
	if err != nil {
		return nil, err
	}

	return s.toProto(result), nil
}

// GetProduct 获取商品
func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.Product, error) {
	product, err := s.uc.Get(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return s.toProto(product), nil
}

// UpdateProduct 更新商品
func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.Product, error) {
	product := &biz.Product{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Price:       req.Price,
		Stock:       req.Stock,
		ImageURL:    req.ImageUrl,
		Status:      req.Status,
	}

	result, err := s.uc.Update(ctx, product)
	if err != nil {
		return nil, err
	}

	return s.toProto(result), nil
}

// DeleteProduct 删除商品
func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	err := s.uc.Delete(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// ListProduct 商品列表
func (s *ProductService) ListProduct(ctx context.Context, req *v1.ListProductRequest) (*v1.ListProductReply, error) {
	products, total, err := s.uc.List(ctx, int(req.Page), int(req.PageSize), req.Category, req.Keyword)
	if err != nil {
		return nil, err
	}

	protoProducts := make([]*v1.Product, len(products))
	for i, product := range products {
		protoProducts[i] = s.toProto(product)
	}

	return &v1.ListProductReply{
		Products: protoProducts,
		Total:    int32(total),
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// DeductStock 扣减库存
func (s *ProductService) DeductStock(ctx context.Context, req *v1.DeductStockRequest) (*v1.DeductStockReply, error) {
	product, err := s.uc.DeductStock(ctx, req.ProductId, req.Quantity)
	if err != nil {
		return &v1.DeductStockReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.DeductStockReply{
		Success:        true,
		Message:        "库存扣减成功",
		RemainingStock: product.Stock,
	}, nil
}

// toProto 转换为 Protobuf 消息
func (s *ProductService) toProto(product *biz.Product) *v1.Product {
	return &v1.Product{
		Id:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Price:       product.Price,
		Stock:       product.Stock,
		ImageUrl:    product.ImageURL,
		Status:      product.Status,
		CreatedAt:   timestamppb.New(product.CreatedAt),
		UpdatedAt:   timestamppb.New(product.UpdatedAt),
	}
}