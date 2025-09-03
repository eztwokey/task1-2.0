package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"context"
	"errors"
	"wb-order-service/internal/cache"
	"wb-order-service/internal/model"
	"wb-order-service/internal/repo"
	"wb-order-service/internal/service"
	"log/slog"
	"os"
)

type stubRepo struct{ m map[string]*model.Order }
func (s *stubRepo) GetOrder(ctx context.Context, id string)(*model.Order,error){ if v,ok:=s.m[id]; ok { return v,nil }; return nil, repo.ErrNotFound }
func (s *stubRepo) UpsertOrder(ctx context.Context, o *model.Order) error { s.m[o.OrderUID]=o; return nil }
func (s *stubRepo) Ping(ctx context.Context) error { return nil }
func (s *stubRepo) Close() {}

func TestGetOrder(t *testing.T){
	c := cache.NewTTL(); c.StartJanitor(time.Minute); defer c.Stop()
	r := &stubRepo{m: map[string]*model.Order{"x": {OrderUID:"x", Data: map[string]any{"p":1}}}}
	s := service.New(r, c, time.Minute)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	h := New(s, logger, nil, []string{"kafka:9092"}, "orders")
	req := httptest.NewRequest(http.MethodGet, "/order/x", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)
	if rw.Code != 200 { t.Fatalf("code=%d", rw.Code) }
	var got model.Order
	if err := json.Unmarshal(rw.Body.Bytes(), &got); err != nil { t.Fatal(err) }
	if got.OrderUID != "x" { t.Fail() }
}
