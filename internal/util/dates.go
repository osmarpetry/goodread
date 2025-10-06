package util

import (
	"fmt"
	"time"
)

// ParseDate parses various date formats and returns a time.Time in the given timezone.
func ParseDate(dateStr, tz string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}

	// Common date formats to try
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"01/02/2006",
		"02/01/2006",
		"January 2, 2006",
		"Jan 2, 2006",
		"2 January 2006",
		"2 Jan 2006",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, loc); err == nil {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("unable to parse date: %s", dateStr)
}

// FormatGoodreadsDate formats a time.Time as MM/DD/YYYY for Goodreads.
func FormatGoodreadsDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("01/02/2006")
}

// TodayInTimezone returns today's date at midnight in the specified timezone.
func TodayInTimezone(tz string) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

// ParseFlexibleDate tries to parse a date string with multiple formats.
// It handles common Audible, Kindle, and Storytel date formats.
func ParseFlexibleDate(dateStr string, tz string) (*time.Time, error) {
	return ParseDate(dateStr, tz)
}
