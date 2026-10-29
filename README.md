# 🎵 Get Meta Music (`get-meta-music`)

[![Go Reference](https://pkg.go.dev/badge/github.com/doi2523/get-meta-music.svg)](https://pkg.go.dev/github.com/doi2523/get-meta-music)
[![Go Report Card](https://goreportcard.com/badge/github.com/doi2523/get-meta-music)](https://goreportcard.com/report/github.com/doi2523/get-meta-music)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

🌐 **Language**: **English** | [Tiếng Việt](README_VI.md)

**`get-meta-music`** is a comprehensive Golang toolkit and microservice designed to extract music metadata from **Spotify** and **Apple Music**:
- 🎧 **Audio Previews**: Extract 30-second high-quality audio preview streams (MP3/AAC).
- 🖼 **HD Artwork**: Fetch original high-resolution album cover artwork (up to 600x600px).
- 🆔 **ISRC Lookup**: Resolve International Standard Recording Codes (ISRC) and cross-link between Spotify ⇄ Apple Music.
- ⚡ **Dual-Mode Architecture**: Works both as an importable **Go Package / SDK** for other Go projects and as a **Standalone HTTP REST API Server** & **Vercel Serverless Function**.

---

## 📑 Table of Contents

1. [Installation & Download](#-installation--download)
2. [Go Library / SDK Usage](#-go-library--sdk-usage)
   - [Quickstart with Helper Functions](#1-quickstart-with-helper-functions)
   - [Custom Client with Advanced Options (Timeout, Proxy, Silent, Logger)](#2-custom-client-with-advanced-options)
   - [Context & Cancellation Support](#3-context--cancellation-support)
   - [Data Models Reference](#4-data-models-reference)
3. [Running the HTTP Server & REST API](#-running-the-http-server--rest-api)
   - [Running Locally](#1-running-locally)
   - [Deploying to Vercel Serverless](#2-deploying-to-vercel-serverless)
   - [API Endpoints Summary](#3-api-endpoints-summary)
   - [Endpoint Details & cURL Examples](#4-endpoint-details--curl-examples)
4. [Environment Variables](#-environment-variables)
5. [Project Structure](#-project-structure)
6. [License](#-license)

---

## 📦 Installation & Download

### Option 1: Install as a Go Module (SDK)
Run the following command inside your Go project:
```bash
go get github.com/doi2523/get-meta-music
```

### Option 2: Clone repository to run Server or contribute
```bash
git clone https://github.com/doi2523/get-meta-music.git
cd get-meta-music
```

---

## 💻 Go Library / SDK Usage

Import the package into your Go application:
```go
import metamusic "github.com/doi2523/get-meta-music"
```

### 1. Quickstart with Helper Functions

```go
package main

import (
	"fmt"
	"log"

	metamusic "github.com/doi2523/get-meta-music"
)

func main() {
	// Auto-detects whether the URL is Spotify or Apple Music
	url := "https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT"

	meta, err := metamusic.Get(url)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Printf("🎵 Song Title:  %s\n", meta.SongTitle)
	fmt.Printf("👤 Artist:      %s\n", meta.Artist)
	fmt.Printf("💿 Album:       %s\n", meta.Album)
	fmt.Printf("🎧 Preview URL: %s\n", meta.PreviewURL)
	fmt.Printf("🖼 Image URL:   %s\n", meta.ImageURL)
	fmt.Printf("🆔 ISRC Code:   %s\n", meta.ISRC)
	fmt.Printf("🍏 Apple Music: %s\n", meta.AppleMusicURL)
	fmt.Printf("🟢 Spotify:     %s\n", meta.SpotifyURL)
}
```

Top-level convenience functions:
- `metamusic.Get(url)`: Automatically detects whether the link is from Spotify or Apple Music.
- `metamusic.GetSpotify(url)`: Extracts metadata specifically for Spotify links.
- `metamusic.GetAppleMusic(url)`: Extracts metadata specifically for Apple Music links.
- `metamusic.GetISRC(url)`: Resolves ISRC and cross-platform URLs across streaming services.
- `metamusic.SetSilent(true)`: Suppresses all internal log messages from printing to standard output.

---

### 2. Custom Client with Advanced Options

For production backend services, microservices, or scraping behind proxies, configure a custom `Client`:

```go
package main

import (
	"fmt"
	"time"

	metamusic "github.com/doi2523/get-meta-music"
)

func main() {
	// Initialize a customized client
	client := metamusic.NewClient(
		metamusic.WithTimeout(10*time.Second),                    // 10s timeout
		metamusic.WithSilent(),                                   // Disable stdout logging
		// metamusic.WithProxy("http://username:pass@proxy-ip:port"), // Optional proxy support
	)

	appleURL := "https://music.apple.com/vn/album/shape-of-you/1201597002?i=1201597040"
	meta, err := client.GetAppleMusic(appleURL)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Song: %s - %s\nISRC: %s\nPreview: %s\n",
		meta.SongTitle, meta.Artist, meta.ISRC, meta.PreviewURL)
}
```

Supported Functional Options:
- `WithTimeout(d time.Duration)`: Custom HTTP timeout.
- `WithProxy(proxyURL string)`: HTTP, HTTPS, or SOCKS5 proxy support.
- `WithSilent()`: Disables all internal logging completely.
- `WithLogger(l music.Logger)`: Injects custom logger (Zap, Logrus, Slog, etc.).
- `WithHTTPClient(client *http.Client)`: Uses an existing shared HTTP client pool.

---

### 3. Context & Cancellation Support

Every client method supports a `context.Context` variant for timeouts and cooperative cancellation:
- `client.GetWithContext(ctx, url)`
- `client.GetSpotifyWithContext(ctx, url)`
- `client.GetAppleMusicWithContext(ctx, url)`
- `client.GetISRCWithContext(ctx, url)`

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

meta, err := client.GetWithContext(ctx, "https://open.spotify.com/track/...")
if err != nil {
    // Returns context.DeadlineExceeded or context.Canceled if the request exceeds time limit
    fmt.Println("Request canceled or timed out:", err)
}
```

---

### 4. Data Models Reference

#### `SongMetadata`
```go
type SongMetadata struct {
	SongTitle     string `json:"song_title,omitempty"`     // Song title
	Artist        string `json:"artist,omitempty"`         // Artist or band name
	Album         string `json:"album,omitempty"`          // Album name
	SpotifyURL    string `json:"spotify_url,omitempty"`    // Spotify canonical or matched URL
	AppleMusicURL string `json:"apple_music_url,omitempty"`// Apple Music canonical or matched URL
	PreviewURL    string `json:"preview_url,omitempty"`    // 30-second audio stream URL (MP3/M4A)
	ImageURL      string `json:"image_url,omitempty"`      // High-resolution artwork URL
	ISRC          string `json:"isrc,omitempty"`           // International Standard Recording Code
	Source        string `json:"source,omitempty"`         // Data source (Spotify, Apple Music, Deezer...)
}
```

#### `ISRCData`
```go
type ISRCData struct {
	ISRC          string `json:"isrc"`                     // Resolved ISRC code
	Title         string `json:"title,omitempty"`          // Song title from ISRC lookup
	Artist        string `json:"artist,omitempty"`         // Artist from ISRC lookup
	PreviewURL    string `json:"preview_url,omitempty"`    // Audio preview stream URL
	SpotifyURL    string `json:"spotify_url,omitempty"`    // Spotify track link
	AppleMusicURL string `json:"apple_music_url,omitempty"`// Apple Music track link
	Source        string `json:"source,omitempty"`         // Resolution provider
}
```

---

## 🚀 Running the HTTP Server & REST API

### 1. Running Locally

Run directly from source:
```bash
go run ./cmd/server
```

Or build a standalone binary:
```bash
go build -o server ./cmd/server
./server
```

The server defaults to port `8080` (accessible at `http://localhost:8080`). If port 8080 is already occupied by another service (like Apache or XAMPP), it automatically falls back to port `8088`.

---

### 2. Deploying to Vercel Serverless

The repository comes pre-configured with `vercel.json` and `api/index.go` targeting the **`@vercel/go`** serverless builder:

```bash
# 1. Install Vercel CLI if not already installed
npm i -g vercel

# 2. Deploy preview build
vercel

# 3. Deploy to production
vercel --prod
```

---

### 3. API Endpoints Summary

Both local HTTP server and Vercel Serverless Function expose the following unified routes:

| HTTP Method | Route | Vercel Alias | Description |
| :--- | :--- | :--- | :--- |
| `GET / POST` | `/metadata` | `/api/metadata` | Extract full metadata (Auto-detects Spotify or Apple Music) |
| `GET / POST` | `/spotify` | `/api/spotify` | Extract metadata strictly for Spotify URLs |
| `GET / POST` | `/apple-music` | `/api/apple-music` | Extract metadata strictly for Apple Music URLs |
| `POST` | `/getInfoMusicV3` | `/api/getInfoMusicV3` | **Locket Dio Ecosystem Standard**: Formatted envelope for mobile / client apps |
| `GET` | `/health` | `/api/health` | Healthcheck endpoint (`{"status":"ok"}`) |
| `GET` | `/` | `/api` | Server info and usage banner |

---

### 4. Endpoint Details & cURL Examples

#### 🔹 1. Endpoint `/metadata` (Recommended)
Fetches metadata, audio preview, artwork, and ISRC.

**Example 1: GET Request with Query Parameter**
```bash
curl "http://localhost:8080/metadata?url=https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT"
```

**Example 2: POST Request with JSON Body**
```bash
curl -X POST "http://localhost:8080/metadata" \
     -H "Content-Type: application/json" \
     -d '{"url": "https://music.apple.com/vn/album/shape-of-you/1201597002?i=1201597040"}'
```

**Sample Response (`200 OK`):**
```json
{
  "song_title": "Never Gonna Give You Up",
  "artist": "Rick Astley",
  "spotify_url": "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
  "apple_music_url": "https://music.apple.com/us/album/never-gonna-give-you-up/...",
  "preview_url": "https://p.scdn.co/mp3-preview/...",
  "image_url": "https://i.scdn.co/image/...",
  "isrc": "GBARL8700021",
  "source": "Spotify"
}
```

---

#### 🔹 2. Endpoint `/getInfoMusicV3` (Locket Dio App Compatible)
Compatible with Locket Dio mobile/web clients, accepting `{ "url": "...", "platform": "spotify|apple" }`.

**POST Request:**
```bash
curl -X POST "http://localhost:8080/getInfoMusicV3" \
     -H "Content-Type: application/json" \
     -d '{
       "url": "https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT",
       "platform": "spotify"
     }'
```

**Sample Response:**
```json
{
  "status": "success",
  "message": "ok",
  "data": {
    "title": "Never Gonna Give You Up",
    "song_name": "Never Gonna Give You Up",
    "song_title": "Never Gonna Give You Up",
    "artist": "Rick Astley",
    "album": null,
    "image_url": "https://i.scdn.co/image/...",
    "preview_url": "https://p.scdn.co/mp3-preview/...",
    "isrc": "GBARL8700021",
    "spotify_url": "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
    "apple_music_url": null,
    "platform": "spotify"
  }
}
```

---

#### 🔹 3. Endpoint `/health`
Used by Docker, Kubernetes, or load balancers for liveness/readiness probes:
```bash
curl "http://localhost:8080/health"
```
**Response:**
```json
{
  "status": "ok",
  "time": "ok"
}
```

---

## ⚙️ Environment Variables

Configure these variables locally, in `.env`, or within your Vercel Project Settings:

| Variable | Default | Description |
| :--- | :--- | :--- |
| `PORT` | `8080` | Port for the local HTTP server listener. |
| `PROXY_URL` | *(empty)* | Optional HTTP/HTTPS/SOCKS5 proxy (e.g. `http://user:pass@1.2.3.4:8080`) applied to all outgoing outbound requests. |

---

## 📂 Project Structure

```
get-meta-music/
├── client.go              # Client struct, Functional Options & Context-aware methods
├── metamusic.go           # Top-level API facade: Get, GetSpotify, GetAppleMusic, GetISRC
├── models.go              # Exported models & Type aliases (SongMetadata, ISRCData)
├── metamusic_test.go      # Unit tests for metamusic package
├── go.mod                 # Module name: github.com/doi2523/get-meta-music
├── LICENSE                # MIT Open Source License
├── README.md              # English documentation (Default)
├── README_VI.md           # Vietnamese documentation
├── vercel.json            # Vercel serverless configuration
├── api/
│   └── index.go           # Vercel Serverless Function entrypoint
├── cmd/
│   └── server/
│       └── main.go        # Standalone HTTP Server binary entrypoint
└── pkg/
    ├── handler/           # HTTP handlers, CORS, Routing, V3 adapter
    │   └── handler.go
    └── music/             # Core scraping logic, ISRC lookup & metadata parsers
        ├── isrc.go
        ├── logger.go
        ├── metadata.go
        └── models.go
```

---

## 📄 License

This project is licensed under the [MIT License](LICENSE).
