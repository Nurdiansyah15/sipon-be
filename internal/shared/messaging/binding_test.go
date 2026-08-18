package messaging

import "testing"

func TestBinding_Validate(t *testing.T) {
	if err := (Binding{Queue: "q", RoutingKey: "a.b"}).Validate(); err != nil {
		t.Fatalf("binding valid harus lolos: %v", err)
	}
	if err := (Binding{Queue: "", RoutingKey: "a.b"}).Validate(); err == nil {
		t.Fatal("queue kosong harus error")
	}
	if err := (Binding{Queue: "q", RoutingKey: ""}).Validate(); err == nil {
		t.Fatal("routing key kosong harus error")
	}
}
