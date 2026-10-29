package metamusic

import (
	"context"
	"testing"
	"time"
)

func TestNewClientOptions(t *testing.T) {
	client := NewClient(
		WithTimeout(5*time.Second),
		WithSilent(),
	)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.timeout != 5*time.Second {
		t.Errorf("expected timeout 5s, got %v", client.timeout)
	}
}

func TestContextCancellation(t *testing.T) {
	client := NewClient(WithSilent())
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetWithContext(ctx, "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT")
	if err == nil {
		t.Fatal("expected error due to canceled context, got nil")
	}
}
