package kafka

import "context"

type KafkaController interface {
	Consume(ctx context.Context) error
	Close() error
}
