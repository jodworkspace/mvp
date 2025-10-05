package mqx

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerHandler func(ctx context.Context, body []byte) error

type RabbitClient interface {
	Close() error
	Publish(ctx context.Context) error
	Consume(ctx context.Context, handler ConsumerHandler) error
}

type rabbitClient struct {
	mu       sync.RWMutex
	conn     *amqp.Connection
	channels chan amqp.Channel
}

func NewRabbitClient(url string, poolSize int) (RabbitClient, error) {
	conn, err := amqp.DialConfig(url, amqp.Config{})
	if err != nil {
		return nil, err
	}

	c := &rabbitClient{
		conn:     conn,
		channels: make(chan amqp.Channel, poolSize),
	}

	return c, nil
}

func (rc *rabbitClient) Close() error {
	close(rc.channels)
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return rc.conn.Close()
}

func (rc *rabbitClient) Publish(ctx context.Context) error {
	return nil
}

func (rc *rabbitClient) Consume(ctx context.Context, handler ConsumerHandler) error {
	return handler(ctx, nil)
}
