package mq

import "sipon-be/internal/shared/messaging"

// RegisterHandlers mendaftarkan seluruh handler module akademik ke shared registry
// dan mengembalikan daftar binding queue. Memanggil dengan routing key yang sama
// lebih dari sekali akan menghasilkan error (registry menolak duplikat).
//
// Handler didaftarkan untuk routing key canonical dan legacy alias sekaligus,
// sehingga job lama yang masih menyimpan routing key format lama tetap dieksekusi.
func RegisterHandlers(registry *messaging.Registry, deps Dependencies) ([]messaging.Binding, error) {
	h := handlers{deps: deps}

	if err := registry.Register(RoutingFingerprintSync, h.handleFingerprintSync); err != nil {
		return nil, err
	}
	if err := registry.Register(LegacyRoutingFingerprintSync, h.handleFingerprintSync); err != nil {
		return nil, err
	}
	if err := registry.Register(RoutingSessionAutoClose, h.handleSessionAutoClose); err != nil {
		return nil, err
	}
	if err := registry.Register(LegacyRoutingSessionAutoClose, h.handleSessionAutoClose); err != nil {
		return nil, err
	}

	return Bindings, nil
}
