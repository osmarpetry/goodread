package util

import (
	"media2goodreads/internal/model"
	"sort"
)

// DedupeKey returns a deduplication key for a book item.
// Priority: ISBN13 > ISBN10 > ASIN > slug(title+firstAuthor)
func DedupeKey(item model.BookItem) string {
	if item.ISBN13 != "" {
		return "isbn13:" + NormalizeISBN(item.ISBN13)
	}
	if item.ISBN10 != "" {
		return "isbn10:" + NormalizeISBN(item.ISBN10)
	}
	if item.ASIN != "" {
		return "asin:" + item.ASIN
	}
	// Fallback to title+author slug
	slug := Slugify(item.Title)
	if len(item.Authors) > 0 {
		slug += "-" + Slugify(item.Authors[0])
	}
	return "slug:" + slug
}

// MergeBooks merges two book items, preferring non-empty values and unioning sources.
func MergeBooks(existing, incoming model.BookItem) model.BookItem {
	result := existing

	// Prefer non-empty title
	if result.Title == "" && incoming.Title != "" {
		result.Title = incoming.Title
	}

	// Merge authors (union, dedupe)
	result.Authors = unionStrings(result.Authors, incoming.Authors)

	// Prefer non-empty identifiers
	if result.ASIN == "" && incoming.ASIN != "" {
		result.ASIN = incoming.ASIN
	}
	if result.ISBN10 == "" && incoming.ISBN10 != "" {
		result.ISBN10 = incoming.ISBN10
	}
	if result.ISBN13 == "" && incoming.ISBN13 != "" {
		result.ISBN13 = incoming.ISBN13
	}

	// Prefer non-empty format
	if result.Format == "" && incoming.Format != "" {
		result.Format = incoming.Format
	}

	// Union sources
	result.Source = unionStrings(result.Source, incoming.Source)

	// Prefer earliest StartedAt
	if result.StartedAt == nil {
		result.StartedAt = incoming.StartedAt
	} else if incoming.StartedAt != nil && incoming.StartedAt.Before(*result.StartedAt) {
		result.StartedAt = incoming.StartedAt
	}

	// Prefer latest FinishedAt
	if result.FinishedAt == nil {
		result.FinishedAt = incoming.FinishedAt
	} else if incoming.FinishedAt != nil && incoming.FinishedAt.After(*result.FinishedAt) {
		result.FinishedAt = incoming.FinishedAt
	}

	// Prefer non-zero rating
	if result.Rating == nil {
		result.Rating = incoming.Rating
	}

	// Prefer non-empty review
	if result.Review == "" && incoming.Review != "" {
		result.Review = incoming.Review
	}

	// Prefer non-empty publisher
	if result.Publisher == "" && incoming.Publisher != "" {
		result.Publisher = incoming.Publisher
	}

	// Prefer non-zero page count
	if result.PageCount == nil {
		result.PageCount = incoming.PageCount
	}

	// Prefer non-zero year published
	if result.YearPublished == nil {
		result.YearPublished = incoming.YearPublished
	}

	// Merge notes
	result.Notes = unionStrings(result.Notes, incoming.Notes)

	return result
}

// DedupeLibrary deduplicates a library based on deduplication keys.
func DedupeLibrary(items []model.BookItem) []model.BookItem {
	keyMap := make(map[string]model.BookItem)
	
	for _, item := range items {
		key := DedupeKey(item)
		if existing, found := keyMap[key]; found {
			keyMap[key] = MergeBooks(existing, item)
		} else {
			keyMap[key] = item
		}
	}
	
	// Convert map back to slice
	result := make([]model.BookItem, 0, len(keyMap))
	for _, item := range keyMap {
		result = append(result, item)
	}
	
	// Sort by title for deterministic output
	sort.Slice(result, func(i, j int) bool {
		return result[i].Title < result[j].Title
	})
	
	return result
}

// unionStrings returns a deduplicated union of two string slices.
func unionStrings(a, b []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(a)+len(b))
	
	for _, s := range a {
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	for _, s := range b {
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	
	return result
}
