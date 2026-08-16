package notification

import (
	"context"
	"errors"
	"fmt"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"sipon-be/internal/shared/config"
)

type SendRequest struct {
	Token string
	Topic string
	Title string
	Body  string
	Data  map[string]string
}

type Service struct {
	client       *messaging.Client
	defaultTopic string
	enabled      bool
}

func NewService(cfg config.FirebaseConfig) (*Service, error) {
	if !cfg.Enabled {
		return nil, errors.New("firebase notifications disabled")
	}

	ctx := context.Background()
	options := []option.ClientOption{}

	serviceAccountPath := strings.TrimSpace(cfg.ServiceAccountPath)
	serviceAccountJSON := strings.TrimSpace(cfg.ServiceAccountJSON)
	if serviceAccountPath != "" {
		options = append(options, option.WithCredentialsFile(serviceAccountPath))
	} else if serviceAccountJSON != "" {
		options = append(options, option.WithCredentialsJSON([]byte(serviceAccountJSON)))
	}

	projectID := strings.TrimSpace(cfg.ProjectID)
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, options...)
	if err != nil {
		return nil, fmt.Errorf("firebase app init failed: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging init failed: %w", err)
	}

	return &Service{
		client:       client,
		defaultTopic: strings.TrimSpace(cfg.DefaultTopic),
		enabled:      true,
	}, nil
}

func (s *Service) Send(ctx context.Context, req SendRequest) error {
	if s == nil || !s.enabled || s.client == nil {
		return errors.New("firebase service unavailable")
	}

	token := strings.TrimSpace(req.Token)
	topic := strings.TrimSpace(req.Topic)
	if topic == "" {
		topic = s.defaultTopic
	}
	if strings.TrimSpace(req.Title) == "" {
		req.Title = "Sipon Notification"
	}
	if strings.TrimSpace(req.Body) == "" {
		req.Body = "Anda menerima notifikasi baru dari Sipon"
	}
	if req.Data == nil {
		req.Data = map[string]string{}
	}

	msg := buildFCMMessage(req.Title, req.Body, token, topic, req.Data)
	_, err := s.client.Send(ctx, msg)
	return err
}

func buildFCMMessage(title, body, token, topic string, data map[string]string) *messaging.Message {
	if data == nil {
		data = map[string]string{}
	}

	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
				},
			},
		},
	}
	if strings.TrimSpace(token) != "" {
		msg.Token = token
	}
	if strings.TrimSpace(topic) != "" {
		msg.Topic = topic
	}
	return msg
}
