package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

func serveStaticFiles(r *chi.Mux, srv *Server) {
	if srv.StaticFiles == nil {
		return
	}
	fileServer := http.FileServer(http.FS(srv.StaticFiles))
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		f, err := srv.StaticFiles.Open(path[1:])
		if err != nil {
			// SPA fallback: serve index.html with nonce injection
			serveIndexWithNonce(w, r, srv)
			return
		}
		f.Close()

		// Serve index.html with nonce injection
		if path == "/index.html" {
			serveIndexWithNonce(w, r, srv)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}

func serveIndexWithNonce(w http.ResponseWriter, r *http.Request, srv *Server) {
	f, err := srv.StaticFiles.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	nonce := getNonce(r.Context())
	html := string(data)
	if nonce != "" {
		html = strings.ReplaceAll(html, "<script", `<script nonce="`+nonce+`"`)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
