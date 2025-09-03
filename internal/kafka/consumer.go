package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"wb-order-service/internal/model"
	"wb-order-service/internal/service"
)

type Consumer struct {
	reader *kafka.Reader
	svc    *service.Service
	log    *slog.Logger
}

func NewConsumer(brokers []string, topic, groupID string, svc *service.Service, logger *slog.Logger) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     brokers,
			Topic:       topic,
			GroupID:     groupID,
			StartOffset: kafka.LastOffset,
			MaxWait:     time.Second,
		}),
		svc: svc,
		log: logger,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	c.log.Info("kafka consumer started")
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil { return ctx.Err() }
			c.log.Error("kafka read", "err", err)
			continue
		}
		var o model.Order
		if err := json.Unmarshal(m.Value, &o); err != nil {
			c.log.Error("json unmarshal", "err", err)
			continue
		}
		if err := o.Validate(); err != nil {
			c.log.Warn("invalid order", "err", err)
			continue
		}
		if err := c.svc.UpsertOrder(ctx, &o); err != nil {
			c.log.Error("upsert order", "err", err)
			continue
		}
		c.log.Debug("stored order", "uid", o.OrderUID)
	}
}

func (c *Consumer) Close() error { return c.reader.Close() }
