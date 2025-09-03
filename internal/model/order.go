package model

import (
	"errors"
	"time"
)

// Order хранит метаданные + полный JSON в Data.
// Для простоты схема упрощена: order_uid обязателен.
type Order struct {
	OrderUID    string                 `json:"order_uid"`
	TrackNumber string                 `json:"track_number,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Data        map[string]any         `json:"data"`
}

var (
	ErrInvalidOrder = errors.New("invalid order: missing order_uid or data")
)

func (o *Order) Validate() error {
	if o == nil || o.OrderUID == "" || o.Data == nil {
		return ErrInvalidOrder
	}
	if o.CreatedAt.IsZero() { o.CreatedAt = time.Now().UTC() }
	return nil
}
