package cache

import (
	"testing"
	"time"
	"wb-order-service/internal/model"
)

func TestTTLCache(t *testing.T){
	c := NewTTL()
	c.StartJanitor(10 * time.Millisecond)
	defer c.Stop()
	o := &model.Order{OrderUID: "1", Data: map[string]any{"x":1}}
	c.Set("1", o, 20*time.Millisecond)
	if v, ok := c.Get("1"); !ok || v.OrderUID != "1" { t.Fatal("miss") }
	time.Sleep(30*time.Millisecond)
	if _, ok := c.Get("1"); ok { t.Fatal("should expire") }
}
