# Quick Start Guide

## Build and Run

```bash
# Build the application
make build

# Run the interactive TUI (recommended)
./bin/media2goodreads tui

# Or use the CLI directly
./bin/media2goodreads --help
```

## Basic Workflow

### 1. Prepare Your Exports

- **Audible**: Export from [OpenAudible](https://openaudible.org/) as JSON or CSV
- **Kindle**: Copy `My Clippings.txt` from your Kindle device
- **Storytel**: Export your history as CSV (if available in your region)

### 2. Import Your Data

Using TUI:

```bash
./bin/media2goodreads tui
# Select "Import from Audible", "Import from Kindle", or "Import from Storytel"
```

Using CLI:

```bash
# Import from Audible
./bin/media2goodreads audible import --from ./audible_books.json --format json

# Import from Kindle
./bin/media2goodreads kindle import --clippings "./My Clippings.txt"

# Import from Storytel
./bin/media2goodreads storytel import --from ./storytel.csv --format csv
```

### 3. Export to Goodreads

Using TUI:

```bash
./bin/media2goodreads tui
# Select "Export to Goodreads CSV"
```

Using CLI:

```bash
./bin/media2goodreads goodreads-csv \
  --in ./out/library.json \
  --out ./out/goodreads_import.csv \
  --shelf read
```

### 4. Import to Goodreads

1. Go to [Goodreads Import & Export](https://www.goodreads.com/review/import)
2. Click "Choose File" and select `out/goodreads_import.csv`
3. Click "Import books"

## Configuration

Create a `.env` file (copy from `configs/env.example`):

```bash
cp configs/env.example .env
# Edit .env with your paths and preferences
```

Key settings:

- `MEDIA2GR_OUT_DIR`: Where to save output files (default: `./out`)
- `MEDIA2GR_TZ`: Your timezone (e.g., `America/New_York`)
- `MEDIA2GR_DEFAULT_SHELF`: read, to-read, or currently-reading

## Output Files

All output is saved to `./out/` (configurable):

- `library.json`: Your unified book library
- `goodreads_import.csv`: CSV ready for Goodreads import

## Tips

- Import from multiple sources—they'll be automatically merged and deduplicated
- The library is cumulative—each import adds to the existing library
- Open the output folder with `o` key in TUI
- Use `--verbose` flag with CLI for detailed logging

## Troubleshooting

**Build fails?**

```bash
make deps
make build
```

**Tests failing?**

```bash
make test
```

**Need to start fresh?**

```bash
rm -rf out/library.json
# Then reimport your data
```
