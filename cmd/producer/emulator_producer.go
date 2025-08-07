package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"order/config"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
)

// Order соответствует структуре entity.Order
type Order struct {
	OrderUID          string   `json:"order_uid"`
	TrackNumber       string   `json:"track_number"`
	Entry             string   `json:"entry"`
	Delivery          Delivery `json:"delivery"`
	Payment           Payment  `json:"payment"`
	Items             []Item   `json:"items"`
	Locale            string   `json:"locale"`
	InternalSignature string   `json:"internal_signature"`
	CustomerID        string   `json:"customer_id"`
	DeliveryService   string   `json:"delivery_service"`
	Shardkey          string   `json:"shardkey"`
	SmID              int      `json:"sm_id"`
	DateCreated       string   `json:"date_created"`
	OofShard          string   `json:"oof_shard"`
}

type Delivery struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Zip     string `json:"zip"`
	City    string `json:"city"`
	Address string `json:"address"`
	Region  string `json:"region"`
	Email   string `json:"email"`
}

type Payment struct {
	Transaction  string `json:"transaction"`
	RequestID    string `json:"request_id"`
	Currency     string `json:"currency"`
	Provider     string `json:"provider"`
	Amount       int    `json:"amount"`
	PaymentDt    int64  `json:"payment_dt"`
	Bank         string `json:"bank"`
	DeliveryCost int    `json:"delivery_cost"`
	GoodsTotal   int    `json:"goods_total"`
	CustomFee    int    `json:"custom_fee"`
}

type Item struct {
	ChrtID      int    `json:"chrt_id"`
	TrackNumber string `json:"track_number"`
	Price       int    `json:"price"`
	Rid         string `json:"rid"`
	Name        string `json:"name"`
	Sale        int    `json:"sale"`
	Size        string `json:"size"`
	TotalPrice  int    `json:"total_price"`
	NmID        int    `json:"nm_id"`
	Brand       string `json:"brand"`
	Status      int    `json:"status"`
}

func main() {
	// Загрузка конфигурации из переменных окружения
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Ошибка env: %v", err)
	}
	var builder strings.Builder
	builder.Grow(64)
	fmt.Fprintf(&builder, "%s:%s,%s:%s,%s:%s",
		cfg.Kafka.Host, cfg.Kafka.Port1,
		cfg.Kafka.Host, cfg.Kafka.Port2,
		cfg.Kafka.Host, cfg.Kafka.Port3,
	)
	brokers := builder.String()
	topic := cfg.Kafka.Topic
	if brokers == "" || topic == "" {
		log.Fatal("KAFKA_BROKERS or KAFKA_TOPIC not set")
	}
	log.Printf("Kafka brokers: %s, topic: %s", brokers, topic)

	// Инициализация продюсера
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer producer.Close()

	// Канал для получения отчётов о доставке
	deliveryChan := make(chan kafka.Event)
	go func() {
		for e := range deliveryChan {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Failed to deliver message: %v", ev.TopicPartition.Error)
				} else {
					log.Printf("Message delivered to topic %s [partition %d] at offset %v",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			case kafka.Error:
				log.Printf("Kafka error: %v", ev)
			}
		}
	}()

	// Контекст для грациозного завершения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Генерация и отправка сообщений
	go func() {
		rand.Seed(time.Now().UnixNano())
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Генерация тестового заказа
				order := generateOrder()
				data, err := json.Marshal(order)
				if err != nil {
					log.Printf("Failed to marshal order: %v", err)
					continue
				}

				// Отправка сообщения
				err = producer.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
					Value:          data,
					Key:            []byte(order.OrderUID),
				}, deliveryChan)
				if err != nil {
					log.Printf("Failed to produce message: %v", err)
				}

				log.Printf("Produced order %s", order.OrderUID)
				time.Sleep(2 * time.Second) // Пауза между сообщениями
			}
		}
	}()

	// Ожидание сигналов для завершения
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	// Остановка продюсера
	cancel()
	producer.Flush(5000) // Ждём доставки оставшихся сообщений
	log.Println("Producer stopped")
}

// generateOrder создаёт тестовый заказ
func generateOrder() Order {
	uid := uuid.New().String()
	track := "TN" + strconv.Itoa(rand.Intn(1000))
	return Order{
		OrderUID:    uid,
		TrackNumber: track,
		Entry:       "WBIL",
		Delivery: Delivery{
			Name:    "John Doe",
			Phone:   "+1234567890",
			Zip:     "12345",
			City:    "Moscow",
			Address: "123 Main St",
			Region:  "Central",
			Email:   "john@example.com",
		},
		Payment: Payment{
			Transaction:  "TX" + strconv.Itoa(rand.Intn(1000)),
			RequestID:    "REQ" + strconv.Itoa(rand.Intn(1000)),
			Currency:     "USD",
			Provider:     "stripe",
			Amount:       rand.Intn(10000) + 100,
			PaymentDt:    time.Now().Unix(),
			Bank:         "Sberbank",
			DeliveryCost: 200,
			GoodsTotal:   rand.Intn(5000) + 500,
			CustomFee:    0,
		},
		Items: []Item{
			{
				ChrtID:      rand.Intn(1000000),
				TrackNumber: track,
				Price:       rand.Intn(1000) + 100,
				Rid:         "RID" + strconv.Itoa(rand.Intn(1000)),
				Name:        "Item " + strconv.Itoa(rand.Intn(100)),
				Sale:        rand.Intn(50),
				Size:        "M",
				TotalPrice:  rand.Intn(900) + 90,
				NmID:        rand.Intn(1000),
				Brand:       "BrandX",
				Status:      1,
			},
		},
		Locale:            "en",
		InternalSignature: "",
		CustomerID:        "CUST" + strconv.Itoa(rand.Intn(1000)),
		DeliveryService:   "UPS",
		Shardkey:          "1",
		SmID:              99,
		DateCreated:       time.Now().Format(time.RFC3339),
		OofShard:          "1",
	}
}
