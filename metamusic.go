// Package metamusic provides high-level tools to extract song metadata,
// audio previews, album artwork, and ISRC codes from Spotify and Apple Music URLs.
package metamusic

import (
	"context"

	"github.com/doi2523/get-music-metadata/pkg/music"
)

var defaultClient = NewClient()

// Get extracts song metadata auto-detecting whether the URL is Spotify or Apple Music.
func Get(rawURL string) (*SongMetadata, error) {
	return defaultClient.Get(rawURL)
}

// GetWithContext extracts song metadata with context support.
func GetWithContext(ctx context.Context, rawURL string) (*SongMetadata, error) {
	return defaultClient.GetWithContext(ctx, rawURL)
}

// GetSpotify extracts song metadata specifically for Spotify track URLs.
func GetSpotify(rawURL string) (*SongMetadata, error) {
	return defaultClient.GetSpotify(rawURL)
}

// GetSpotifyWithContext extracts Spotify metadata with context.
func GetSpotifyWithContext(ctx context.Context, rawURL string) (*SongMetadata, error) {
	return defaultClient.GetSpotifyWithContext(ctx, rawURL)
}

// GetAppleMusic extracts song metadata specifically for Apple Music track URLs.
func GetAppleMusic(rawURL string) (*SongMetadata, error) {
	return defaultClient.GetAppleMusic(rawURL)
}

// GetAppleMusicWithContext extracts Apple Music metadata with context.
func GetAppleMusicWithContext(ctx context.Context, rawURL string) (*SongMetadata, error) {
	return defaultClient.GetAppleMusicWithContext(ctx, rawURL)
}

// GetISRC resolves the ISRC code and cross-platform URLs for a song link.
func GetISRC(rawURL string) (*ISRCData, error) {
	return defaultClient.GetISRC(rawURL)
}

// GetISRCWithContext resolves ISRC with context.
func GetISRCWithContext(ctx context.Context, rawURL string) (*ISRCData, error) {
	return defaultClient.GetISRCWithContext(ctx, rawURL)
}

// SetSilent controls whether internal logging is silenced or printed to standard output.
func SetSilent(silent bool) {
	music.SetSilent(silent)
}
