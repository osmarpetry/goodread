package tui

import (
	"fmt"
	"path/filepath"

	"media2goodreads/internal/export"
	"media2goodreads/internal/ingest"
	"media2goodreads/internal/model"
	"media2goodreads/internal/util"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
)

type importDoneMsg struct {
	stats ingest.ImportStats
	err   error
}

type exportDoneMsg struct {
	count int
	err   error
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.spinner.Tick,
	)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := boxStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-4)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case importDoneMsg:
		m.state = ViewResult
		if msg.err != nil {
			m.err = msg.err
			m.message = fmt.Sprintf("Import failed: %v", msg.err)
		} else {
			m.message = "Import completed successfully!"
			m.stats = fmt.Sprintf("Total: %d | New: %d | Merged: %d", 
				msg.stats.Total, msg.stats.Added, msg.stats.Merged)
		}
		return m, nil

	case exportDoneMsg:
		m.state = ViewResult
		if msg.err != nil {
			m.err = msg.err
			m.message = fmt.Sprintf("Export failed: %v", msg.err)
		} else {
			m.message = "Export completed successfully!"
			m.stats = fmt.Sprintf("Exported %d items", msg.count)
		}
		return m, nil

	case tea.KeyMsg:
		// Handle specific keys based on state
		if m.state == ViewHome {
			switch msg.String() {
			case "q", "ctrl+c":
				m.quitting = true
				return m, tea.Quit
			case "o":
				outDir := viper.GetString("OUT_DIR")
				_ = util.OpenDirectory(outDir)
				return m, nil
			case "enter":
				return m.handleEnter()
			}
			// Let the list handle all other keys (navigation)
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		return m.handleKeyPress(msg)
	}

	// Update active component based on state
	switch m.state {
	case ViewHome:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == ViewHome || m.state == ViewResult {
			m.quitting = true
			return m, tea.Quit
		}
		// In other views, go back to home
		m.state = ViewHome
		m.err = nil
		return m, nil

	case "esc":
		if m.state != ViewHome && m.state != ViewProgress {
			m.state = ViewHome
			m.err = nil
		}
		return m, nil

	case "enter":
		return m.handleEnter()

	case "o":
		if m.state == ViewHome || m.state == ViewResult {
			outDir := viper.GetString("OUT_DIR")
			_ = util.OpenDirectory(outDir)
		}
		return m, nil
	}

	// Handle input navigation
	if m.state == ViewAudibleImport || m.state == ViewKindleImport || 
	   m.state == ViewStorytelImport || m.state == ViewGoodreadsExport {
		switch msg.String() {
		case "tab", "down":
			m.focused++
			if m.focused >= len(m.inputs) {
				m.focused = 0
			}
			m.updateInputFocus()
			return m, nil
		case "shift+tab", "up":
			m.focused--
			if m.focused < 0 {
				m.focused = len(m.inputs) - 1
			}
			m.updateInputFocus()
			return m, nil
		}

		// Update focused input
		if m.focused < len(m.inputs) {
			var cmd tea.Cmd
			m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.state {
	case ViewHome:
		selected := m.list.SelectedItem()
		if selected == nil {
			return m, nil
		}
		
		item := selected.(menuItem)
		switch item.title {
		case "Import from Audible":
			m.state = ViewAudibleImport
			m.inputs = m.createAudibleInputs()
			m.focused = 0
			m.updateInputFocus()
		case "Import from Kindle":
			m.state = ViewKindleImport
			m.inputs = m.createKindleInputs()
			m.focused = 0
			m.updateInputFocus()
		case "Import from Storytel":
			m.state = ViewStorytelImport
			m.inputs = m.createStorytelInputs()
			m.focused = 0
			m.updateInputFocus()
		case "Export to Goodreads CSV":
			m.state = ViewGoodreadsExport
			m.inputs = m.createGoodreadsInputs()
			m.focused = 0
			m.updateInputFocus()
		case "Open Output Folder":
			outDir := viper.GetString("OUT_DIR")
			_ = util.OpenDirectory(outDir)
		case "Quit":
			m.quitting = true
			return m, tea.Quit
		}

	case ViewAudibleImport:
		return m, m.importAudible()
	case ViewKindleImport:
		return m, m.importKindle()
	case ViewStorytelImport:
		return m, m.importStorytél()
	case ViewGoodreadsExport:
		return m, m.exportGoodreads()
	case ViewResult:
		m.state = ViewHome
		m.err = nil
	}

	return m, nil
}

func (m *Model) updateInputFocus() {
	for i := range m.inputs {
		if i == m.focused {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m Model) importAudible() tea.Cmd {
	m.state = ViewProgress
	m.message = "Importing Audible library..."

	return func() tea.Msg {
		path := m.inputs[0].Value()
		format := m.inputs[1].Value()
		if format == "" {
			format = "json"
		}

		timezone := viper.GetString("TZ")
		outDir := viper.GetString("OUT_DIR")
		libraryPath := filepath.Join(outDir, "library.json")

		var items []model.BookItem
		var err error

		if format == "csv" {
			items, err = ingest.ParseAudibleCSV(m.fs, path, timezone)
		} else {
			items, err = ingest.ParseAudibleJSON(m.fs, path, timezone)
		}

		if err != nil {
			return importDoneMsg{err: err}
		}

		existingLib, _ := util.LoadLibrary(m.fs, libraryPath)
		merged, err := ingest.MergeIntoLibrary(m.fs, libraryPath, items)
		if err != nil {
			return importDoneMsg{err: err}
		}

		stats := ingest.CalculateImportStats(existingLib, merged, len(items))
		return importDoneMsg{stats: stats}
	}
}

func (m Model) importKindle() tea.Cmd {
	m.state = ViewProgress
	m.message = "Importing Kindle library..."

	return func() tea.Msg {
		clippings := m.inputs[0].Value()
		notebookDir := m.inputs[1].Value()

		timezone := viper.GetString("TZ")
		outDir := viper.GetString("OUT_DIR")
		libraryPath := filepath.Join(outDir, "library.json")

		var allItems []model.BookItem

		if clippings != "" {
			items, err := ingest.ParseKindleClippings(m.fs, clippings, timezone)
			if err != nil {
				return importDoneMsg{err: err}
			}
			allItems = append(allItems, items...)
		}

		if notebookDir != "" {
			items, err := ingest.ParseKindleNotebookHTML(m.fs, notebookDir, timezone)
			if err == nil {
				allItems = append(allItems, items...)
			}
		}

		if len(allItems) == 0 {
			return importDoneMsg{err: fmt.Errorf("no items found")}
		}

		existingLib, _ := util.LoadLibrary(m.fs, libraryPath)
		merged, err := ingest.MergeIntoLibrary(m.fs, libraryPath, allItems)
		if err != nil {
			return importDoneMsg{err: err}
		}

		stats := ingest.CalculateImportStats(existingLib, merged, len(allItems))
		return importDoneMsg{stats: stats}
	}
}

func (m Model) importStorytél() tea.Cmd {
	m.state = ViewProgress
	m.message = "Importing Storytel library..."

	return func() tea.Msg {
		path := m.inputs[0].Value()
		format := m.inputs[1].Value()
		if format == "" {
			format = "csv"
		}

		timezone := viper.GetString("TZ")
		outDir := viper.GetString("OUT_DIR")
		libraryPath := filepath.Join(outDir, "library.json")

		var items []model.BookItem
		var err error

		if format == "csv" {
			items, err = ingest.ParseStorytelCSV(m.fs, path, timezone)
		} else {
			items, err = ingest.ParseStorytelJSON(m.fs, path, timezone)
		}

		if err != nil {
			return importDoneMsg{err: err}
		}

		existingLib, _ := util.LoadLibrary(m.fs, libraryPath)
		merged, err := ingest.MergeIntoLibrary(m.fs, libraryPath, items)
		if err != nil {
			return importDoneMsg{err: err}
		}

		stats := ingest.CalculateImportStats(existingLib, merged, len(items))
		return importDoneMsg{stats: stats}
	}
}

func (m Model) exportGoodreads() tea.Cmd {
	m.state = ViewProgress
	m.message = "Exporting to Goodreads CSV..."

	return func() tea.Msg {
		inPath := m.inputs[0].Value()
		outPath := m.inputs[1].Value()
		shelf := m.inputs[2].Value()
		added := m.inputs[3].Value()

		if shelf == "" {
			shelf = viper.GetString("DEFAULT_SHELF")
		}

		timezone := viper.GetString("TZ")

		library, err := util.LoadLibrary(m.fs, inPath)
		if err != nil {
			return exportDoneMsg{err: err}
		}

		opts := export.ExportOptions{
			Shelf:     shelf,
			DateAdded: added,
			Timezone:  timezone,
		}

		if err := export.ExportGoodreadsCSV(m.fs, library, outPath, opts); err != nil {
			return exportDoneMsg{err: err}
		}

		return exportDoneMsg{count: len(library.Items)}
	}
}

func (m Model) createAudibleInputs() []textinput.Model {
	inputs := make([]textinput.Model, 2)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "/path/to/openaudible/books.json"
	inputs[0].CharLimit = 256
	inputs[0].Width = 60
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "json (or csv)"
	inputs[1].CharLimit = 10
	inputs[1].Width = 20
	
	return inputs
}

func (m Model) createKindleInputs() []textinput.Model {
	inputs := make([]textinput.Model, 2)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "/path/to/My Clippings.txt"
	inputs[0].CharLimit = 256
	inputs[0].Width = 60
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "/path/to/notebook/html/dir (optional)"
	inputs[1].CharLimit = 256
	inputs[1].Width = 60
	
	return inputs
}

func (m Model) createStorytelInputs() []textinput.Model {
	inputs := make([]textinput.Model, 2)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "/path/to/storytel_history.csv"
	inputs[0].CharLimit = 256
	inputs[0].Width = 60
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "csv (or json)"
	inputs[1].CharLimit = 10
	inputs[1].Width = 20
	
	return inputs
}

func (m Model) createGoodreadsInputs() []textinput.Model {
	outDir := viper.GetString("OUT_DIR")
	
	inputs := make([]textinput.Model, 4)
	
	inputs[0] = textinput.New()
	inputs[0].Placeholder = filepath.Join(outDir, "library.json")
	inputs[0].SetValue(filepath.Join(outDir, "library.json"))
	inputs[0].CharLimit = 256
	inputs[0].Width = 60
	
	inputs[1] = textinput.New()
	inputs[1].Placeholder = filepath.Join(outDir, "goodreads_import.csv")
	inputs[1].SetValue(filepath.Join(outDir, "goodreads_import.csv"))
	inputs[1].CharLimit = 256
	inputs[1].Width = 60
	
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "read (or to-read, currently-reading)"
	inputs[2].CharLimit = 30
	inputs[2].Width = 30
	
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "YYYY-MM-DD (optional, defaults to today)"
	inputs[3].CharLimit = 10
	inputs[3].Width = 30
	
	return inputs
}

// Key bindings helper
var _ = key.NewBinding()
