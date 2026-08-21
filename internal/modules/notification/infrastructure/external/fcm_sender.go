package external

import (
	"context"
	"fmt"
	"log/slog"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"

	"sipon-be/internal/shared/config"
)

// PushMessage adalah pesan yang akan dikirim ke device via push provider.
type PushMessage struct {
	Token    string
	Title    string
	Body     string
	ImageURL *string
	Payload  map[string]string
	Priority string
	// UnreadCount dipakai sebagai badge APNs (iOS). 0 = kosongkan badge.
	UnreadCount int
}

// PushResult adalah status pengiriman untuk satu token.
type PushResult struct {
	Token        string
	TokenInvalid bool
}

// PushSender adalah kontrak untuk mengirim push notification ke device.
type PushSender interface {
	Send(ctx context.Context, msg PushMessage) error
	SendBatch(ctx context.Context, msgs []PushMessage) ([]PushResult, error)
}

// FCMPushSender adalah implementasi PushSender menggunakan Firebase Cloud Messaging.
type FCMPushSender struct {
	client *messaging.Client
	logger *slog.Logger
}

func NewFCMPushSender(ctx context.Context, cfg *config.FCMConfig) (PushSender, error) {
	if cfg.CredentialsPath == "" {
		slog.Warn("FCM credentials path kosong, push notification dinonaktifkan")
		return &NoopPushSender{}, nil
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: cfg.ProjectID}, option.WithCredentialsFile(cfg.CredentialsPath))
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi Firebase app: %w", err)
	}
	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi FCM messaging client: %w", err)
	}
	return &FCMPushSender{client: client, logger: slog.Default()}, nil
}

func (s *FCMPushSender) Send(ctx context.Context, msg PushMessage) error {
	_, err := s.client.Send(ctx, buildMessage(msg))
	if err != nil {
		return fmt.Errorf("fcm send gagal: %w", err)
	}
	return nil
}

func (s *FCMPushSender) SendBatch(ctx context.Context, msgs []PushMessage) ([]PushResult, error) {
	if len(msgs) == 0 {
		return nil, nil
	}

	messages := make([]*messaging.Message, 0, len(msgs))
	for _, m := range msgs {
		messages = append(messages, buildMessage(m))
	}

	resp, err := s.client.SendEach(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("fcm send batch gagal: %w", err)
	}

	if resp.FailureCount == 0 {
		s.logger.Info("fcm send batch sukses", slog.Int("sent", resp.SuccessCount))
		return nil, nil
	}

	results := make([]PushResult, 0, resp.FailureCount)
	for i, r := range resp.Responses {
		if !r.Success {
			token := msgs[i].Token
			if isPermanentTokenError(r.Error) {
				results = append(results, PushResult{Token: token, TokenInvalid: true})
			}
		}
	}

	if resp.SuccessCount == 0 {
		return results, fmt.Errorf("fcm: semua %d pesan gagal dikirim", resp.FailureCount)
	}
	return results, nil
}

func isPermanentTokenError(err error) bool {
	return messaging.IsUnregistered(err) ||
		messaging.IsRegistrationTokenNotRegistered(err) ||
		messaging.IsInvalidArgument(err) ||
		messaging.IsSenderIDMismatch(err)
}

func buildMessage(msg PushMessage) *messaging.Message {
	priority := msg.Priority
	if priority == "" {
		priority = "normal"
	}
	apnsPriority := "5"
	if priority == "high" {
		apnsPriority = "10"
	}
	badge := msg.UnreadCount
	if badge < 0 {
		badge = 0
	}

	return &messaging.Message{
		Token: msg.Token,
		Notification: &messaging.Notification{
			Title: msg.Title,
			Body:  msg.Body,
			ImageURL: func() string {
				if msg.ImageURL != nil && *msg.ImageURL != "" {
					return *msg.ImageURL
				}
				return ""
			}(),
		},
		Data: msg.Payload,
		Android: &messaging.AndroidConfig{
			Priority: priority,
			Notification: &messaging.AndroidNotification{
				ChannelID: "sipon_notifications",
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{"apns-priority": apnsPriority},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound:            "default",
					ContentAvailable: true,
					Badge:            &badge,
				},
			},
		},
	}
}

// NoopPushSender tidak melakukan apa-apa (development/test).
type NoopPushSender struct{}

func (s *NoopPushSender) Send(_ context.Context, _ PushMessage) error { return nil }
func (s *NoopPushSender) SendBatch(_ context.Context, _ []PushMessage) ([]PushResult, error) {
	return nil, nil
}
