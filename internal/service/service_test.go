package service

import (
    "context"
    "errors"
    "order/internal/entity"
    "order/internal/storage/mock"
    "testing"

    "github.com/hashicorp/golang-lru/v2"
    "github.com/stretchr/testify/assert"
    "go.uber.org/mock/gomock"
)

func setupService(t *testing.T) (*service, *mock.MockStore, *gomock.Controller) {
    ctrl := gomock.NewController(t)
    mockStore := mock.NewMockStore(ctrl)
    cache, _ := lru.New[string, entity.Order](1000)
    svc := &service{
        store: mockStore,
        cache: cache,
    }
    return svc, mockStore, ctrl
}

func TestService_ProcessOrder(t *testing.T) {
    ctx := context.Background()

    t.Run("Valid order", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        order := entity.Order{
            OrderUID:    "test-uid",
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        mockStore.EXPECT().SaveOrder(ctx, order).Return(nil)

        err := svc.ProcessOrder(ctx, order)
        assert.NoError(t, err)
        cachedOrder, ok := svc.cache.Get(order.OrderUID)
        assert.True(t, ok)
        assert.Equal(t, order, cachedOrder)
    })

    t.Run("Empty order", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        order := entity.Order{} // Пустой заказ проходит валидацию
        mockStore.EXPECT().SaveOrder(ctx, order).Return(nil)

        err := svc.ProcessOrder(ctx, order)
        assert.NoError(t, err)
        cachedOrder, ok := svc.cache.Get(order.OrderUID)
        assert.True(t, ok)
        assert.Equal(t, order, cachedOrder)
    })

    t.Run("SaveOrder error", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        order := entity.Order{
            OrderUID:    "test-uid",
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        mockStore.EXPECT().SaveOrder(ctx, order).Return(errors.New("db error"))

        err := svc.ProcessOrder(ctx, order)
        assert.Error(t, err)
        assert.Equal(t, "db error", err.Error())
        assert.Equal(t, 0, svc.cache.Len())
    })

    t.Run("Order with empty OrderUID", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        order := entity.Order{
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        mockStore.EXPECT().SaveOrder(ctx, order).Return(nil)

        err := svc.ProcessOrder(ctx, order)
        assert.NoError(t, err)
        cachedOrder, ok := svc.cache.Get(order.OrderUID)
        assert.True(t, ok)
        assert.Equal(t, order, cachedOrder)
    })

    t.Run("Order with negative Amount", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        order := entity.Order{
            OrderUID:    "test-uid",
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: -1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        mockStore.EXPECT().SaveOrder(ctx, order).Return(nil)

        err := svc.ProcessOrder(ctx, order)
        assert.NoError(t, err)
        cachedOrder, ok := svc.cache.Get(order.OrderUID)
        assert.True(t, ok)
        assert.Equal(t, order, cachedOrder)
    })
}

func TestService_GetOrder(t *testing.T) {
    ctx := context.Background()

    t.Run("Order in cache", func(t *testing.T) {
        svc, _, ctrl := setupService(t)
        defer ctrl.Finish()

        orderUID := "test-uid"
        order := entity.Order{
            OrderUID:    orderUID,
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        svc.cache.Add(orderUID, order)

        result, err := svc.GetOrder(ctx, orderUID)
        assert.NoError(t, err)
        assert.Equal(t, order, result)
    })

    t.Run("Order from DB", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        orderUID := "test-uid"
        order := entity.Order{
            OrderUID:    orderUID,
            Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
            Payment:     entity.Payment{Amount: 1000},
            Items:       []entity.Item{{ChrtID: 1, Price: 500}},
            DateCreated: "2025-08-09T10:30:00Z",
        }
        mockStore.EXPECT().GetOrder(ctx, orderUID).Return(order, nil)

        result, err := svc.GetOrder(ctx, orderUID)
        assert.NoError(t, err)
        assert.Equal(t, order, result)
        cachedOrder, ok := svc.cache.Get(orderUID)
        assert.True(t, ok)
        assert.Equal(t, order, cachedOrder)
    })

    t.Run("Order not found", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        orderUID := "test-uid"
        mockStore.EXPECT().GetOrder(ctx, orderUID).Return(entity.Order{}, errors.New("not found"))

        result, err := svc.GetOrder(ctx, orderUID)
        assert.Error(t, err)
        assert.Equal(t, "not found", err.Error())
        assert.Equal(t, entity.Order{}, result)
        assert.Equal(t, 0, svc.cache.Len())
    })

    t.Run("Empty orderUID", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        orderUID := ""
        mockStore.EXPECT().GetOrder(ctx, orderUID).Return(entity.Order{}, errors.New("not found"))

        result, err := svc.GetOrder(ctx, orderUID)
        assert.Error(t, err)
        assert.Equal(t, "not found", err.Error())
        assert.Equal(t, entity.Order{}, result)
        assert.Equal(t, 0, svc.cache.Len())
    })
}

func TestService_LoadCacheFromDB(t *testing.T) {
    ctx := context.Background()

    t.Run("Successful load", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        orders := []entity.Order{
            {
                OrderUID:    "uid1",
                Delivery:    entity.Delivery{Name: "John", Phone: "1234567890"},
                Payment:     entity.Payment{Amount: 1000},
                Items:       []entity.Item{{ChrtID: 1, Price: 500}},
                DateCreated: "2025-08-09T10:30:00Z",
            },
            {
                OrderUID:    "uid2",
                Delivery:    entity.Delivery{Name: "Jane", Phone: "0987654321"},
                Payment:     entity.Payment{Amount: 2000},
                Items:       []entity.Item{{ChrtID: 2, Price: 1000}},
                DateCreated: "2025-08-09T10:30:00Z",
            },
        }
        mockStore.EXPECT().GetAllOrders(ctx).Return(orders, nil)

        err := svc.LoadCacheFromDB(ctx)
        assert.NoError(t, err)
        assert.Equal(t, 2, svc.cache.Len())
        for _, order := range orders {
            cachedOrder, ok := svc.cache.Get(order.OrderUID)
            assert.True(t, ok)
            assert.Equal(t, order, cachedOrder)
        }
    })

    t.Run("DB error", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        mockStore.EXPECT().GetAllOrders(ctx).Return(nil, errors.New("db error"))

        err := svc.LoadCacheFromDB(ctx)
        assert.Error(t, err)
        assert.Equal(t, "db error", err.Error())
        assert.Equal(t, 0, svc.cache.Len())
    })

    t.Run("Empty result from DB", func(t *testing.T) {
        svc, mockStore, ctrl := setupService(t)
        defer ctrl.Finish()

        mockStore.EXPECT().GetAllOrders(ctx).Return([]entity.Order{}, nil)

        err := svc.LoadCacheFromDB(ctx)
        assert.NoError(t, err)
        assert.Equal(t, 0, svc.cache.Len())
    })
}
