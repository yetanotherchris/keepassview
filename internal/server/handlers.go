package server

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"text/template"

	"keepassview/internal/vault"
)

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /app.js", s.handleAppJS)
	mux.HandleFunc("GET /app.css", s.handleAppCSS)
	mux.HandleFunc("GET /api/groups", s.apiGroups)
	mux.HandleFunc("GET /api/search", s.apiSearch)
	mux.HandleFunc("GET /api/entries/{uuid}", s.apiEntry)
	mux.HandleFunc("POST /api/quit", s.apiQuit)
}

// --- static assets ---

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmplBytes, err := fs.ReadFile(s.assets, "assets/templates/index.html.tmpl")
	if err != nil {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}

	type pageData struct {
		IdleTimeoutSeconds int
	}

	tmpl, err := template.New("index").Parse(string(tmplBytes))
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, pageData{
		IdleTimeoutSeconds: s.cfg.IdleTimeoutSeconds,
	})
}

func (s *Server) handleAppJS(w http.ResponseWriter, r *http.Request) {
	tmplBytes, err := fs.ReadFile(s.assets, "assets/templates/app.js.tmpl")
	if err != nil {
		http.Error(w, "asset not found", http.StatusInternalServerError)
		return
	}

	type jsData struct {
		IdleTimeoutSeconds int
	}

	tmpl, err := template.New("app.js").Parse(string(tmplBytes))
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	_ = tmpl.Execute(w, jsData{
		IdleTimeoutSeconds: s.cfg.IdleTimeoutSeconds,
	})
}

func (s *Server) handleAppCSS(w http.ResponseWriter, r *http.Request) {
	data, err := fs.ReadFile(s.assets, "assets/static/app.css")
	if err != nil {
		http.Error(w, "asset not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	_, _ = w.Write(data)
}

// --- API ---

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func (s *Server) apiGroups(w http.ResponseWriter, r *http.Request) {
	s.touchActivity()
	writeJSON(w, s.vault.Groups())
}

func (s *Server) apiSearch(w http.ResponseWriter, r *http.Request) {
	s.touchActivity()
	q := r.URL.Query().Get("q")
	group := r.URL.Query().Get("group")
	results := s.vault.Search(q, group)
	if results == nil {
		results = []vault.EntrySummary{}
	}
	writeJSON(w, results)
}

func (s *Server) apiEntry(w http.ResponseWriter, r *http.Request) {
	s.touchActivity()
	uuid := r.PathValue("uuid")
	detail, err := s.vault.GetEntry(uuid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, detail)
}

func (s *Server) apiQuit(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	go s.triggerShutdown()
}
