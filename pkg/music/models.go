package music

import "github.com/doi2523/get-music-metadata/internal/logger"

// Logger defines the minimal logging interface.
type Logger = logger.Logger

// NopLogger is a no-op logger that discards all log messages.
var NopLogger = logger.NopLogger

// SetLogger sets the active logger.
func SetLogger(l Logger) {
	logger.SetLogger(l)
}

// SetSilent silences or enables internal logging.
func SetSilent(silent bool) {
	logger.SetSilent(silent)
}

func logPrintf(format string, v ...interface{}) {
	logger.Printf(format, v...)
}

func logPrintln(v ...interface{}) {
	logger.Println(v...)
}

type SongMetadata struct {
	SongTitle     string `json:"song_title,omitempty"`
	Artist        string `json:"artist,omitempty"`
	Album         string `json:"album,omitempty"`
	SpotifyURL    string `json:"spotify_url,omitempty"`
	AppleMusicURL string `json:"apple_music_url,omitempty"`
	PreviewURL    string `json:"preview_url,omitempty"`
	ImageURL      string `json:"image_url,omitempty"`
	ISRC          string `json:"isrc,omitempty"`
	Source        string `json:"source,omitempty"`
}

type ISRCData struct {
	ISRC          string `json:"isrc"`
	Title         string `json:"title,omitempty"`
	Artist        string `json:"artist,omitempty"`
	PreviewURL    string `json:"preview_url,omitempty"`
	SpotifyURL    string `json:"spotify_url,omitempty"`
	AppleMusicURL string `json:"apple_music_url,omitempty"`
	Source        string `json:"source,omitempty"`
}

type APIResponse struct {
	Success bool          `json:"success"`
	Data    *SongMetadata `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}

