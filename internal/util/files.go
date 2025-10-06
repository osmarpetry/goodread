package util

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"media2goodreads/internal/model"

	"github.com/spf13/afero"
)

// LoadLibrary loads a library from a JSON file.
func LoadLibrary(fs afero.Fs, path string) (*model.Library, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read library: %w", err)
	}

	var lib model.Library
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, fmt.Errorf("failed to parse library JSON: %w", err)
	}

	return &lib, nil
}

// SaveLibrary saves a library to a JSON file with pretty printing.
func SaveLibrary(fs afero.Fs, path string, lib *model.Library) error {
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal library: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := afero.WriteFile(fs, path, data, 0644); err != nil {
		return fmt.Errorf("failed to write library: %w", err)
	}

	return nil
}

// MergeLibraries merges new items into an existing library.
func MergeLibraries(existing *model.Library, newItems []model.BookItem) *model.Library {
	if existing == nil {
		existing = &model.Library{}
	}

	// Combine all items
	allItems := append(existing.Items, newItems...)

	// Deduplicate
	deduped := DedupeLibrary(allItems)

	return &model.Library{Items: deduped}
}

// EnsureDir ensures a directory exists.
func EnsureDir(fs afero.Fs, path string) error {
	return fs.MkdirAll(path, 0755)
}

// OpenDirectory opens a directory in the system file explorer.
func OpenDirectory(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", path)
	default: // linux, bsd, etc.
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

// FileExists checks if a file exists.
func FileExists(fs afero.Fs, path string) bool {
	_, err := fs.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
