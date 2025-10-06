package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

// ViewState represents the current view in the TUI.
type ViewState int

const (
	ViewHome ViewState = iota
	ViewAudibleImport
	ViewKindleImport
	ViewStorytelImport
	ViewGoodreadsExport
	ViewProgress
	ViewResult
)

// Model is the main TUI model.
type Model struct {
	fs            afero.Fs
	state         ViewState
	list          list.Model
	inputs        []textinput.Model
	spinner       spinner.Model
	progress      progress.Model
	focused       int
	err           error
	message       string
	stats         string
	quitting      bool
	width         int
	height        int
}

// Run starts the TUI application.
func Run() error {
	viper.SetDefault("OUT_DIR", "./out")
	viper.SetDefault("TZ", "America/Sao_Paulo")
	viper.SetDefault("DEFAULT_SHELF", "read")

	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func initialModel() Model {
	// Create home menu list
	items := []list.Item{
		menuItem{title: "Import from Audible", desc: "Import OpenAudible JSON or CSV export"},
		menuItem{title: "Import from Kindle", desc: "Import My Clippings.txt or Notebook HTML"},
		menuItem{title: "Import from Storytel", desc: "Import Storytel history export"},
		menuItem{title: "Export to Goodreads CSV", desc: "Generate Goodreads-compatible CSV"},
		menuItem{title: "Open Output Folder", desc: "Open the output directory"},
		menuItem{title: "Quit", desc: "Exit the application"},
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 80, 20)
	l.Title = "media2goodreads"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	prog := progress.New(progress.WithDefaultGradient())

	return Model{
		fs:       afero.NewOsFs(),
		state:    ViewHome,
		list:     l,
		spinner:  sp,
		progress: prog,
	}
}

// menuItem implements list.Item interface.
type menuItem struct {
	title string
	desc  string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }
