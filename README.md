# media2goodreads

A production-ready Go application that imports your reading history from Audible, Kindle, and Storytel into a unified JSON format, then exports it to a Goodreads-compatible CSV for easy import.

## Features

- 📚 **Multiple Sources**: Import from Audible (via OpenAudible), Kindle, and Storytel
- 🔄 **Unified Library**: Consolidates all your books into a single, deduplicated JSON library
- 📊 **Goodreads Export**: Generates a CSV file that can be imported directly into Goodreads
- 🎨 **Dual Interface**: Both a classic CLI and an interactive TUI (Text-based User Interface)
- 🔒 **Privacy-First**: All processing happens locally—no logins, no API calls
- ✨ **Smart Deduplication**: Merges duplicate entries using ISBN13, ISBN10, ASIN, or title+author matching
- 🕐 **Timezone-Aware**: Properly handles date normalization and formatting

## What It Does

`media2goodreads` solves the problem of consolidating your scattered reading history:

1. **Import** from multiple sources (Audible, Kindle, Storytel)
2. **Merge** into a unified library with smart deduplication
3. **Export** to Goodreads-compatible CSV format
4. **Track** your complete reading history in one place

## Prerequisites

- Go 1.22 or higher
- (Optional) [golangci-lint](https://golangci-lint.run/welcome/install/) for development

## Installation

```bash
git clone https://github.com/yourusername/media2goodreads.git
cd media2goodreads

# Install dependencies
make deps

# Build the application
make build

# (Optional) Install to your PATH
make install
```

## Configuration

Copy the example configuration file and customize it:

```bash
cp configs/env.example .env
```

Edit `.env` with your preferred settings:

```bash
# Output directory for generated files
MEDIA2GR_OUT_DIR=./out

# Timezone for date normalization
MEDIA2GR_TZ=America/Sao_Paulo

# Default shelf for Goodreads export (read|to-read|currently-reading)
MEDIA2GR_DEFAULT_SHELF=read

# Optional date added to Goodreads (YYYY-MM-DD); defaults to today if empty
MEDIA2GR_DATE_ADDED=

# Audible import settings
MEDIA2GR_AUDIBLE_EXPORT_PATH=/path/to/openaudible/books.json
MEDIA2GR_AUDIBLE_FORMAT=json
MEDIA2GR_AUDIBLE_REGION=br

# Kindle import settings
MEDIA2GR_KINDLE_CLIPPINGS_PATH=/path/to/My Clippings.txt
MEDIA2GR_KINDLE_NOTEBOOK_HTML_DIR=

# Storytel import settings
MEDIA2GR_STORYTEL_EXPORT_PATH=/path/to/storytel_history.csv
MEDIA2GR_STORYTEL_FORMAT=csv
```

## How to Export Your Data

### Audible (via OpenAudible)

1. Download and install [OpenAudible](https://openaudible.org/)
2. Connect your Audible account and download your library metadata
3. Export your library: **File → Export → JSON** or **CSV**
4. Use the exported `books.json` or `books.csv` file

### Kindle

**Option 1: My Clippings.txt**

1. Connect your Kindle device to your computer
2. Navigate to the Kindle drive (appears as a USB drive)
3. Copy `documents/My Clippings.txt` from your Kindle

**Option 2: Kindle Notebook (for more detail)**

1. Go to [Amazon Kindle Notebook](https://read.amazon.com/notebook)
2. For each book, save the notebook page as HTML (Ctrl+S or Cmd+S)
3. Save all HTML files to a directory

### Storytel

1. Log in to your Storytel account on the web
2. Go to your listening history or library
3. Export your history (if available) as CSV or JSON
   - _Note: Storytel's export capabilities may vary by region_

### Goodreads Import

After generating your CSV:

1. Go to [Goodreads Import & Export](https://www.goodreads.com/review/import)
2. Click "Choose File" and select your `goodreads_import.csv`
3. Click "Import books"

## Quick Start

### Using the TUI (Recommended for beginners)

```bash
./bin/media2goodreads tui
```

The TUI provides an interactive menu to guide you through:

- Selecting your import source
- Entering file paths
- Viewing import results
- Exporting to Goodreads CSV

**Keyboard shortcuts:**

- `↑/↓` or `Tab`: Navigate
- `Enter`: Select/Confirm
- `Esc`: Go back
- `o`: Open output folder
- `q`: Quit

### Using the CLI

**1. Import from Audible**

```bash
./bin/media2goodreads audible import \
  --from ./exports/books.json \
  --format json \
  --region br
```

**2. Import from Kindle**

```bash
./bin/media2goodreads kindle import \
  --clippings "./My Clippings.txt" \
  --notebook-html ./kindle_notebooks
```

**3. Import from Storytel**

```bash
./bin/media2goodreads storytel import \
  --from ./storytel_history.csv \
  --format csv
```

**4. Generate Goodreads CSV**

```bash
./bin/media2goodreads goodreads-csv \
  --in ./out/library.json \
  --out ./out/goodreads_import.csv \
  --shelf read \
  --added 2025-10-05
```

**All-in-one example:**

```bash
# Import from all sources
./bin/media2goodreads audible import --from ./audible_books.json --format json
./bin/media2goodreads kindle import --clippings "./My Clippings.txt"
./bin/media2goodreads storytel import --from ./storytel.csv --format csv

# Export to Goodreads
./bin/media2goodreads goodreads-csv --in ./out/library.json --out ./out/goodreads_import.csv
```

## Data Model

The unified library uses this data model:

```go
type BookItem struct {
    ID            string        // Auto-generated unique ID
    Title         string        // Normalized title
    Authors       []string      // List of authors (normalized)
    ASIN          string        // Amazon Standard Identification Number
    ISBN10        string        // 10-digit ISBN
    ISBN13        string        // 13-digit ISBN
    Format        string        // "audiobook" or "ebook"
    Source        []string      // e.g., ["audible", "kindle"]
    StartedAt     *time.Time    // When you started reading
    FinishedAt    *time.Time    // When you finished
    Rating        *int          // Your rating (0-5)
    Review        string        // Your review text
    Publisher     string        // Publisher name
    PageCount     *int          // Number of pages
    YearPublished *int          // Year of publication
    Notes         []string      // Additional notes
}
```

## Deduplication Rules

Books are deduplicated using this priority order:

1. **ISBN13** - Highest priority, most reliable
2. **ISBN10** - Second priority
3. **ASIN** - Third priority (for Audible books)
4. **Title + First Author Slug** - Fallback for books without identifiers

When merging duplicate entries:

- Non-empty values are preferred
- Authors and sources are combined (union)
- Earliest `StartedAt` is kept
- Latest `FinishedAt` is kept
- Existing rating is preserved

## Goodreads CSV Mapping

The CSV export follows Goodreads' exact format requirements:

| Goodreads Field    | Source                                        |
| ------------------ | --------------------------------------------- |
| Title              | `BookItem.Title`                              |
| Author             | First author                                  |
| Author l-f         | First author in "Last, First" format          |
| Additional Authors | Remaining authors (comma-separated)           |
| ISBN               | `BookItem.ISBN10`                             |
| ISBN13             | `BookItem.ISBN13`                             |
| My Rating          | `BookItem.Rating` (0-5 scale)                 |
| Publisher          | `BookItem.Publisher`                          |
| Year Published     | `BookItem.YearPublished`                      |
| Date Read          | `BookItem.FinishedAt` (MM/DD/YYYY)            |
| Date Added         | Today or `--added` flag (MM/DD/YYYY)          |
| Bookshelves        | Format + sources (e.g., "audiobook, audible") |
| Exclusive Shelf    | "read" if `FinishedAt` exists, else "to-read" |
| My Review          | `BookItem.Review`                             |
| Private Notes      | `BookItem.Notes` (joined)                     |

## Development

### Running Tests

```bash
make test
```

### Running with Coverage

```bash
make test-coverage
```

### Linting

```bash
make lint
```

### Format Code

```bash
make fmt
```

### Run All Checks

```bash
make check
```

## Project Structure

```
media2goodreads/
├── cmd/media2goodreads/          # Main application entry point
│   └── main.go
├── internal/
│   ├── model/                    # Data models
│   │   └── types.go
│   ├── util/                     # Utility functions
│   │   ├── normalize.go          # Text normalization
│   │   ├── dedupe.go             # Deduplication logic
│   │   ├── dates.go              # Date parsing/formatting
│   │   ├── files.go              # File operations
│   │   └── logger.go             # Structured logging
│   ├── ingest/                   # Import parsers
│   │   ├── audible.go
│   │   ├── kindle.go
│   │   ├── storytel.go
│   │   └── merge.go
│   ├── export/                   # Export logic
│   │   └── goodreads_csv.go
│   └── tui/                      # Text-based UI
│       ├── model.go
│       ├── update.go
│       ├── view.go
│       └── components.go
├── configs/
│   └── env.example               # Example configuration
├── testdata/                     # Test fixtures
├── out/                          # Output directory (generated)
├── Makefile                      # Build automation
├── .golangci.yml                 # Linter configuration
└── README.md
```

## Troubleshooting

### CSV Rejected by Goodreads

- **Issue**: Goodreads rejects the CSV file
- **Solution**:
  - Ensure the CSV has the exact header format (run `make test` to verify)
  - Check that dates are in MM/DD/YYYY format
  - Verify ISBN format (no hyphens, correct length)

### Missing Dates

- **Issue**: Some books don't have read dates
- **Solution**:
  - For Kindle: Only books with highlights/notes will have dates (latest highlight timestamp)
  - For Audible: Uses purchase date as finished date
  - Manually edit `out/library.json` if needed before exporting

### Duplicate Books

- **Issue**: Books appear multiple times in the library
- **Solution**:
  - Check if books have proper ISBNs
  - Different editions may have different ISBNs (working as intended)
  - Manually merge in `out/library.json` if needed

### Wrong Timezone

- **Issue**: Dates are off by a day or more
- **Solution**:
  - Set `MEDIA2GR_TZ` in your `.env` file to your timezone
  - Common values: `America/New_York`, `Europe/London`, `Asia/Tokyo`
  - List of timezones: [Wikipedia TZ Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

### Kindle Clippings Not Parsing

- **Issue**: No books detected from My Clippings.txt
- **Solution**:
  - Ensure the file is in UTF-8 encoding
  - Check that highlights are separated by `==========`
  - Verify the file isn't corrupted

## Privacy & Security

✅ **All processing is local** - No data leaves your computer  
✅ **No logins required** - Works entirely with exported files  
✅ **No API calls** - Doesn't connect to any external services  
✅ **Open source** - Inspect the code yourself

Your reading data is personal. This tool respects your privacy by keeping everything on your machine.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Acknowledgments

- [OpenAudible](https://openaudible.org/) - For making Audible exports possible
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - For the excellent TUI framework
- [Cobra](https://github.com/spf13/cobra) - For CLI scaffolding
- The Goodreads community for maintaining import compatibility

---

**Made with ❤️ for book lovers who want their reading history in one place**
