package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/doi2523/get-music-metadata/pkg/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.Handler)

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Printf("⚠️ Port %s đã bị chiếm bởi ứng dụng khác (Apache/XAMPP/IIS). Tự động đổi sang port 8088...", port)
		port = "8088"
		listener, err = net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("❌ Không thể mở cổng HTTP server: %v", err)
		}
	}

	log.Printf("🚀 Server music metadata đang chạy tại: http://localhost:%s", port)
	log.Printf("📌 Ví dụ GET: http://localhost:%s/metadata?url=https://open.spotify.com/intl-vi/track/4cOdK2wGLETKBW3PvgPWqT?autoplay_ok=1", port)
	log.Printf("📌 Ví dụ GET: http://localhost:%s/metadata?url=https://music.apple.com/us/album/shape-of-you/1193701079?i=1193701392", port)

	server := &http.Server{Handler: mux}
	if err := server.Serve(listener); err != nil {
		log.Fatalf("❌ Lỗi khởi động server: %v", err)
	}
}
