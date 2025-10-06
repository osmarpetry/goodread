package util

import (
	"testing"
	"time"
)

func TestFormatGoodreadsDate(t *testing.T) {
	tests := []struct {
		name     string
		date     *time.Time
		expected string
	}{
		{
			name:     "nil date",
			date:     nil,
			expected: "",
		},
		{
			name:     "valid date",
			date:     timePtr(2025, 10, 5),
			expected: "10/05/2025",
		},
		{
			name:     "single digit month and day",
			date:     timePtr(2025, 1, 3),
			expected: "01/03/2025",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatGoodreadsDate(tt.date)
			if result != tt.expected {
				t.Errorf("FormatGoodreadsDate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func timePtr(year, month, day int) *time.Time {
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &t
}
