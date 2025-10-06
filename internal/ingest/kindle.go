package ingest

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"media2goodreads/internal/model"
	"media2goodreads/internal/util"

	"github.com/spf13/afero"
	"golang.org/x/net/html"
)

var (
	// My Clippings.txt patterns
	clippingHeaderRegex = regexp.MustCompile(`^(.+?)\s*\(([^)]+)\)$`)
	clippingDateRegex   = regexp.MustCompile(`Added on (.+)$`)
	
	// Alternative patterns
	altHeaderRegex = regexp.MustCompile(`^(.+?)(?:\s*[-–—]\s*(.+))?$`)
)

// KindleClipping represents a single clipping from My Clippings.txt.
type KindleClipping struct {
	Title     string
	Authors   []string
	Location  string
	Timestamp time.Time
	Content   string
}

// ParseKindleClippings parses a Kindle My Clippings.txt file.
func ParseKindleClippings(fs afero.Fs, path string, timezone string) ([]model.BookItem, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open Kindle clippings: %w", err)
	}
	defer file.Close()

	clippings, err := parseClippingsFile(file, timezone)
	if err != nil {
		return nil, err
	}

	// Group clippings by book
	bookMap := make(map[string]*model.BookItem)
	for _, clip := range clippings {
		key := util.Slugify(clip.Title)
		if item, exists := bookMap[key]; exists {
			// Update latest timestamp
			if item.FinishedAt == nil || clip.Timestamp.After(*item.FinishedAt) {
				item.FinishedAt = &clip.Timestamp
			}
		} else {
			bookMap[key] = &model.BookItem{
				Title:      util.NormalizeTitle(clip.Title),
				Authors:    util.NormalizeAuthors(clip.Authors),
				Format:     "ebook",
				Source:     []string{"kindle"},
				FinishedAt: &clip.Timestamp,
			}
		}
	}

	// Convert map to slice
	items := make([]model.BookItem, 0, len(bookMap))
	for _, item := range bookMap {
		items = append(items, *item)
	}

	return items, nil
}

// ParseKindleNotebookHTML parses Kindle Notebook HTML files.
func ParseKindleNotebookHTML(fs afero.Fs, dirPath string, timezone string) ([]model.BookItem, error) {
	files, err := afero.ReadDir(fs, dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read notebook directory: %w", err)
	}

	items := make([]model.BookItem, 0)
	for _, fileInfo := range files {
		if fileInfo.IsDir() || !strings.HasSuffix(strings.ToLower(fileInfo.Name()), ".html") {
			continue
		}

		filePath := dirPath + "/" + fileInfo.Name()
		item, err := parseNotebookHTML(fs, filePath, timezone)
		if err != nil {
			// Log error but continue
			continue
		}
		if item != nil {
			items = append(items, *item)
		}
	}

	return items, nil
}

func parseClippingsFile(r io.Reader, timezone string) ([]KindleClipping, error) {
	loc, _ := time.LoadLocation(timezone)
	if loc == nil {
		loc = time.UTC
	}

	scanner := bufio.NewScanner(r)
	var clippings []KindleClipping
	var currentClip *KindleClipping

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Separator between clippings
		if strings.HasPrefix(line, "==========") {
			if currentClip != nil && currentClip.Title != "" {
				clippings = append(clippings, *currentClip)
			}
			currentClip = &KindleClipping{}
			continue
		}

		if currentClip == nil {
			currentClip = &KindleClipping{}
		}

		// Parse title and author
		if currentClip.Title == "" && line != "" {
			if matches := clippingHeaderRegex.FindStringSubmatch(line); len(matches) > 2 {
				currentClip.Title = strings.TrimSpace(matches[1])
				authorsStr := strings.TrimSpace(matches[2])
				// Split by semicolon or comma
				if strings.Contains(authorsStr, ";") {
					currentClip.Authors = strings.Split(authorsStr, ";")
				} else {
					currentClip.Authors = []string{authorsStr}
				}
			} else if altMatches := altHeaderRegex.FindStringSubmatch(line); len(altMatches) > 1 {
				currentClip.Title = strings.TrimSpace(altMatches[1])
				if len(altMatches) > 2 && altMatches[2] != "" {
					currentClip.Authors = []string{strings.TrimSpace(altMatches[2])}
				}
			} else {
				currentClip.Title = line
			}
			continue
		}

		// Parse metadata line (location, date)
		if strings.Contains(line, "Location") || strings.Contains(line, "Page") {
			if matches := clippingDateRegex.FindStringSubmatch(line); len(matches) > 1 {
				dateStr := strings.TrimSpace(matches[1])
				// Parse various date formats
				if t, err := parseKindleDate(dateStr, loc); err == nil {
					currentClip.Timestamp = t
				}
			}
			continue
		}

		// Content
		if line != "" && currentClip.Title != "" {
			if currentClip.Content != "" {
				currentClip.Content += " "
			}
			currentClip.Content += line
		}
	}

	// Add last clipping
	if currentClip != nil && currentClip.Title != "" {
		clippings = append(clippings, *currentClip)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading clippings: %w", err)
	}

	return clippings, nil
}

func parseKindleDate(dateStr string, loc *time.Location) (time.Time, error) {
	// Kindle date formats
	formats := []string{
		"Monday, January 2, 2006 3:04:05 PM",
		"Monday, 2 January 2006 15:04:05",
		"January 2, 2006",
		"2 January 2006",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateStr, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse Kindle date: %s", dateStr)
}

func parseNotebookHTML(fs afero.Fs, path string, timezone string) (*model.BookItem, error) {
	file, err := fs.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	doc, err := html.Parse(file)
	if err != nil {
		return nil, err
	}

	item := &model.BookItem{
		Format: "ebook",
		Source: []string{"kindle"},
	}

	var latestTime *time.Time

	// Traverse HTML and extract title, authors, highlights
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Look for book title and author in common locations
			if n.Data == "h1" || n.Data == "h2" {
				text := extractText(n)
				if item.Title == "" && text != "" {
					item.Title = util.NormalizeTitle(text)
				}
			}
			
			// Look for author
			if hasClass(n, "author") {
				text := extractText(n)
				if text != "" && len(item.Authors) == 0 {
					item.Authors = []string{util.NormalizeAuthor(text)}
				}
			}

			// Look for timestamps in highlights
			if hasClass(n, "timestamp") || hasClass(n, "noteHeading") {
				text := extractText(n)
				if t, err := util.ParseDate(text, timezone); err == nil {
					if latestTime == nil || t.After(*latestTime) {
						latestTime = t
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	item.FinishedAt = latestTime

	if item.Title == "" {
		return nil, fmt.Errorf("no title found in HTML")
	}

	return item, nil
}

func extractText(n *html.Node) string {
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}
	var text string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text += extractText(c)
	}
	return strings.TrimSpace(text)
}

func hasClass(n *html.Node, className string) bool {
	for _, attr := range n.Attr {
		if attr.Key == "class" {
			classes := strings.Fields(attr.Val)
			for _, c := range classes {
				if c == className {
					return true
				}
			}
		}
	}
	return false
}
