package model

import "time"

// BookItem represents a single book with all available metadata from various sources.
type BookItem struct {
	ID            string     `json:"id,omitempty"`
	Title         string     `json:"title"`
	Authors       []string   `json:"authors,omitempty"`
	ASIN          string     `json:"asin,omitempty"`
	ISBN10        string     `json:"isbn10,omitempty"`
	ISBN13        string     `json:"isbn13,omitempty"`
	Format        string     `json:"format,omitempty"`        // audiobook|ebook
	Source        []string   `json:"sources,omitempty"`       // e.g., ["audible"]
	StartedAt     *time.Time `json:"startedAt,omitempty"`
	FinishedAt    *time.Time `json:"finishedAt,omitempty"`
	Rating        *int       `json:"rating,omitempty"`        // 0..5
	Review        string     `json:"review,omitempty"`
	Publisher     string     `json:"publisher,omitempty"`
	PageCount     *int       `json:"pageCount,omitempty"`
	YearPublished *int       `json:"yearPublished,omitempty"`
	Notes         []string   `json:"notes,omitempty"`
}

// Library represents a collection of books from multiple sources.
type Library struct {
	Items []BookItem `json:"items"`
}
