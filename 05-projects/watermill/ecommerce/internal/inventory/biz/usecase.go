package biz

import (
	"context"

	"go.uber.org/zap"
)

type Inventory struct {
	ProductID string
	Stock     int32
}

type InventoryRepo interface {
	Deduct(ctx context.Context, productID string, quantity int32) error
	Restore(ctx context.Context, productID string, quantity int32) error
	FindByProductID(ctx context.Context, productID string) (*Inventory, error)
}

type InventoryUseCase struct {
	repo InventoryRepo
	log  *zap.Logger
}

func NewInventoryUseCase(repo InventoryRepo, log *zap.Logger) *InventoryUseCase {
	return &InventoryUseCase{repo: repo, log: log}
}

func (uc *InventoryUseCase) Deduct(ctx context.Context, productID string, quantity int32) error {
	return uc.repo.Deduct(ctx, productID, quantity)
}

func (uc *InventoryUseCase) Restore(ctx context.Context, productID string, quantity int32) error {
	return uc.repo.Restore(ctx, productID, quantity)
}
