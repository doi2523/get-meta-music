# 🎵 Get Meta Music (`get-meta-music`)

[![Go CI](https://github.com/doi2523/get-meta-music/actions/workflows/go.yml/badge.svg)](https://github.com/doi2523/get-meta-music/actions/workflows/go.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/doi2523/get-meta-music.svg)](https://pkg.go.dev/github.com/doi2523/get-meta-music)
[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

🌐 **Language**: [English](README.md) | **Tiếng Việt**

**`get-meta-music`** là bộ giải pháp toàn diện bằng Golang dùng để trích xuất metadata âm nhạc từ **Spotify** và **Apple Music**:
- 🎧 **Audio Preview**: Lấy link nghe thử 30s chất lượng cao (MP3/AAC).
- 🖼 **Artwork**: Ảnh bìa bài hát chất lượng gốc (HD 600x600px).
- 🆔 **ISRC Lookup**: Tra cứu mã định danh âm nhạc quốc tế (ISRC) và liên kết chéo giữa Spotify ⇄ Apple Music.
- ⚡ **Dual-Mode**: Vừa là **Go Package (Thư viện/SDK)** để import trực tiếp vào dự án Go khác, vừa là **Standalone HTTP REST API Server** & **Vercel Serverless Function**.

---

## 📑 Mục lục

1. [Cài đặt & Tải về](#-cài-đặt--tải-về)
2. [Sử dụng như Thư viện Go (Go Package SDK)](#-sử-dụng-như-thư-viện-go-go-package-sdk)
   - [Gọi nhanh qua hàm tiện ích](#1-gọi-nhanh-qua-hàm-tiện-ích-quickstart)
   - [Dùng Client với tùy biến (Timeout, Proxy, Silent, Logger)](#2-khởi-tạo-client-với-tùy-biến-nâng-cao)
   - [Hỗ trợ Context & Timeout](#3-hỗ-trợ-contextcontext-hủy-request)
   - [Cấu trúc dữ liệu trả về](#4-cấu-trúc-dữ-liệu-model)
3. [Chạy Server & Sử dụng qua REST API](#-chạy-server--sử-dụng-qua-rest-api)
   - [Chạy Server Local](#1-khởi-chạy-server-local)
   - [Deploy lên Vercel](#2-deploy-lên-vercel-serverless)
   - [Bảng tổng hợp API Endpoints](#3-bảng-tổng-hợp-api-endpoints)
   - [Chi tiết từng Endpoint & Mẫu cURL](#4-chi-tiết-từng-endpoint--ví-dụ-curl)
4. [Biến môi trường (Environment Variables)](#-biến-môi-trường-environment-variables)
5. [Cấu trúc dự án](#-cấu-trúc-dự-án)

---

## 📦 Cài đặt & Tải về

### Cách 1: Cài đặt làm Go Module cho dự án Go của bạn
Chạy lệnh sau trong thư mục dự án Go của bạn:
```bash
go get github.com/doi2523/get-meta-music
```

### Cách 2: Clone mã nguồn về máy để chạy Server hoặc đóng góp code
```bash
git clone https://github.com/doi2523/get-meta-music.git
cd get-meta-music
```

---

## 💻 Sử dụng như Thư viện Go (Go Package SDK)

Import thư viện vào code Go của bạn:
```go
import metamusic "github.com/doi2523/get-meta-music"
```

### 1. Gọi nhanh qua hàm tiện ích (Quickstart)

```go
package main

import (
	"fmt"
	"log"

	metamusic "github.com/doi2523/get-meta-music"
)

func main() {
	// Tự động nhận diện link (Spotify hoặc Apple Music)
	url := "https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT"
	
	meta, err := metamusic.Get(url)
	if err != nil {
		log.Fatalf("Lỗi: %v", err)
	}

	fmt.Printf("🎵 Tên bài hát: %s\n", meta.SongTitle)
	fmt.Printf("👤 Nghệ sĩ:    %s\n", meta.Artist)
	fmt.Printf("💿 Album:      %s\n", meta.Album)
	fmt.Printf("🎧 Audio Demo: %s\n", meta.PreviewURL)
	fmt.Printf("🖼 Ảnh bìa:    %s\n", meta.ImageURL)
	fmt.Printf("🆔 Mã ISRC:    %s\n", meta.ISRC)
	fmt.Printf("🍏 Apple URL:  %s\n", meta.AppleMusicURL)
	fmt.Printf("🟢 Spotify URL:%s\n", meta.SpotifyURL)
}
```

Các hàm tiện ích có sẵn:
- `metamusic.Get(url)`: Tự động nhận diện link Spotify hoặc Apple Music.
- `metamusic.GetSpotify(url)`: Chỉ xử lý link Spotify.
- `metamusic.GetAppleMusic(url)`: Chỉ xử lý link Apple Music.
- `metamusic.GetISRC(url)`: Lấy mã ISRC và tìm chéo link nền tảng còn lại.
- `metamusic.SetSilent(true)`: Tắt toàn bộ log internal ra stdout.

---

### 2. Khởi tạo `Client` với tùy biến nâng cao

Trong ứng dụng backend hoặc microservice production, bạn nên khởi tạo `Client` để cấu hình Timeout, Proxy (nếu cần vượt giới hạn rate limit hoặc IP bị chặn), và chế độ log:

```go
package main

import (
	"fmt"
	"time"

	metamusic "github.com/doi2523/get-meta-music"
)

func main() {
	// Khởi tạo client tùy chỉnh
	client := metamusic.NewClient(
		metamusic.WithTimeout(10*time.Second),                    // Timeout 10 giây
		metamusic.WithSilent(),                                   // Tắt hoàn toàn log ra stdout
		// metamusic.WithProxy("http://username:pass@proxy-ip:port"), // Cấu hình Proxy (nếu có)
	)

	// Lấy thông tin bài hát từ Apple Music
	appleURL := "https://music.apple.com/vn/album/shape-of-you/1201597002?i=1201597040"
	meta, err := client.GetAppleMusic(appleURL)
	if err != nil {
		fmt.Println("Lỗi:", err)
		return
	}

	fmt.Printf("Bài hát: %s - %s\nISRC: %s\nPreview: %s\n",
		meta.SongTitle, meta.Artist, meta.ISRC, meta.PreviewURL)
}
```

Các Options hỗ trợ cho `NewClient`:
- `WithTimeout(d time.Duration)`: Thời gian chờ tối đa cho các request HTTP.
- `WithProxy(proxyURL string)`: Hỗ trợ proxy HTTP/HTTPS/SOCKS5.
- `WithSilent()`: Không in log ra terminal.
- `WithLogger(l music.Logger)`: Thay thế bằng logger tùy chỉnh (Zap, Logrus, Slog...).
- `WithHTTPClient(client *http.Client)`: Dùng chung pool HTTP Client sẵn có của ứng dụng.

---

### 3. Hỗ trợ `context.Context` (Hủy request)

Tất cả phương thức của `Client` đều có phiên bản hỗ trợ Context:
- `client.GetWithContext(ctx, url)`
- `client.GetSpotifyWithContext(ctx, url)`
- `client.GetAppleMusicWithContext(ctx, url)`
- `client.GetISRCWithContext(ctx, url)`

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

meta, err := client.GetWithContext(ctx, "https://open.spotify.com/track/...")
if err != nil {
    // Sẽ trả về context.DeadlineExceeded hoặc context.Canceled nếu hết thời gian
    fmt.Println("Request bị gián đoạn:", err)
}
```

---

### 4. Cấu trúc dữ liệu Model

#### `SongMetadata`
```go
type SongMetadata struct {
	SongTitle     string `json:"song_title,omitempty"`     // Tên bài hát
	Artist        string `json:"artist,omitempty"`         // Tên ca sĩ / nghệ sĩ
	Album         string `json:"album,omitempty"`          // Tên Album
	SpotifyURL    string `json:"spotify_url,omitempty"`    // Link gốc hoặc link tương ứng trên Spotify
	AppleMusicURL string `json:"apple_music_url,omitempty"`// Link gốc hoặc link tương ứng trên Apple Music
	PreviewURL    string `json:"preview_url,omitempty"`    // Link audio nghe thử MP3/M4A 30 giây
	ImageURL      string `json:"image_url,omitempty"`      // Link ảnh bìa chất lượng cao
	ISRC          string `json:"isrc,omitempty"`           // Mã định danh quốc tế (ISRC)
	Source        string `json:"source,omitempty"`         // Nguồn dữ liệu (Spotify / Apple Music / Deezer...)
}
```

#### `ISRCData`
```go
type ISRCData struct {
	ISRC          string `json:"isrc"`                     // Mã ISRC
	Title         string `json:"title,omitempty"`          // Tên bài hát tra cứu từ ISRC
	Artist        string `json:"artist,omitempty"`         // Nghệ sĩ
	PreviewURL    string `json:"preview_url,omitempty"`    // Link audio demo
	SpotifyURL    string `json:"spotify_url,omitempty"`    // Link Spotify của bài hát
	AppleMusicURL string `json:"apple_music_url,omitempty"`// Link Apple Music của bài hát
	Source        string `json:"source,omitempty"`         // Nguồn phân giải ISRC
}
```

---

## 🚀 Chạy Server & Sử dụng qua REST API

### 1. Khởi chạy Server Local

Chạy trực tiếp từ mã nguồn:
```bash
go run ./cmd/server
```

Hoặc build ra file binary thực thi:
```bash
go build -o server ./cmd/server
./server
```

Mặc định server sẽ chạy tại cổng `8080` (URL: `http://localhost:8080`). Nếu cổng 8080 đang bị phần mềm khác sử dụng, server sẽ tự động chuyển sang cổng `8088`.

---

### 2. Deploy lên Vercel Serverless

Dự án đã có sẵn file `vercel.json` và cấu trúc entrypoint `api/index.go` tương thích chuẩn với **Vercel Go Runtime**:

```bash
# 1. Cài đặt Vercel CLI nếu chưa có
npm i -g vercel

# 2. Deploy thử nghiệm
vercel

# 3. Deploy Production
vercel --prod
```

---

### 3. Bảng tổng hợp API Endpoints

Cả bản chạy Local Server lẫn Vercel Serverless Function đều hỗ trợ các đường dẫn sau:

| Phương thức | Đường dẫn (Route) | Alias Vercel | Mô tả |
| :--- | :--- | :--- | :--- |
| `GET / POST` | `/metadata` | `/api/metadata` | Lấy đầy đủ metadata (Tự động nhận diện Spotify hoặc Apple Music) |
| `GET / POST` | `/spotify` | `/api/spotify` | Chỉ xử lý liên kết Spotify |
| `GET / POST` | `/apple-music` | `/api/apple-music` | Chỉ xử lý liên kết Apple Music |
| `POST` | `/getInfoMusicV3` | `/api/getInfoMusicV3` | **Chuẩn Locket Dio**: Trả về dữ liệu đóng gói chuẩn cho mobile app / web client |
| `GET` | `/health` | `/api/health` | Kiểm tra tình trạng hoạt động của Server (Healthcheck) |
| `GET` | `/` | `/api` | Trang thông tin server & hướng dẫn API |

---

### 4. Chi tiết từng Endpoint & Ví dụ cURL

#### 🔹 1. Endpoint `/metadata` (Khuyên dùng)
Trích xuất toàn bộ metadata, audio preview, ảnh bìa, mã ISRC.

**Cách 1: Gửi GET Request với Query Parameter**
```bash
curl "http://localhost:8080/metadata?url=https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT"
```

**Cách 2: Gửi POST Request với JSON Body**
```bash
curl -X POST "http://localhost:8080/metadata" \
     -H "Content-Type: application/json" \
     -d '{"url": "https://music.apple.com/vn/album/shape-of-you/1201597002?i=1201597040"}'
```

**JSON Response mẫu (HTTP 200 OK):**
```json
{
  "song_title": "Never Gonna Give You Up",
  "artist": "Rick Astley",
  "spotify_url": "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
  "apple_music_url": "https://music.apple.com/us/album/never-won-give-you-up/...",
  "preview_url": "https://p.scdn.co/mp3-preview/...",
  "image_url": "https://i.scdn.co/image/...",
  "isrc": "GBARL8700021",
  "source": "Spotify"
}
```

---

#### 🔹 2. Endpoint `/getInfoMusicV3` (Tương thích App Locket Dio)
Được tối ưu cho hệ thống Locket Dio, nhận format yêu cầu đặc thù và trả về cấu trúc `{ status, message, data }`.

**Gửi POST Request:**
```bash
curl -X POST "http://localhost:8080/getInfoMusicV3" \
     -H "Content-Type: application/json" \
     -d '{
       "url": "https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT",
       "platform": "spotify"
     }'
```

**JSON Response mẫu:**
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
Dùng cho load balancer hoặc Kubernetes / Docker kiểm tra sống còn:
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

## ⚙️ Biến môi trường (Environment Variables)

Khi khởi chạy server trên Server/VPS hoặc thiết lập trên Vercel Environment Variables:

| Biến | Mặc định | Ý nghĩa |
| :--- | :--- | :--- |
| `PORT` | `8080` | Cổng lắng nghe của HTTP Server. |
| `PROXY_URL` | *(trống)* | Địa chỉ Proxy HTTP/HTTPS/SOCKS5 (ví dụ: `http://user:pass@1.2.3.4:8080`) dùng cho mọi request ra ngoài. |

---

## 📂 Cấu trúc dự án

```
get-meta-music/
├── client.go              # Client struct, Functional Options & Context-aware methods
├── metamusic.go           # Top-level API facade: Get, GetSpotify, GetAppleMusic, GetISRC
├── models.go              # Exported models & Type aliases (SongMetadata, ISRCData)
├── metamusic_test.go      # Unit tests cho package metamusic
├── go.mod                 # Module name: github.com/doi2523/get-meta-music
├── LICENSE                # Giấy phép mã nguồn mở MIT
├── README.md              # Tài liệu hướng dẫn sử dụng chi tiết
├── vercel.json            # Cấu hình Serverless deployment trên Vercel
├── api/
│   └── index.go           # Entry point cho Vercel Serverless Function
├── cmd/
│   └── server/
│       └── main.go        # Entry point cho Standalone HTTP Server Local / Production
└── pkg/
    ├── handler/           # HTTP handlers, CORS, Routing, V3 adapter
    │   └── handler.go
    └── music/             # Core logic trích xuất metadata, phân giải ISRC & scrape
        ├── isrc.go
        ├── logger.go
        ├── metadata.go
        └── models.go
```

---

## 📄 License

Mã nguồn được phân phối dưới giấy phép [MIT](LICENSE).
