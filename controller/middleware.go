package controller

import (
	"github.com/integration-system/isp-lib/logger"
	"github.com/integration-system/isp-lib/utils"
	"net/http"
	"strings"
)

func HandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		if utils.DEV {
			logger.Debug(r.RequestURI)
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func StaticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var contentType string
		currentPath := r.URL.Path[1:]
		if strings.HasSuffix(currentPath, ".json") || r.Header.Get("Content-Type") == "application/json" {
			contentType = "application/json"
		} else if strings.HasSuffix(currentPath, ".css") {
			contentType = "text/css"
		} else if currentPath == "swagger/" {
			contentType = "text/html"
		} else if strings.HasSuffix(currentPath, ".html") {
			contentType = "text/html"
		} else if strings.HasSuffix(currentPath, ".js") {
			contentType = "application/javascript"
		} else if strings.HasSuffix(currentPath, ".png") {
			contentType = "image/png"
		} else if strings.HasSuffix(currentPath, ".svg") {
			contentType = "image/svg+xml"
		} else {
			contentType = "text/plain"
		}
		w.Header().Add("Content-Type", contentType)
		next.ServeHTTP(w, r)
	})
}
