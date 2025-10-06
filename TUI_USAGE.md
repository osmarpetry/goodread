# TUI Navigation Guide

## Starting the TUI

```bash
./bin/media2goodreads tui
```

## Navigation Keys

### Home Menu

- **↑/↓ or j/k**: Navigate up/down through menu items
- **Enter**: Select the highlighted menu item
- **o**: Open output folder in file explorer
- **q or Ctrl+C**: Quit the application

### Import/Export Forms

- **Tab or ↓**: Move to next input field
- **Shift+Tab or ↑**: Move to previous input field
- **Enter**: Submit the form and start processing
- **Esc**: Go back to home menu

### Result Screen

- **Enter**: Return to home menu
- **o**: Open output folder
- **q**: Quit application

## Menu Options

1. **Import from Audible**

   - Enter path to OpenAudible export file
   - Select format (json or csv)
   - Press Enter to import

2. **Import from Kindle**

   - Enter path to My Clippings.txt
   - Optionally enter path to Notebook HTML directory
   - Press Enter to import

3. **Import from Storytel**

   - Enter path to Storytel export file
   - Select format (json or csv)
   - Press Enter to import

4. **Export to Goodreads CSV**

   - Input and output paths are pre-filled
   - Select shelf (read, to-read, or currently-reading)
   - Optionally set date added (YYYY-MM-DD format)
   - Press Enter to export

5. **Open Output Folder**

   - Opens the output directory in your system file explorer
   - Works on Linux (xdg-open), macOS (open), and Windows (start)

6. **Quit**
   - Exits the application

## Tips

- The library is cumulative—each import adds to the existing `out/library.json`
- Duplicate books are automatically merged based on ISBN, ASIN, or title+author
- You can import from multiple sources before exporting to Goodreads
- Use the output folder option to quickly access your exported files

## Troubleshooting

**Keys not responding?**

- Try pressing Enter or Esc to ensure you're in the correct view
- Make sure your terminal supports arrow keys
- Try using j/k keys as alternatives to arrow keys

**Can't see the full menu?**

- Resize your terminal window to at least 80x24 characters
- The TUI adapts to terminal size automatically

**TUI looks weird?**

- Ensure your terminal supports UTF-8 and ANSI colors
- Try a modern terminal emulator (iTerm2, Windows Terminal, GNOME Terminal)
