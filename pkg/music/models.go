package music

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
