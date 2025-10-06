package export

import (
	"encoding/csv"
	"fmt"
	"strings"

	"media2goodreads/internal/model"
	"media2goodreads/internal/util"

	"github.com/spf13/afero"
)

// GoodreadsCSVHeader is the exact header expected by Goodreads.
var GoodreadsCSVHeader = []string{
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

// ExportOptions contains options for exporting to Goodreads CSV.
type ExportOptions struct {
	Shelf     string // read|to-read|currently-reading
	DateAdded string // YYYY-MM-DD or empty for today
	Timezone  string
}

// ExportGoodreadsCSV exports a library to a Goodreads-compatible CSV file.
func ExportGoodreadsCSV(fs afero.Fs, library *model.Library, outputPath string, opts ExportOptions) error {
	file, err := fs.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(GoodreadsCSVHeader); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Parse date added
	var dateAdded string
	if opts.DateAdded != "" {
		if t, err := util.ParseDate(opts.DateAdded, opts.Timezone); err == nil {
			dateAdded = util.FormatGoodreadsDate(t)
		}
	}
	if dateAdded == "" {
		today := util.TodayInTimezone(opts.Timezone)
		dateAdded = util.FormatGoodreadsDate(&today)
	}

	// Write rows
	for _, item := range library.Items {
		row := itemToGoodreadsRow(item, opts.Shelf, dateAdded)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

func itemToGoodreadsRow(item model.BookItem, defaultShelf, dateAdded string) []string {
	row := make([]string, len(GoodreadsCSVHeader))

	// Book Id - leave empty
	row[0] = ""

	// Title
	row[1] = item.Title

	// Author (first author)
	if len(item.Authors) > 0 {
		row[2] = item.Authors[0]
		// Author l-f (Last, First)
		row[3] = authorToLastFirst(item.Authors[0])
	}

	// Additional Authors (remaining authors, comma-separated)
	if len(item.Authors) > 1 {
		row[4] = strings.Join(item.Authors[1:], ", ")
	}

	// ISBN
	row[5] = item.ISBN10

	// ISBN13
	row[6] = item.ISBN13

	// My Rating (0-5)
	if item.Rating != nil {
		row[7] = fmt.Sprintf("%d", *item.Rating)
	}

	// Average Rating - leave empty
	row[8] = ""

	// Publisher
	row[9] = item.Publisher

	// Binding - leave empty (or could map format)
	row[10] = ""

	// Number of Pages
	if item.PageCount != nil {
		row[11] = fmt.Sprintf("%d", *item.PageCount)
	}

	// Year Published
	if item.YearPublished != nil {
		row[12] = fmt.Sprintf("%d", *item.YearPublished)
	}

	// Original Publication Year - leave empty
	row[13] = ""

	// Date Read (MM/DD/YYYY)
	if item.FinishedAt != nil {
		row[14] = util.FormatGoodreadsDate(item.FinishedAt)
	}

	// Date Added (MM/DD/YYYY)
	row[15] = dateAdded

	// Bookshelves (format + sources)
	bookshelves := buildBookshelves(item)
	row[16] = strings.Join(bookshelves, ", ")

	// Bookshelves with positions - leave empty
	row[17] = ""

	// Exclusive Shelf
	shelf := defaultShelf
	if shelf == "" {
		if item.FinishedAt != nil {
			shelf = "read"
		} else {
			shelf = "to-read"
		}
	}
	row[18] = shelf

	// My Review
	row[19] = item.Review

	// Spoiler - leave empty
	row[20] = ""

	// Private Notes (join notes)
	if len(item.Notes) > 0 {
		row[21] = strings.Join(item.Notes, "\n")
	}

	// Read Count - leave empty
	row[22] = ""

	// Owned Copies - leave empty
	row[23] = ""

	return row
}

func authorToLastFirst(author string) string {
	author = strings.TrimSpace(author)
	parts := strings.Fields(author)
	if len(parts) < 2 {
		return author
	}
	// Simple heuristic: last word is last name
	lastName := parts[len(parts)-1]
	firstName := strings.Join(parts[:len(parts)-1], " ")
	return lastName + ", " + firstName
}

func buildBookshelves(item model.BookItem) []string {
	shelves := make([]string, 0)
	
	// Add format
	if item.Format != "" {
		shelves = append(shelves, item.Format)
	}
	
	// Add sources
	for _, source := range item.Source {
		if source != "" {
			shelves = append(shelves, source)
		}
	}
	
	return shelves
}
