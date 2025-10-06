package util

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	whitespaceRegex = regexp.MustCompile(`\s+`)
	subtitleRegex   = regexp.MustCompile(`:.*$`)
	tagsRegex       = regexp.MustCompile(`\s*\([^)]*\)\s*$`)
)

// NormalizeText normalizes Unicode to NFC, trims whitespace, and collapses spaces.
func NormalizeText(s string) string {
	// Normalize to NFC
	s = norm.NFC.String(s)
	// Trim and collapse whitespace
	s = strings.TrimSpace(s)
	s = whitespaceRegex.ReplaceAllString(s, " ")
	return s
}

// NormalizeTitle standardizes a book title by removing subtitles and tags.
func NormalizeTitle(title string) string {
	title = NormalizeText(title)
	// Remove subtitle after colon
	title = subtitleRegex.ReplaceAllString(title, "")
	// Remove tags like (Unabridged), (Book 1), etc.
	title = tagsRegex.ReplaceAllString(title, "")
	return strings.TrimSpace(title)
}

// NormalizeAuthor converts "Last, First" to "First Last" and normalizes text.
func NormalizeAuthor(author string) string {
	author = NormalizeText(author)
	// Check for "Last, First" format
	if idx := strings.Index(author, ","); idx > 0 && idx < len(author)-1 {
		last := strings.TrimSpace(author[:idx])
		first := strings.TrimSpace(author[idx+1:])
		return first + " " + last
	}
	return author
}

// NormalizeAuthors normalizes a slice of authors.
func NormalizeAuthors(authors []string) []string {
	result := make([]string, 0, len(authors))
	for _, author := range authors {
		if normalized := NormalizeAuthor(author); normalized != "" {
			result = append(result, normalized)
		}
	}
	return result
}

// Slugify creates a normalized slug from text for matching purposes.
func Slugify(s string) string {
	s = NormalizeText(s)
	s = strings.ToLower(s)
	// Remove non-alphanumeric characters
	var builder strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
		} else if unicode.IsSpace(r) {
			builder.WriteRune('-')
		}
	}
	result := builder.String()
	// Collapse multiple dashes
	result = regexp.MustCompile(`-+`).ReplaceAllString(result, "-")
	return strings.Trim(result, "-")
}

// ISBN10to13 converts an ISBN-10 to ISBN-13 by adding 978 prefix and recalculating checksum.
func ISBN10to13(isbn10 string) string {
	if len(isbn10) != 10 {
		return ""
	}
	
	// Remove any hyphens
	isbn10 = strings.ReplaceAll(isbn10, "-", "")
	
	// Add 978 prefix
	isbn13 := "978" + isbn10[:9]
	
	// Calculate checksum
	sum := 0
	for i, ch := range isbn13 {
		digit := int(ch - '0')
		if i%2 == 0 {
			sum += digit
		} else {
			sum += digit * 3
		}
	}
	checksum := (10 - (sum % 10)) % 10
	
	return isbn13 + string(rune('0'+checksum))
}

// NormalizeISBN removes hyphens and spaces from an ISBN.
func NormalizeISBN(isbn string) string {
	isbn = strings.ReplaceAll(isbn, "-", "")
	isbn = strings.ReplaceAll(isbn, " ", "")
	return isbn
}
