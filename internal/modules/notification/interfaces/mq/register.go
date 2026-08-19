package mq

import "sipon-be/internal/modules/messaging"

func RegisterHandlers(registry messaging.Contract, deps Dependencies) ([]messaging.Binding, error) {
	h := handlers{deps: deps}

	if err := registry.Register(RoutingLoginSucceeded, h.handleLoginSucceeded); err != nil {
		return nil, err
	}

	return Bindings, nil
}
