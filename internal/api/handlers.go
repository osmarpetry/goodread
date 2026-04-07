package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"media2goodreads/internal/export"
	"media2goodreads/internal/ingest"
	"media2goodreads/internal/model"
	"media2goodreads/internal/util"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleLibrary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	lib, err := util.LoadLibrary(s.fs, s.libraryPath)
	if err != nil {
		writeJSON(w, http.StatusOK, model.Library{Items: []model.BookItem{}})
		return
	}
	writeJSON(w, http.StatusOK, lib)
}

// Stats holds library statistics.
type Stats struct {
	Total      int            `json:"total"`
	ByFormat   map[string]int `json:"byFormat"`
	BySource   map[string]int `json:"bySource"`
	WithRating int            `json:"withRating"`
	Read       int            `json:"read"`
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	lib, err := util.LoadLibrary(s.fs, s.libraryPath)
	if err != nil {
		writeJSON(w, http.StatusOK, Stats{
			Total:    0,
			ByFormat: map[string]int{},
			BySource: map[string]int{},
		})
		return
	}

	stats := Stats{
		Total:    len(lib.Items),
		ByFormat: map[string]int{},
		BySource: map[string]int{},
	}

	for _, item := range lib.Items {
		if item.Format != "" {
			stats.ByFormat[item.Format]++
		}
		for _, src := range item.Source {
			stats.BySource[src]++
		}
		if item.Rating != nil && *item.Rating > 0 {
			stats.WithRating++
		}
		if item.FinishedAt != nil {
			stats.Read++
		}
	}

	writeJSON(w, http.StatusOK, stats)
}

// ImportResponse is the response for import operations.
type ImportResponse struct {
	Total   int    `json:"total"`
	Added   int    `json:"added"`
	Merged  int    `json:"merged"`
	Message string `json:"message"`
}

func saveUploadedFile(r *http.Request, fieldName string) (string, func(), error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", nil, fmt.Errorf("missing file field %q: %w", fieldName, err)
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	tmp, err := os.CreateTemp("", "m2gr-*"+ext)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return "", nil, fmt.Errorf("failed to write temp file: %w", err)
	}
	tmp.Close()

	return tmp.Name(), func() { os.Remove(tmp.Name()) }, nil
}

func (s *Server) handleImportAudible(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	format := r.FormValue("format")
	if format == "" {
		format = "json"
	}
	// region is accepted for future region-aware parsing
	_ = r.FormValue("region")

	tmpPath, cleanup, err := saveUploadedFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cleanup()

	var items []model.BookItem
	if format == "csv" {
		items, err = ingest.ParseAudibleCSV(s.fs, tmpPath, s.timezone)
	} else {
		items, err = ingest.ParseAudibleJSON(s.fs, tmpPath, s.timezone)
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse Audible export: %v", err))
		return
	}

	existingLib, _ := util.LoadLibrary(s.fs, s.libraryPath)
	merged, err := ingest.MergeIntoLibrary(s.fs, s.libraryPath, items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats := ingest.CalculateImportStats(existingLib, merged, len(items))
	writeJSON(w, http.StatusOK, ImportResponse{
		Total:   stats.Total,
		Added:   stats.Added,
		Merged:  stats.Merged,
		Message: fmt.Sprintf("Imported %d books (%d new, %d merged)", stats.Total, stats.Added, stats.Merged),
	})
}

func (s *Server) handleImportKindle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	tmpPath, cleanup, err := saveUploadedFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cleanup()

	items, err := ingest.ParseKindleClippings(s.fs, tmpPath, s.timezone)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse Kindle clippings: %v", err))
		return
	}

	existingLib, _ := util.LoadLibrary(s.fs, s.libraryPath)
	merged, err := ingest.MergeIntoLibrary(s.fs, s.libraryPath, items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats := ingest.CalculateImportStats(existingLib, merged, len(items))
	writeJSON(w, http.StatusOK, ImportResponse{
		Total:   stats.Total,
		Added:   stats.Added,
		Merged:  stats.Merged,
		Message: fmt.Sprintf("Imported %d books (%d new, %d merged)", stats.Total, stats.Added, stats.Merged),
	})
}

func (s *Server) handleImportStoritel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	format := r.FormValue("format")
	if format == "" {
		format = "csv"
	}

	tmpPath, cleanup, err := saveUploadedFile(r, "file")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer cleanup()

	var (
		items []model.BookItem
		err2  error
	)
	if format == "json" {
		items, err2 = ingest.ParseStorytelJSON(s.fs, tmpPath, s.timezone)
	} else {
		items, err2 = ingest.ParseStorytelCSV(s.fs, tmpPath, s.timezone)
	}
	if err2 != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse Storytel export: %v", err2))
		return
	}

	existingLib, _ := util.LoadLibrary(s.fs, s.libraryPath)
	merged, err := ingest.MergeIntoLibrary(s.fs, s.libraryPath, items)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	stats := ingest.CalculateImportStats(existingLib, merged, len(items))
	writeJSON(w, http.StatusOK, ImportResponse{
		Total:   stats.Total,
		Added:   stats.Added,
		Merged:  stats.Merged,
		Message: fmt.Sprintf("Imported %d books (%d new, %d merged)", stats.Total, stats.Added, stats.Merged),
	})
}

// ExportRequest is the request body for the export endpoint.
type ExportRequest struct {
	Shelf     string `json:"shelf"`
	DateAdded string `json:"dateAdded"`
}

func (s *Server) handleExportGoodreads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Shelf == "" {
		req.Shelf = "read"
	}
	if req.DateAdded == "" {
		req.DateAdded = time.Now().Format("2006-01-02")
	}

	lib, err := util.LoadLibrary(s.fs, s.libraryPath)
	if err != nil {
		writeError(w, http.StatusNotFound, "library not found; import books first")
		return
	}

	tmp, err := os.CreateTemp("", "goodreads-*.csv")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp file")
		return
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	opts := export.ExportOptions{
		Shelf:     req.Shelf,
		DateAdded: req.DateAdded,
		Timezone:  s.timezone,
	}
	if err := export.ExportGoodreadsCSV(s.fs, lib, tmp.Name(), opts); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to export CSV: %v", err))
		return
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read export file")
		return
	}

	var buf bytes.Buffer
	buf.Write(data)

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=goodreads_import.csv")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
