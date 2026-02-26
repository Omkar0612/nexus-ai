// Package webui provides the NEXUS embedded Web UI HTTP server.
package webui

import (
	"embed"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Omkar0612/nexus-ai/internal/router"
	"github.com/rs/zerolog"
)

//go:embed static/*
var staticFS embed.FS

// Server is the NEXUS Web UI HTTP server.
type Server struct {
	addr           string
	hub            *SSEHub
	logger         zerolog.Logger
	router         *router.Router
	allowedOrigins []string
}

// New creates a Server.
func New(addr string, log zerolog.Logger, r *router.Router) *Server {
	origins := []string{"http://localhost", "http://127.0.0.1"}
	if v := os.Getenv("NEXUS_ALLOWED_ORIGINS"); v != "" {
		origins = strings.Split(v, ",")
	}
	return &Server{
		addr:           addr,
		hub:            newSSEHub(),
		logger:         log,
		router:         r,
		allowedOrigins: origins,
	}
}

// Hub returns the SSE hub so agents can publish activity events.
func (s *Server) Hub() *SSEHub { return s.hub }

// Start registers all routes and blocks serving HTTP.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Static UI
	sub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /", http.FileServer(http.FS(sub)))

	// API routes
	mux.HandleFunc("POST /api/chat", s.handleChat)
	mux.HandleFunc("GET /api/events", s.hub.ServeHTTP)
	mux.HandleFunc("GET /api/health", s.handleHealth)

	// Wrap entire mux with security middleware chain:
	// securityHeaders → corsMiddleware → rateLimiter → mux
	handler := s.securityHeaders(s.corsMiddleware(rateLimiter(mux)))

	s.logger.Info().Str("addr", s.addr).Msg("[webui] server started")
	return http.ListenAndServe(s.addr, handler)
}

// handleHealth returns server status.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","version":"1.7"}`)) //nolint:errcheck
}

// --- Security middleware ---

// securityHeaders adds hardening headers to every response.
func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware enforces a strict origin allowlist.
// Access-Control-Allow-Origin: * is intentionally NOT used.
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			allowed := false
			for _, o := range s.allowedOrigins {
				if strings.HasPrefix(origin, strings.TrimSpace(o)) {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			} else if r.Method == http.MethodOptions {
				http.Error(w, "CORS: origin not allowed", http.StatusForbidden)
				return
			}
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// --- Rate limiter (token bucket, 100 req/min per IP) ---

const (
	rateLimitPerMin = 100
	rateLimitWindow = time.Minute
)

type ipBucket struct {
	count     int
	windowEnd time.Time
}

var (
	rateMu      sync.Mutex
	rateBuckets sync.Map // map[string]*ipBucket
)

func rateLimiter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := realIP(r)
		now := time.Now()

		rateMu.Lock()
		raw, _ := rateBuckets.LoadOrStore(ip, &ipBucket{windowEnd: now.Add(rateLimitWindow)})
		bucket := raw.(*ipBucket)
		if now.After(bucket.windowEnd) {
			bucket.count = 0
			bucket.windowEnd = now.Add(rateLimitWindow)
		}
		bucket.count++
		count := bucket.count
		rateMu.Unlock()

		if count > rateLimitPerMin {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// realIP extracts the client IP, honouring X-Forwarded-For when behind a proxy.
// Falls back to RemoteAddr.
func realIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first (leftmost) IP — closest to the real client.
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
