package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
	"strings"

	"github.com/brianvoe/gofakeit/v7"
	kafka "github.com/segmentio/kafka-go"
)

func main() {
	brokers := getenv("KAFKA_BROKERS", "kafka:9092")
	topic := getenv("KAFKA_ORDERS_TOPIC", "orders")
	n := flag.Int("n", 10, "number of orders to generate")
	flag.Parse()

	w := &kafka.Writer{
		Addr:         kafka.TCP(splitComma(brokers)...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	defer w.Close()

	ctx := context.Background()
	rand.Seed(time.Now().UnixNano())
	gofakeit.Seed(time.Now().UnixNano())

	for i := 0; i < *n; i++ {
		uid := gofakeit.UUID()
		order := map[string]any{
			"order_uid": uid,
			"track_number": gofakeit.BS(),
			"created_at": time.Now().UTC(),
			"data": map[string]any{
				"customer_id": gofakeit.UUID(),
				"name": gofakeit.Name(),
				"address": gofakeit.Address().Address,
				"price": gofakeit.Price(10, 1000),
			},
		}
		b, _ := json.Marshal(order)
		if err := w.WriteMessages(ctx, kafka.Message{Value: b}); err != nil {
			log.Printf("write message: %v", err)
		} else {
			fmt.Println(uid)
		}
	}
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" { out = append(out, p) }
	}
	return out
}

