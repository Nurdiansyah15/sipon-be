package notification

import (
	"testing"
)

func TestBuildFCMMessage_UsesTokenAndData(t *testing.T) {
	title := "Sipon test"
	body := "This is a push test"
	token := "device-token-123"
	data := map[string]string{"screen": "/dashboard", "type": "test"}

	msg := buildFCMMessage(title, body, token, "", data)
	if msg == nil {
		t.Fatal("expected message to be built")
	}
	if msg.Token != token {
		t.Fatalf("expected token %q, got %q", token, msg.Token)
	}
	if msg.Notification == nil || msg.Notification.Title != title || msg.Notification.Body != body {
		t.Fatalf("expected notification payload to match title/body")
	}
	if msg.Data["screen"] != "/dashboard" || msg.Data["type"] != "test" {
		t.Fatalf("expected data payload to match, got %#v", msg.Data)
	}
	if msg.Android == nil || msg.Android.Priority != "high" {
		t.Fatalf("expected Android priority to be high")
	}
}
