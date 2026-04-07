package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

// Server handles HTTP requests for the web interface.
type Server struct {
	fs          afero.Fs
	libraryPath string
	outDir      string
	timezone    string
	staticDir   string
	mux         *http.ServeMux
}

// NewServer creates a new Server.
func NewServer(fs afero.Fs, outDir, timezone, staticDir string) *Server {
	s := &Server{
		fs:          fs,
		libraryPath: filepath.Join(outDir, "library.json"),
		outDir:      outDir,
		timezone:    timezone,
		staticDir:   staticDir,
	}
	s.mux = http.NewServeMux()
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/health", s.handleHealth)
	s.mux.HandleFunc("/api/library", s.handleLibrary)
	s.mux.HandleFunc("/api/stats", s.handleStats)
	s.mux.HandleFunc("/api/import/audible", s.handleImportAudible)
	s.mux.HandleFunc("/api/import/kindle", s.handleImportKindle)
	s.mux.HandleFunc("/api/import/storytel", s.handleImportStoritel)
	s.mux.HandleFunc("/api/export/goodreads", s.handleExportGoodreads)

	if s.staticDir != "" {
		s.mux.HandleFunc("/", s.handleStatic)
	}
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		http.NotFound(w, r)
		return
	}

	path := filepath.Join(s.staticDir, filepath.Clean(r.URL.Path))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(s.staticDir, "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
}

// Run starts the HTTP server on the given address.
func Run(addr, outDir, timezone, staticDir string) error {
	fsys := afero.NewOsFs()
	srv := NewServer(fsys, outDir, timezone, staticDir)
	return http.ListenAndServe(addr, srv)
}
