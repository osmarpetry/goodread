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

// AudibleBook represents the structure of an OpenAudible book export.
type AudibleBook struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Narrators     []string `json:"narrators"`
	ASIN          string   `json:"asin"`
	ISBN          string   `json:"isbn"`
	ISBN13        string   `json:"isbn13"`
	Publisher     string   `json:"publisher"`
	Purchased     string   `json:"purchased"`
	Released      string   `json:"released"`
	Rating        float64  `json:"rating"`
	Series        string   `json:"series"`
	SeriesOrder   string   `json:"series_order"`
	Duration      int      `json:"duration"`
	Language      string   `json:"language"`
	Region        string   `json:"region"`
}

// ParseAudibleJSON parses an OpenAudible JSON export file.
func ParseAudibleJSON(fs afero.Fs, path string, timezone string) ([]model.BookItem, error) {
	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read Audible JSON: %w", err)
	}

	var audibleBooks []AudibleBook
	if err := json.Unmarshal(data, &audibleBooks); err != nil {
		return nil, fmt.Errorf("failed to parse Audible JSON: %w", err)
	}

	items := make([]model.BookItem, 0, len(audibleBooks))
	for _, ab := range audibleBooks {
		item := convertAudibleBook(ab, timezone)
		items = append(items, item)
	}

	return items, nil
}

// ParseAudibleCSV parses an OpenAudible CSV export file.
func ParseAudibleCSV(fs afero.Fs, path string, timezone string) ([]model.BookItem, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Audible CSV: %w", err)
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

		item := parseAudibleCSVRow(record, headerMap, timezone)
		items = append(items, item)
	}

	return items, nil
}

func convertAudibleBook(ab AudibleBook, timezone string) model.BookItem {
	item := model.BookItem{
		Title:     util.NormalizeTitle(ab.Title),
		Authors:   util.NormalizeAuthors(ab.Authors),
		ASIN:      ab.ASIN,
		ISBN10:    util.NormalizeISBN(ab.ISBN),
		ISBN13:    util.NormalizeISBN(ab.ISBN13),
		Format:    "audiobook",
		Source:    []string{"audible"},
		Publisher: util.NormalizeText(ab.Publisher),
	}

	// Parse purchased date as finished date
	if ab.Purchased != "" {
		if t, err := util.ParseDate(ab.Purchased, timezone); err == nil {
			item.FinishedAt = t
		}
	}

	// Parse rating (convert 0-5 scale)
	if ab.Rating > 0 {
		rating := int(ab.Rating)
		if rating > 5 {
			rating = 5
		}
		item.Rating = &rating
	}

	// Parse year published from release date
	if ab.Released != "" {
		if t, err := util.ParseDate(ab.Released, timezone); err == nil {
			year := t.Year()
			item.YearPublished = &year
		}
	}

	return item
}

func parseAudibleCSVRow(record []string, headerMap map[string]int, timezone string) model.BookItem {
	getField := func(name string) string {
		if idx, ok := headerMap[strings.ToLower(name)]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	item := model.BookItem{
		Format: "audiobook",
		Source: []string{"audible"},
	}

	if title := getField("title"); title != "" {
		item.Title = util.NormalizeTitle(title)
	}

	// Parse authors (may be comma-separated)
	if authors := getField("authors"); authors != "" {
		authorList := strings.Split(authors, ",")
		item.Authors = util.NormalizeAuthors(authorList)
	} else if author := getField("author"); author != "" {
		item.Authors = []string{util.NormalizeAuthor(author)}
	}

	item.ASIN = getField("asin")
	item.ISBN10 = util.NormalizeISBN(getField("isbn"))
	item.ISBN13 = util.NormalizeISBN(getField("isbn13"))
	item.Publisher = util.NormalizeText(getField("publisher"))

	// Parse dates
	if purchased := getField("purchased"); purchased != "" {
		if t, err := util.ParseDate(purchased, timezone); err == nil {
			item.FinishedAt = t
		}
	}
	if finished := getField("finished"); finished != "" {
		if t, err := util.ParseDate(finished, timezone); err == nil {
			item.FinishedAt = t
		}
	}

	// Parse rating
	if ratingStr := getField("rating"); ratingStr != "" {
		if rating, err := strconv.ParseFloat(ratingStr, 64); err == nil && rating > 0 {
			r := int(rating)
			if r > 5 {
				r = 5
			}
			item.Rating = &r
		}
	}

	// Parse year published
	if released := getField("released"); released != "" {
		if t, err := util.ParseDate(released, timezone); err == nil {
			year := t.Year()
			item.YearPublished = &year
		}
	}

	return item
}
