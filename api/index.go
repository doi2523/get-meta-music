// Package handler is the Vercel Serverless Function entry point.
// All routing and business logic is delegated to pkg/handler.
package handler

import (
	"net/http"

	apphandler "get-meta-music/pkg/handler"
)

// Handler is the Vercel Serverless Function entry point.
func Handler(w http.ResponseWriter, r *http.Request) {
	apphandler.Handler(w, r)
}
