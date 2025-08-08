package service

import (
	"context"
	"log"
	"order/internal/entity"
	"order/internal/storage"
	"sync"
)

type service struct {
	store storage.Store
	cache map[string]entity.Order
	mu    sync.RWMutex
}

func NewService(store storage.Store) Service {
	return &service{
		store: store,
		cache: make(map[string]entity.Order),
		mu:    sync.RWMutex{},
	}
}

func (s *service) ProcessOrder(ctx context.Context, order entity.Order) error {
	if order.OrderUID == "" {
		log.Printf("Invalid order: empty order_uid")
		return nil
	}
	if len(order.Items) == 0 {
		log.Printf("Invalid order %s: empty items", order.OrderUID)
		return nil
	}
	if order.Delivery.Name == "" || order.Delivery.Phone == "" {
		log.Printf("Invalid order %s: invalid delivery (name: %s, phone: %s)",
			order.OrderUID, order.Delivery.Name, order.Delivery.Phone)
		return nil
	}
	if order.Payment.Amount < 0 {
		log.Printf("Invalid order %s: negative payment amount (%d)",
			order.OrderUID, order.Payment.Amount)
		return nil
	}

	// Сохранение в БД
	if err := s.store.SaveOrder(ctx, order); err != nil {
		log.Printf("Failed to save order %s: %v", order.OrderUID, err)
		return err
	}

	// Обновление кэша
	s.mu.Lock()
	s.cache[order.OrderUID] = order
	s.mu.Unlock()
	log.Printf("Order %s added to cache", order.OrderUID)

	return nil
}

func (s *service) GetOrder(ctx context.Context, orderUID string) (entity.Order, error) {
	// Проверка кэша
	s.mu.RLock()
	order, found := s.cache[orderUID]
	s.mu.RUnlock()
	if found {
		log.Printf("Order %s found in cache", orderUID)
		return order, nil
	}

	// Загрузка из БД
	log.Printf("Order %s not in cache, fetching from DB", orderUID)
	order, err := s.store.GetOrder(ctx, orderUID)
	if err != nil {
		log.Printf("Failed to fetch order %s from DB: %v", orderUID, err)
		return entity.Order{}, err
	}

	// Обновление кэша
	s.mu.Lock()
	s.cache[orderUID] = order
	s.mu.Unlock()
	log.Printf("Order %s added to cache from DB", orderUID)

	return order, nil
}

func (s *service) LoadCacheFromDB(ctx context.Context) error {
	orders, err := s.store.GetAllOrders(ctx)
	if err != nil {
		log.Printf("Failed to load cache from DB: %v", err)
		return err
	}

	s.mu.Lock()
	for _, order := range orders {
		s.cache[order.OrderUID] = order
	}
	s.mu.Unlock()
	log.Printf("Loaded %d orders into cache", len(orders))

	return nil
}
