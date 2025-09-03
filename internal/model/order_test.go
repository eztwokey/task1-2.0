package model

import "testing"

func TestOrderValidate(t *testing.T){
	if err := (&Order{}).Validate(); err == nil { t.Fatal("expected error") }
	o := &Order{OrderUID: "u", Data: map[string]any{"a":1}}
	if err := o.Validate(); err != nil { t.Fatalf("unexpected: %v", err) }
}
