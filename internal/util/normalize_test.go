package util

import (
	"testing"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  world  ", "hello world"},
		{"hello\n\nworld", "hello world"},
		{"café", "café"},
		{"", ""},
	}

	for _, tt := range tests {
		result := NormalizeText(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeText(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The Great Book: A Subtitle", "The Great Book"},
		{"Book Title (Unabridged)", "Book Title"},
		{"Book Title (Book 1)", "Book Title"},
		{"Simple Title", "Simple Title"},
	}

	for _, tt := range tests {
		result := NormalizeTitle(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeTitle(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestNormalizeAuthor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Smith, John", "John Smith"},
		{"John Smith", "John Smith"},
		{"", ""},
	}

	for _, tt := range tests {
		result := NormalizeAuthor(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeAuthor(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"The Great Book", "the-great-book"},
		{"Book & Title!", "book-title"},
		{"  Multiple   Spaces  ", "multiple-spaces"},
	}

	for _, tt := range tests {
		result := Slugify(tt.input)
		if result != tt.expected {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestISBN10to13(t *testing.T) {
	tests := []struct {
		isbn10   string
		expected string
	}{
		{"0306406152", "9780306406157"},
		{"", ""},
	}

	for _, tt := range tests {
		result := ISBN10to13(tt.isbn10)
		if result != tt.expected {
			t.Errorf("ISBN10to13(%q) = %q, want %q", tt.isbn10, result, tt.expected)
		}
	}
}
