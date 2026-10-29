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

func TestUnsupportedURL(t *testing.T) {
	_, err := Get("https://youtube.com/watch?v=dQw4w9WgXcQ")
	if err == nil {
		t.Fatal("expected error for unsupported domain, got nil")
	}
}

func TestSetSilent(t *testing.T) {
	// Ensure SetSilent toggles without panic
	SetSilent(true)
	SetSilent(false)
}

func TestLiveSpotifyGet(t *testing.T) {
	url := "https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT"
	meta, err := Get(url)
	if err != nil {
		t.Fatalf("failed to get Spotify song: %v", err)
	}
	t.Logf("Title:       %s", meta.SongTitle)
	t.Logf("Artist:      %s", meta.Artist)
	t.Logf("Preview URL: %s", meta.PreviewURL)
	t.Logf("Image URL:   %s", meta.ImageURL)
	t.Logf("ISRC:        %s", meta.ISRC)
	t.Logf("Source:      %s", meta.Source)

	if meta.SongTitle == "" {
		t.Error("expected non-empty song title")
	}
}

func TestLiveAppleMusicGet(t *testing.T) {
	url := "https://music.apple.com/us/album/shape-of-you/1193701079?i=1193701392"
	meta, err := Get(url)
	if err != nil {
		t.Fatalf("failed to get Apple Music song: %v", err)
	}
	t.Logf("Title:       %s", meta.SongTitle)
	t.Logf("Artist:      %s", meta.Artist)
	t.Logf("Preview URL: %s", meta.PreviewURL)
	t.Logf("Image URL:   %s", meta.ImageURL)
	t.Logf("ISRC:        %s", meta.ISRC)
	t.Logf("Source:      %s", meta.Source)

	if meta.SongTitle == "" {
		t.Error("expected non-empty song title")
	}
}


