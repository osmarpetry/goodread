package export

import (
	"testing"
)

func TestGoodreadsCSVHeader(t *testing.T) {
	expectedColumns := 24
	if len(GoodreadsCSVHeader) != expectedColumns {
		t.Errorf("Expected %d columns, got %d", expectedColumns, len(GoodreadsCSVHeader))
	}

	// Verify exact header text
	expectedHeaders := []string{
		"Book Id",
		"Title",
		"Author",
		"Author l-f",
		"Additional Authors",
		"ISBN",
		"ISBN13",
		"My Rating",
		"Average Rating",
		"Publisher",
		"Binding",
		"Number of Pages",
		"Year Published",
		"Original Publication Year",
		"Date Read",
		"Date Added",
		"Bookshelves",
		"Bookshelves with positions",
		"Exclusive Shelf",
		"My Review",
		"Spoiler",
		"Private Notes",
		"Read Count",
		"Owned Copies",
	}

	for i, expected := range expectedHeaders {
		if GoodreadsCSVHeader[i] != expected {
			t.Errorf("Header[%d] = %q, want %q", i, GoodreadsCSVHeader[i], expected)
		}
	}
}

func TestAuthorToLastFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"John Smith", "Smith, John"},
		{"Mary Jane Watson", "Watson, Mary Jane"},
		{"Prince", "Prince"},
	}

	for _, tt := range tests {
		result := authorToLastFirst(tt.input)
		if result != tt.expected {
			t.Errorf("authorToLastFirst(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
