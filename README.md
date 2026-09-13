# Golang Music Metadata Server 🎵 (Support Vercel Deployment)

Server Golang REST API lấy thông tin bài hát (Spotify & Apple Music) bằng cách trích xuất Metadata và tra cứu mã ISRC.

Sẵn sàng chạy dưới dạng **Serverless Function trên Vercel** hoặc **Local Go HTTP Server**.

---

## 🛠 Cấu trúc dự án Vercel

```
get-meta-music/
├── go.mod            # Cấu hình Go module
├── vercel.json       # Cấu hình Deploy Vercel (@vercel/go builder)
├── api/
│   └── index.go      # Vercel Serverless Function entrypoint
├── cmd/
│   └── server/
│       └── main.go   # Server HTTP local (go run ./cmd/server)
├── pkg/
│   ├── handler/      # Shared HTTP handlers & CORS routing
│   │   └── handler.go
│   └── music/        # Logic trích xuất metadata & ISRC
│       ├── models.go
│       ├── isrc.go
│       └── metadata.go
└── README.md         # Hướng dẫn sử dụng
```

---

## 🚀 Hướng dẫn Deploy lên Vercel

### Cách 1: Deploy qua Vercel CLI (Nhanh nhất)

1. Cài đặt Vercel CLI (nếu chưa có):
   ```bash
   npm i -g vercel
   ```

2. Đăng nhập Vercel:
   ```bash
   vercel login
   ```

3. Chạy lệnh Deploy tại thư mục dự án `spotify`:
   ```bash
   vercel
   ```

4. Khi deploy Production:
   ```bash
   vercel --prod
   ```

---

### Cách 2: Deploy qua GitHub / GitLab / Bitbucket

1. Push mã nguồn dự án `spotify` lên repository GitHub/GitLab của bạn.
2. Truy cập [Vercel Dashboard](https://vercel.com/dashboard) -> Chọn **Add New Project**.
3. Import Repository của bạn.
4. Nhấn **Deploy** (Vercel tự động nhận diện `vercel.json` và biên dịch Go Serverless Function).

---

## 💻 Chạy dưới máy Local (Development)

Chạy lệnh sau tại thư mục dự án:

```bash
go run ./cmd/server
```

Mặc định server khởi chạy tại `http://localhost:8080` (hoặc tự động đổi sang `8088` nếu 8080 bị bận).

---

## 📡 Chi tiết API Endpoints (Vercel & Local)

- **GET / POST** `/metadata` (hoặc `/api/metadata`)
- **GET / POST** `/spotify` (hoặc `/api/spotify`)
- **GET / POST** `/apple-music` (hoặc `/api/apple-music`)
- **POST** `/getInfoMusicV3` (hoặc `/api/getInfoMusicV3`) với body `{"url": "...", "platform": "apple|spotify"}`

#### Ví dụ Request (GET):
```bash
curl "https://your-project.vercel.app/metadata?url=https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT?autoplay_ok=1"
```

#### Ví dụ Request (POST):
```bash
curl -X POST "https://your-project.vercel.app/metadata" \
     -H "Content-Type: application/json" \
     -d '{"url": "https://music.apple.com/vn/album/shape-of-you/1201597002?i=1201597040"}'
```

#### Response mẫu (JSON):
```json
{
  "song_title": "Never Gonna Give You Up",
  "artist": "Rick Astley",
  "spotify_url": "https://open.spotify.com/track/4cOdK2wGLETKBW3PvgPWqT",
  "preview_url": "https://p.scdn.co/mp3-preview/...",
  "image_url": "https://i.scdn.co/image/...",
  "isrc": "GBARL8700021",
  "source": "Deezer"
}
```
