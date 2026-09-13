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
	var targetURL string

	// Prefer raw query parsing to preserve URLs with nested query strings
	if strings.Contains(r.URL.RawQuery, "url=") {
		parts := strings.SplitN(r.URL.RawQuery, "url=", 2)
		if len(parts) > 1 && parts[1] != "" {
			targetURL = parts[1]
			// Strip endpoint query params like &platform= if appended unencoded
			if idx := strings.Index(targetURL, "&platform="); idx != -1 {
				targetURL = targetURL[:idx]
			}
			if unescaped, err := url.QueryUnescape(targetURL); err == nil && strings.HasPrefix(unescaped, "http") {
				targetURL = unescaped
			}
		}
	}

	if targetURL == "" {
		targetURL = r.URL.Query().Get("url")
	}

	if targetURL == "" && r.Method == http.MethodPost && r.Body != nil {
		var reqBody struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil && reqBody.URL != "" {
			targetURL = reqBody.URL
		}
	}

	return targetURL
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
	Link     string `json:"link"`
	MusicURL string `json:"music_url"`
	SongURL  string `json:"song_url"`
	Platform string `json:"platform"`
}

type GetInfoMusicV3Response struct {
	Artist        *string `json:"artist,omitempty"`
	ImageURL      *string `json:"image_url,omitempty"`
	ISRC          *string `json:"isrc,omitempty"`
	PreviewURL    *string `json:"preview_url,omitempty"`
	SongName      *string `json:"song_name,omitempty"`
	AppleMusicURL *string `json:"apple_music_url,omitempty"`
	SpotifyURL    *string `json:"spotify_url,omitempty"`
	Title         *string `json:"title,omitempty"`
	SongTitle     *string `json:"song_title,omitempty"`
	Album         *string `json:"album,omitempty"`
	Platform      *string `json:"platform,omitempty"`
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type GetInfoMusicV3APIResponse struct {
	Status  string                  `json:"status"`
	Message string                  `json:"message"`
	Data    *GetInfoMusicV3Response `json:"data"`
}

func handleGetInfoMusicV3(w http.ResponseWriter, r *http.Request) {
	log.Printf("==================================================")
	log.Printf("🚀 [getInfoMusicV3] Received Request [%s %s]", r.Method, r.URL.Path)

	var reqBody GetInfoMusicV3Request
	if r.Method == http.MethodPost && r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			log.Printf("⚠️ [getInfoMusicV3] Failed to decode JSON body: %v", err)
		}
	}

	targetURL := reqBody.URL
	if targetURL == "" {
		if reqBody.Link != "" {
			targetURL = reqBody.Link
		} else if reqBody.MusicURL != "" {
			targetURL = reqBody.MusicURL
		} else if reqBody.SongURL != "" {
			targetURL = reqBody.SongURL
		} else {
			targetURL = ExtractURLFromRequest(r)
		}
	}

	platformReq := reqBody.Platform
	if platformReq == "" {
		platformReq = r.URL.Query().Get("platform")
	}

	log.Printf("📥 [getInfoMusicV3] Target URL: %s | Platform: %s", targetURL, platformReq)

	if targetURL == "" {
		log.Printf("❌ [getInfoMusicV3] Rejected: Missing 'url' parameter")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": "Thiếu tham số 'url'.",
			"data":    nil,
		})
		return
	}

	meta, err := music.GetSongMetadata(targetURL)
	if err != nil {
		log.Printf("❌ [getInfoMusicV3] Metadata extraction failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "error",
			"message": fmt.Sprintf("Không thể lấy thông tin bài hát: %v", err),
			"data":    nil,
		})
		return
	}

	platformRes := strings.ToLower(platformReq)
	if platformRes == "" {
		if strings.Contains(strings.ToLower(meta.Source), "spotify") || strings.Contains(targetURL, "spotify.com") {
			platformRes = "spotify"
		} else {
			platformRes = "apple"
		}
	}

	var appleMusicURL *string
	var spotifyURL *string

	switch platformRes {
	case "spotify":
		spotifyURL = strPtr(meta.SpotifyURL)

	case "apple":
		appleMusicURL = strPtr(meta.AppleMusicURL)
	}

	log.Printf("✨ [getInfoMusicV3] Extracted: Title='%s', Artist='%s', ISRC='%s', PreviewURL='%s'", meta.SongTitle, meta.Artist, meta.ISRC, meta.PreviewURL)

	resp := GetInfoMusicV3APIResponse{
		Status:  "success",
		Message: "ok",
		Data: &GetInfoMusicV3Response{
			Artist:        strPtr(meta.Artist),
			ImageURL:      strPtr(meta.ImageURL),
			ISRC:          strPtr(meta.ISRC),
			PreviewURL:    strPtr(meta.PreviewURL),
			SongName:      strPtr(meta.SongTitle),
			AppleMusicURL: appleMusicURL,
			SpotifyURL:    spotifyURL,
			Title:         strPtr(meta.SongTitle),
			SongTitle:     strPtr(meta.SongTitle),
			Album:         strPtr(meta.Album),
			Platform:      strPtr(platformRes),
		},
	}

	log.Printf("📤 [getInfoMusicV3] Returning HTTP 200 OK")
	log.Printf("==================================================")

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
