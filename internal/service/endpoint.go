package service

import (
	"context"
	"order/internal/entity"
)

type Service interface {
	ProcessOrder(ctx context.Context, order entity.Order) error
	GetOrder(ctx context.Context, orderUID string) (entity.Order, error)
	LoadCacheFromDB(ctx context.Context) error
}
