package ingest

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"media2goodreads/internal/model"
	"media2goodreads/internal/util"

	"github.com/spf13/afero"
)

// StorytelBook represents a book from Storytel export.
type StorytelBook struct {
	Title      string `json:"title"`
	Author     string `json:"author"`
	Authors    string `json:"authors"`
	ISBN       string `json:"isbn"`
	ISBN13     string `json:"isbn13"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	Rating     int    `json:"rating"`
	Type       string `json:"type"` // audiobook|ebook
}

// ParseStorytelJSON parses a Storytel JSON export file.
func ParseStorytelJSON(fs afero.Fs, path string, timezone string) ([]model.BookItem, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Storytel JSON: %w", err)
	}

	var storytelBooks []StorytelBook
	if err := json.Unmarshal(data, &storytelBooks); err != nil {
		return nil, fmt.Errorf("failed to parse Storytel JSON: %w", err)
	}

	items := make([]model.BookItem, 0, len(storytelBooks))
	for _, sb := range storytelBooks {
		item := convertStorytelBook(sb, timezone)
		items = append(items, item)
	}

	return items, nil
}

// ParseStorytelCSV parses a Storytel CSV export file.
func ParseStorytelCSV(fs afero.Fs, path string, timezone string) ([]model.BookItem, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Storytel CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Create header map (case-insensitive)
	headerMap := make(map[string]int)
	for i, h := range header {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	items := make([]model.BookItem, 0)
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		item := parseStorytelCSVRow(record, headerMap, timezone)
		items = append(items, item)
	}

	return items, nil
}

func convertStorytelBook(sb StorytelBook, timezone string) model.BookItem {
	item := model.BookItem{
		Title:  util.NormalizeTitle(sb.Title),
		Source: []string{"storytel"},
		Format: "audiobook", // Default
	}

	// Parse authors
	authorsStr := sb.Authors
	if authorsStr == "" {
		authorsStr = sb.Author
	}
	if authorsStr != "" {
		authorList := strings.Split(authorsStr, ",")
		item.Authors = util.NormalizeAuthors(authorList)
	}

	// ISBNs
	item.ISBN10 = util.NormalizeISBN(sb.ISBN)
	item.ISBN13 = util.NormalizeISBN(sb.ISBN13)

	// Format from type
	if sb.Type != "" {
		formatLower := strings.ToLower(sb.Type)
		if strings.Contains(formatLower, "ebook") {
			item.Format = "ebook"
		} else if strings.Contains(formatLower, "audio") {
			item.Format = "audiobook"
		}
	}

	// Dates
	if sb.StartedAt != "" {
		if t, err := util.ParseDate(sb.StartedAt, timezone); err == nil {
			item.StartedAt = t
		}
	}
	if sb.FinishedAt != "" {
		if t, err := util.ParseDate(sb.FinishedAt, timezone); err == nil {
			item.FinishedAt = t
		}
	}

	// Rating
	if sb.Rating > 0 {
		rating := sb.Rating
		if rating > 5 {
			rating = 5
		}
		item.Rating = &rating
	}

	return item
}

func parseStorytelCSVRow(record []string, headerMap map[string]int, timezone string) model.BookItem {
	getField := func(name string) string {
		if idx, ok := headerMap[strings.ToLower(name)]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	item := model.BookItem{
		Source: []string{"storytel"},
		Format: "audiobook", // Default
	}

	if title := getField("title"); title != "" {
		item.Title = util.NormalizeTitle(title)
	}

	// Parse authors
	authorsStr := getField("authors")
	if authorsStr == "" {
		authorsStr = getField("author")
	}
	if authorsStr != "" {
		authorList := strings.Split(authorsStr, ",")
		item.Authors = util.NormalizeAuthors(authorList)
	}

	item.ISBN10 = util.NormalizeISBN(getField("isbn"))
	item.ISBN13 = util.NormalizeISBN(getField("isbn13"))

	// Format from type
	if typeStr := getField("type"); typeStr != "" {
		typeLower := strings.ToLower(typeStr)
		if strings.Contains(typeLower, "ebook") {
			item.Format = "ebook"
		} else if strings.Contains(typeLower, "audio") {
			item.Format = "audiobook"
		}
	}

	// Dates
	if startedAt := getField("started_at"); startedAt != "" {
		if startedAt == "" {
			startedAt = getField("startedat")
		}
		if t, err := util.ParseDate(startedAt, timezone); err == nil {
			item.StartedAt = t
		}
	}
	if finishedAt := getField("finished_at"); finishedAt != "" {
		if finishedAt == "" {
			finishedAt = getField("finishedat")
		}
		if t, err := util.ParseDate(finishedAt, timezone); err == nil {
			item.FinishedAt = t
		}
	}

	// Rating
	if ratingStr := getField("rating"); ratingStr != "" {
		if rating, err := strconv.Atoi(ratingStr); err == nil && rating > 0 {
			if rating > 5 {
				rating = 5
			}
			item.Rating = &rating
		}
	}

	return item
}
