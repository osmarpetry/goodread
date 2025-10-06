package util

import (
	"media2goodreads/internal/model"
	"testing"
)

func TestDedupeKey(t *testing.T) {
	tests := []struct {
		name     string
		item     model.BookItem
		expected string
	}{
		{
			name:     "ISBN13 priority",
			item:     model.BookItem{ISBN13: "9780306406157", ISBN10: "0306406152", ASIN: "B001", Title: "Book"},
			expected: "isbn13:9780306406157",
		},
		{
			name:     "ISBN10 when no ISBN13",
			item:     model.BookItem{ISBN10: "0306406152", ASIN: "B001", Title: "Book"},
			expected: "isbn10:0306406152",
		},
		{
			name:     "ASIN when no ISBN",
			item:     model.BookItem{ASIN: "B001", Title: "Book"},
			expected: "asin:B001",
		},
		{
			name:     "Title+Author slug fallback",
			item:     model.BookItem{Title: "The Great Book", Authors: []string{"John Smith"}},
			expected: "slug:the-great-book-john-smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DedupeKey(tt.item)
			if result != tt.expected {
				t.Errorf("DedupeKey() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestMergeBooks(t *testing.T) {
	rating1 := 4
	rating2 := 5

	existing := model.BookItem{
		Title:   "Book",
		Authors: []string{"Author 1"},
		Source:  []string{"audible"},
		Rating:  &rating1,
	}

	incoming := model.BookItem{
		Title:   "Book",
		Authors: []string{"Author 2"},
		Source:  []string{"kindle"},
		Rating:  &rating2,
		ISBN13:  "9780306406157",
	}

	merged := MergeBooks(existing, incoming)

	if merged.ISBN13 != "9780306406157" {
		t.Errorf("Expected ISBN13 to be merged")
	}

	if len(merged.Authors) != 2 {
		t.Errorf("Expected 2 authors, got %d", len(merged.Authors))
	}

	if len(merged.Source) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(merged.Source))
	}

	if *merged.Rating != rating1 {
		t.Errorf("Expected existing rating to be preserved")
	}
}
