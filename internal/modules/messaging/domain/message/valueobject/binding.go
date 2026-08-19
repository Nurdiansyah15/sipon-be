package valueobject

import "errors"

// Binding memetakan satu queue (consumer role) ke satu routing key. Module
// mengembalikan daftar Binding ini ketika mendaftarkan handler.
type Binding struct {
	Queue      string
	RoutingKey string
}

func (b Binding) Validate() error {
	if b.Queue == "" {
		return errors.New("messaging: binding queue wajib diisi")
	}
	if b.RoutingKey == "" {
		return errors.New("messaging: binding routing_key wajib diisi")
	}
	return nil
}
