package storage

import (
	"context"
	"order/internal/entity"
)

type Store interface {
	SaveOrder(ctx context.Context, order entity.Order) error
	GetOrder(ctx context.Context, orderUID string) (entity.Order, error)
	GetAllOrders(ctx context.Context) ([]entity.Order, error)
}
