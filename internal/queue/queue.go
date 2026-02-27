package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/devrimsoft/bug-notifications-api/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	MainQueue = "bug_reports:queue"
	DLQQueue  = "bug_reports:dlq"
	MaxRetry  = 5
)

type Producer struct {
	rdb *redis.Client
}

func NewProducer(rdb *redis.Client) *Producer {
	return &Producer{rdb: rdb}
}

// Enqueue pushes a message to the main queue.
func (p *Producer) Enqueue(ctx context.Context, msg *model.QueueMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal queue message: %w", err)
	}
	return p.rdb.LPush(ctx, MainQueue, data).Err()
}

type Consumer struct {
	rdb *redis.Client
}

func NewConsumer(rdb *redis.Client) *Consumer {
	return &Consumer{rdb: rdb}
}

// Dequeue blocks until a message is available, then returns it.
func (c *Consumer) Dequeue(ctx context.Context) (*model.QueueMessage, error) {
	result, err := c.rdb.BRPop(ctx, 5*time.Second, MainQueue).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // timeout, no message
		}
		return nil, fmt.Errorf("brpop: %w", err)
	}

	// result[0] is key name, result[1] is value
	var msg model.QueueMessage
	if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
		return nil, fmt.Errorf("unmarshal queue message: %w", err)
	}
	return &msg, nil
}

// Requeue puts a failed message back for retry or into DLQ.
func (c *Consumer) Requeue(ctx context.Context, msg *model.QueueMessage) error {
	msg.RetryCount++
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal retry message: %w", err)
	}

	if msg.RetryCount >= MaxRetry {
		// Move to dead letter queue
		return c.rdb.LPush(ctx, DLQQueue, data).Err()
	}

	// Requeue to main queue
	return c.rdb.LPush(ctx, MainQueue, data).Err()
}

// DLQLength returns the number of messages in the dead letter queue.
func (c *Consumer) DLQLength(ctx context.Context) (int64, error) {
	return c.rdb.LLen(ctx, DLQQueue).Result()
}

// QueueLength returns the number of messages in the main queue.
func (c *Consumer) QueueLength(ctx context.Context) (int64, error) {
	return c.rdb.LLen(ctx, MainQueue).Result()
}
