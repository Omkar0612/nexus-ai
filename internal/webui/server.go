package webui

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/Omkar0612/nexus-ai/internal/router"
	"github.com/rs/zerolog"
)

//go:embed static/*
var staticFS embed.FS

// Server is the NEXUS v1.6 Web UI HTTP server.
type Server struct {
	addr   string
	hub    *SSEHub
	logger zerolog.Logger
	router *router.Router
}

// New creates a Server.
// Pass zerolog.Nop() and nil router in tests.
func New(addr string, log zerolog.Logger, r *router.Router) *Server {
	return &Server{
		addr:   addr,
		hub:    newSSEHub(),
		logger: log,
		router: r,
	}
}

// Hub returns the SSE hub so agents can publish activity events.
func (s *Server) Hub() *SSEHub { return s.hub }

// Start registers all routes and blocks serving HTTP.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Static UI — fully embedded, zero runtime file deps
	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /", http.FileServer(http.FS(sub)))

	// API routes
	mux.HandleFunc("POST /api/chat", s.handleChat)
	mux.HandleFunc("GET /api/events", s.hub.ServeHTTP)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","version":"1.6"}`))
	})

	s.logger.Info().Str("addr", s.addr).Msg("[webui] server started")
	return http.ListenAndServe(s.addr, mux)
}
