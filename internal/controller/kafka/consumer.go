package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"order/internal/entity"
	"order/internal/service"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type kafkaController struct {
	consumer *kafka.Consumer
	service  service.Service
}

func NewKafkaController(brokers, groupID, topic string, service service.Service) (KafkaController, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  brokers,
		"group.id":           groupID,
		"auto.offset.reset":  "earliest", // Начать с самого начала топика
		"enable.auto.commit": false,      // Ручное подтверждение смещений
	})
	if err != nil {
		return nil, err
	}

	// Подписка на топик
	err = consumer.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		consumer.Close()
		return nil, err
	}

	return &kafkaController{
		consumer: consumer,
		service:  service,
	}, nil
}

func (c *kafkaController) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil // Грациозное завершение
		default:
			// Чтение сообщения с таймаутом 1 секунда
			msg, err := c.consumer.ReadMessage(1 * time.Second)
			if err != nil {
				var kafkaErr kafka.Error
				if errors.As(err, &kafkaErr) && kafkaErr.Code() == kafka.ErrTimedOut {
					continue // Таймаут, продолжаем цикл
				}
				log.Printf("Failed to read message: %v", err)
				continue
			}

			// Десериализация сообщения
			var order entity.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Failed to unmarshal message: %v", err)
				continue
			}

			// Обработка заказа через сервис
			if err := c.service.ProcessOrder(ctx, order); err != nil {
				log.Printf("Failed to process order %s: %v", order.OrderUID, err)
				continue
			}

			// Ручное подтверждение смещения
			_, err = c.consumer.CommitMessage(msg)
			if err != nil {
				log.Printf("Failed to commit message: %v", err)
				continue
			}

			log.Printf("Successfully processed order %s", order.OrderUID)
		}
	}
}

func (c *kafkaController) Close() error {
	return c.consumer.Close()
}
