package ingest

import (
	"fmt"

	"media2goodreads/internal/model"
	"media2goodreads/internal/util"

	"github.com/spf13/afero"
)

// MergeIntoLibrary merges new items into an existing library file.
// If the library doesn't exist, it creates a new one.
func MergeIntoLibrary(fs afero.Fs, libraryPath string, newItems []model.BookItem) (*model.Library, error) {
	// Try to load existing library
	existingLib, err := util.LoadLibrary(fs, libraryPath)
	if err != nil {
		// If file doesn't exist, start with empty library
		existingLib = &model.Library{Items: []model.BookItem{}}
	}

	// Merge items
	merged := util.MergeLibraries(existingLib, newItems)

	// Save merged library
	if err := util.SaveLibrary(fs, libraryPath, merged); err != nil {
		return nil, fmt.Errorf("failed to save library: %w", err)
	}

	return merged, nil
}

// ImportStats contains statistics about an import operation.
type ImportStats struct {
	Total    int
	Added    int
	Merged   int
	Skipped  int
}

// CalculateImportStats calculates statistics for an import operation.
func CalculateImportStats(before, after *model.Library, newCount int) ImportStats {
	beforeCount := 0
	if before != nil {
		beforeCount = len(before.Items)
	}
	afterCount := len(after.Items)
	
	added := afterCount - beforeCount
	if added < 0 {
		added = 0
	}
	
	merged := newCount - added
	if merged < 0 {
		merged = 0
	}

	return ImportStats{
		Total:   newCount,
		Added:   added,
		Merged:  merged,
		Skipped: 0,
	}
}
