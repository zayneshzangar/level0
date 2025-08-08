package service

import (
	"context"
	"log"
	"order/internal/entity"
	"order/internal/storage"
	"sync"

	"github.com/go-playground/validator/v10"
	lru "github.com/hashicorp/golang-lru/v2"
)

type service struct {
	store storage.Store
	cache *lru.Cache[string, entity.Order]
	mu    sync.Mutex
}

func NewService(store storage.Store) Service {
	cache, err := lru.New[string, entity.Order](1000) // Лимит 1000 заказов
	if err != nil {
		log.Fatalf("Failed to create LRU cache: %v", err)
	}
	return &service{
		store: store,
		cache: cache,
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

	// Валидация с использованием validator
	validate := validator.New()
	if err := validate.Struct(order); err != nil {
		log.Printf("Invalid order %s: %v", order.OrderUID, err)
		return nil
	}

	// Сохранение в БД
	if err := s.store.SaveOrder(ctx, order); err != nil {
		log.Printf("Failed to save order %s: %v", order.OrderUID, err)
		return err
	}

	// Обновление кэша
	s.mu.Lock()
	s.cache.Add(order.OrderUID, order)
	s.mu.Unlock()
	log.Printf("Order %s added to cache", order.OrderUID)
	return nil
}

func (s *service) GetOrder(ctx context.Context, orderUID string) (entity.Order, error) {
	s.mu.Lock()
	order, ok := s.cache.Get(orderUID)
	s.mu.Unlock()
	if ok {
		log.Printf("Order %s found in cache", orderUID)
		return order, nil
	}

	order, err := s.store.GetOrder(ctx, orderUID)
	if err != nil {
		log.Printf("Failed to get order %s from DB: %v", orderUID, err)
		return entity.Order{}, err
	}

	s.mu.Lock()
	s.cache.Add(orderUID, order)
	s.mu.Unlock()
	log.Printf("Order %s fetched from DB and added to cache", orderUID)
	return order, nil
}

func (s *service) LoadCacheFromDB(ctx context.Context) error {
	orders, err := s.store.GetAllOrders(ctx)
	if err != nil {
		log.Printf("Failed to load orders from DB: %v", err)
		return err
	}

	s.mu.Lock()
	for _, order := range orders {
		s.cache.Add(order.OrderUID, order)
	}
	s.mu.Unlock()
	log.Printf("Loaded %d orders into cache", s.cache.Len())
	return nil
}
