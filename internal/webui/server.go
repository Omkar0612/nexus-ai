package webui

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/rs/zerolog"
)

//go:embed static/*
var staticFS embed.FS

// Server is the v1.6 Web UI HTTP server.
type Server struct {
	addr   string
	hub    *SSEHub
	logger zerolog.Logger
}

// New creates a Server. Pass zerolog.Nop() in tests.
func New(addr string, log zerolog.Logger) *Server {
	return &Server{
		addr:   addr,
		hub:    newSSEHub(),
		logger: log,
	}
}

// Hub returns the SSE hub so agents can publish events.
func (s *Server) Hub() *SSEHub { return s.hub }

// Start registers routes and blocks serving HTTP.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Static UI — fully embedded, single-binary deploy
	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /", http.FileServer(http.FS(sub)))

	// API
	mux.HandleFunc("POST /api/chat", s.handleChat)
	mux.HandleFunc("GET /api/events", s.hub.ServeHTTP)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"1.6"}`))
	})

	s.logger.Info().Str("addr", s.addr).Msg("[webui] server started")
	return http.ListenAndServe(s.addr, mux)
}
