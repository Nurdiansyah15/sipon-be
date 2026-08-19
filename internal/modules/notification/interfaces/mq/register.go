package mq

import "sipon-be/internal/modules/messaging"

func RegisterHandlers(registry messaging.Contract, deps Dependencies) ([]messaging.Binding, error) {
	h := handlers{deps: deps}

	registrations := []struct {
		routingKey string
		handler    messaging.HandlerFunc
	}{
		{RoutingLoginSucceeded, h.handleLoginSucceeded},
		{RoutingPsbPendaftaranSubmitted, h.handlePsbPendaftaranSubmitted},
		{RoutingPsbDaftarUlangSubmitted, h.handlePsbDaftarUlangSubmitted},
		{RoutingPsbDokumenVerified, h.handlePsbDokumenVerified},
		{RoutingPsbDokumenRejected, h.handlePsbDokumenRejected},
		{RoutingPsbRevisionRequested, h.handlePsbRevisionRequested},
		{RoutingPsbRevisionRequestedDaftarUlang, h.handlePsbRevisionRequestedDaftarUlang},
		{RoutingPsbPendaftaranAccepted, h.handlePsbPendaftaranAccepted},
		{RoutingPsbPendaftaranRejected, h.handlePsbPendaftaranRejected},
		{RoutingPsbNISGenerated, h.handlePsbNISGenerated},
		{RoutingArticlePublished, h.handleArticlePublished},
		{RoutingArticlesScraped, h.handleArticlesScraped},
	}

	for _, r := range registrations {
		if err := registry.Register(r.routingKey, r.handler); err != nil {
			return nil, err
		}
	}

	return Bindings, nil
}
