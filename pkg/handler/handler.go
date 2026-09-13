// Package handler provides shared HTTP handlers for both the local server
// (cmd/server) and the Vercel serverless function (api/index.go).
package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"get-meta-music/pkg/music"
)

// Handler is the single entry point for all HTTP requests.
// It applies CORS headers and routes based on URL path.
func Handler(w http.ResponseWriter, r *http.Request) {
	applyCORSHeaders(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	path := r.URL.Path
	switch {
	case strings.HasSuffix(path, "/spotify"), strings.HasSuffix(path, "/api/spotify"):
		handleSpotify(w, r)
	case strings.HasSuffix(path, "/apple-music"), strings.HasSuffix(path, "/api/apple-music"):
		handleAppleMusic(w, r)
	case strings.HasSuffix(path, "/getInfoMusicV3"), strings.HasSuffix(path, "/api/getInfoMusicV3"):
		handleGetInfoMusicV3(w, r)
	case strings.HasSuffix(path, "/metadata"), strings.HasSuffix(path, "/api/metadata"):
		handleMetadata(w, r)
	default:
		handleRoot(w, r)
	}
}

// applyCORSHeaders sets permissive CORS headers and JSON content type.
func applyCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// ExtractURLFromRequest extracts the target music URL from query params or POST body.
func ExtractURLFromRequest(r *http.Request) string {
	// Prefer raw query parsing to preserve URLs with nested query strings
	if strings.Contains(r.URL.RawQuery, "url=") {
		parts := strings.SplitN(r.URL.RawQuery, "url=", 2)
		if len(parts) > 1 && parts[1] != "" {
			targetURL := parts[1]
			if unescaped, err := url.QueryUnescape(targetURL); err == nil && strings.HasPrefix(unescaped, "http") {
				return unescaped
			}
			return targetURL
		}
	}

	if rawURL := r.URL.Query().Get("url"); rawURL != "" {
		return rawURL
	}

	if r.Method == http.MethodPost {
		var reqBody struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil && reqBody.URL != "" {
			return reqBody.URL
		}
	}

	return ""
}

func handleMetadata(w http.ResponseWriter, r *http.Request) {
	rawURL := ExtractURLFromRequest(r)
	if rawURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(music.APIResponse{
			Success: false,
			Error:   "Thiếu tham số 'url'. Vui lòng truyền ?url=https://...",
		})
		return
	}

	log.Printf("📥 [%s %s] Parsing URL: %s", r.Method, r.URL.Path, rawURL)

	meta, err := music.GetSongMetadata(rawURL)
	if err != nil {
		log.Printf("❌ Error parsing metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(music.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Không thể lấy thông tin bài hát: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(meta)
}

func handleSpotify(w http.ResponseWriter, r *http.Request) {
	rawURL := ExtractURLFromRequest(r)
	if rawURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(music.APIResponse{
			Success: false,
			Error:   "Thiếu tham số 'url'. Vui lòng truyền ?url=https://...",
		})
		return
	}

	log.Printf("📥 [%s %s] Parsing Spotify URL: %s", r.Method, r.URL.Path, rawURL)

	meta, err := music.GetSpotifyMetadata(rawURL)
	if err != nil {
		log.Printf("❌ Error parsing Spotify metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(music.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Lỗi lấy thông tin Spotify: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(meta)
}

func handleAppleMusic(w http.ResponseWriter, r *http.Request) {
	rawURL := ExtractURLFromRequest(r)
	if rawURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(music.APIResponse{
			Success: false,
			Error:   "Thiếu tham số 'url'. Vui lòng truyền ?url=https://...",
		})
		return
	}

	log.Printf("📥 [%s %s] Parsing Apple Music URL: %s", r.Method, r.URL.Path, rawURL)

	meta, err := music.GetAppleMusicMetadata(rawURL)
	if err != nil {
		log.Printf("❌ Error parsing Apple Music metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(music.APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Lỗi lấy thông tin Apple Music: %v", err),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(meta)
}

type GetInfoMusicV3Request struct {
	URL      string `json:"url"`
	Platform string `json:"platform"`
}

type GetInfoMusicV3Response struct {
	Artist        string `json:"artist"`
	ImageURL      string `json:"image_url"`
	ISRC          string `json:"isrc,omitempty"`
	PreviewURL    string `json:"preview_url,omitempty"`
	SongName      string `json:"song_name"`
	AppleMusicURL string `json:"apple_music_url,omitempty"`
	SpotifyURL    string `json:"spotify_url,omitempty"`
	Title         string `json:"title"`
	SongTitle     string `json:"song_title"`
	Album         string `json:"album,omitempty"`
	Platform      string `json:"platform"`
}

func handleGetInfoMusicV3(w http.ResponseWriter, r *http.Request) {
	var reqBody GetInfoMusicV3Request
	if r.Method == http.MethodPost && r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
	}

	if reqBody.URL == "" {
		reqBody.URL = ExtractURLFromRequest(r)
	}

	if reqBody.Platform == "" {
		reqBody.Platform = r.URL.Query().Get("platform")
	}

	if reqBody.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Thiếu tham số 'url'.",
		})
		return
	}

	log.Printf("📥 [%s %s] Parsing V3 URL: %s (platform: %s)", r.Method, r.URL.Path, reqBody.URL, reqBody.Platform)

	var meta *music.SongMetadata
	var err error

	platformLower := strings.ToLower(reqBody.Platform)
	if platformLower == "apple" || platformLower == "apple-music" {
		meta, err = music.GetAppleMusicMetadata(reqBody.URL)
	} else if platformLower == "spotify" {
		meta, err = music.GetSpotifyMetadata(reqBody.URL)
	} else {
		meta, err = music.GetSongMetadata(reqBody.URL)
	}

	if err != nil {
		log.Printf("❌ Error parsing V3 metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Không thể lấy thông tin bài hát: %v", err),
		})
		return
	}

	platformRes := "apple"
	if strings.Contains(strings.ToLower(meta.Source), "spotify") || strings.Contains(reqBody.URL, "spotify.com") {
		platformRes = "spotify"
	}

	resp := GetInfoMusicV3Response{
		Artist:        meta.Artist,
		ImageURL:      meta.ImageURL,
		ISRC:          meta.ISRC,
		PreviewURL:    meta.PreviewURL,
		SongName:      meta.SongTitle,
		AppleMusicURL: meta.AppleMusicURL,
		SpotifyURL:    meta.SpotifyURL,
		Title:         meta.SongTitle,
		SongTitle:     meta.SongTitle,
		Platform:      platformRes,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	rawURL := ExtractURLFromRequest(r)
	if rawURL != "" {
		handleMetadata(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"server":  "Music Metadata Extractor (Go Server)",
		"status":  "running",
		"version": "1.0.0",
		"usage": map[string]string{
			"GET":     "/metadata?url=https://open.spotify.com/track/...",
			"POST":    "/metadata with JSON body {\"url\": \"https://...\"}",
			"POST V3": "/getInfoMusicV3 (or /api/getInfoMusicV3) with JSON body {\"url\": \"https://...\", \"platform\": \"spotify|apple\"}",
		},
	})
}
