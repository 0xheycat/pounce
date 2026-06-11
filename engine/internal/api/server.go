// Package api exposes the download engine over a small REST + Server-Sent
// Events (SSE) HTTP interface. SSE is used instead of WebSockets so the engine
// stays dependency-free (stdlib only).
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/0xheycat/pounce/internal/download"
	"github.com/0xheycat/pounce/internal/model"
	"github.com/0xheycat/pounce/internal/settings"
	"github.com/0xheycat/pounce/internal/sources"
)

// Server wires the download manager and the SSE hub to HTTP routes.
type Server struct {
	mgr       *download.Manager
	hub       *Hub
	settings  *settings.Store
	staticDir string
	authToken string
}

// New creates a Server. staticDir may be empty (API-only mode) and authToken
// may be empty (no authentication).
func New(mgr *download.Manager, hub *Hub, set *settings.Store, staticDir, authToken string) *Server {
	return &Server{mgr: mgr, hub: hub, settings: set, staticDir: staticDir, authToken: authToken}
}

// Handler returns the fully-routed HTTP handler.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/downloads", s.list)
	mux.HandleFunc("POST /api/downloads", s.add)
	mux.HandleFunc("POST /api/downloads/{id}/pause", s.pause)
	mux.HandleFunc("POST /api/downloads/{id}/resume", s.resume)
	mux.HandleFunc("POST /api/downloads/{id}/cancel", s.cancel)
	mux.HandleFunc("POST /api/downloads/{id}/speed", s.speed)
	mux.HandleFunc("DELETE /api/downloads/{id}", s.remove)
	mux.HandleFunc("GET /api/events", s.events)
	mux.HandleFunc("GET /api/settings", s.getSettings)
	mux.HandleFunc("PUT /api/settings", s.putSettings)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	if s.staticDir != "" {
		mux.Handle("/", http.FileServer(http.Dir(s.staticDir)))
	}
	return cors(s.auth(mux))
}

func (s *Server) list(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.mgr.List())
}

type addRequest struct {
	URL         string `json:"url"`
	Dir         string `json:"dir"`
	Connections int    `json:"connections"`
	SpeedLimit  int64  `json:"speedLimit"`
	Ytdlp       bool   `json:"ytdlp"`
}

func (s *Server) add(w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		writeErr(w, http.StatusBadRequest, "a valid \"url\" is required")
		return
	}
	target := req.URL
	if req.Ytdlp {
		media, rerr := (sources.YtDlp{}).Resolve(r.Context(), req.URL)
		if rerr != nil || len(media) == 0 {
			writeErr(w, http.StatusBadGateway, "yt-dlp could not resolve this URL: "+errString(rerr))
			return
		}
		target = media[0].URL
	}
	d, err := s.mgr.Add(target, req.Dir, req.Connections, req.SpeedLimit)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	_ = s.mgr.Start(d.ID)
	writeJSON(w, http.StatusCreated, d)
}

func (s *Server) pause(w http.ResponseWriter, r *http.Request) {
	s.act(w, s.mgr.Pause(r.PathValue("id")))
}

func (s *Server) resume(w http.ResponseWriter, r *http.Request) {
	s.act(w, s.mgr.Start(r.PathValue("id")))
}

func (s *Server) cancel(w http.ResponseWriter, r *http.Request) {
	s.act(w, s.mgr.Remove(r.PathValue("id"), true))
}

func (s *Server) remove(w http.ResponseWriter, r *http.Request) {
	s.act(w, s.mgr.Remove(r.PathValue("id"), false))
}

type speedRequest struct {
	Limit int64 `json:"limit"`
}

func (s *Server) speed(w http.ResponseWriter, r *http.Request) {
	var req speedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	s.act(w, s.mgr.SetSpeed(r.PathValue("id"), req.Limit))
}

func (s *Server) act(w http.ResponseWriter, err error) {
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) getSettings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, s.settings.Get())
}

func (s *Server) putSettings(w http.ResponseWriter, r *http.Request) {
	var next settings.Settings
	if err := json.NewDecoder(r.Body).Decode(&next); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid settings body")
		return
	}
	saved, err := s.settings.Save(next)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, saved)
}

// auth enforces a bearer token on /api routes when one is configured. The SSE
// endpoint also accepts the token via ?token= because EventSource cannot set
// headers. /api/health stays open for readiness checks.
func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.authToken == "" || !strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api/health" {
			next.ServeHTTP(w, r)
			return
		}
		if s.tokenOK(r) {
			next.ServeHTTP(w, r)
			return
		}
		writeErr(w, http.StatusUnauthorized, "missing or invalid token")
	})
}

func (s *Server) tokenOK(r *http.Request) bool {
	if t := r.URL.Query().Get("token"); t != "" && t == s.authToken {
		return true
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ") == s.authToken
	}
	return r.Header.Get("X-Pounce-Token") == s.authToken
}

func errString(err error) string {
	if err == nil {
		return "no media found"
	}
	return err.Error()
}

// events streams live download updates as SSE.
func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.hub.Add()
	defer s.hub.Remove(ch)

	// Send an initial snapshot so a fresh client renders immediately.
	for _, d := range s.mgr.List() {
		writeEvent(w, d)
	}
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case b := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", b)
			flusher.Flush()
		}
	}
}

func writeEvent(w http.ResponseWriter, d *model.Download) {
	b, _ := json.Marshal(map[string]any{"type": "download", "data": d})
	fmt.Fprintf(w, "data: %s\n\n", b)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// cors allows the Vite dev server (and remote dashboards) to call the engine.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Pounce-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
