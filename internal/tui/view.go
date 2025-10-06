package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2)
)

// View renders the current view.
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	switch m.state {
	case ViewHome:
		return m.viewHome()
	case ViewAudibleImport:
		return m.viewAudibleImport()
	case ViewKindleImport:
		return m.viewKindleImport()
	case ViewStorytelImport:
		return m.viewStorytelImport()
	case ViewGoodreadsExport:
		return m.viewGoodreadsExport()
	case ViewProgress:
		return m.viewProgress()
	case ViewResult:
		return m.viewResult()
	}

	return ""
}

func (m Model) viewHome() string {
	return m.list.View()
}

func (m Model) viewAudibleImport() string {
	title := titleStyle.Render("Import from Audible")
	
	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Export file path:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Format:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab: next field • enter: import • esc: back"))

	return boxStyle.Render(b.String())
}

func (m Model) viewKindleImport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Import from Kindle"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("My Clippings.txt path:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Notebook HTML directory (optional):"))
	b.WriteString("\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab: next field • enter: import • esc: back"))

	return boxStyle.Render(b.String())
}

func (m Model) viewStorytelImport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Import from Storytel"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Export file path:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Format:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab: next field • enter: import • esc: back"))

	return boxStyle.Render(b.String())
}

func (m Model) viewGoodreadsExport() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Export to Goodreads CSV"))
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Input library JSON:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[0].View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Output CSV path:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[1].View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Shelf:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[2].View())
	b.WriteString("\n\n")

	b.WriteString(labelStyle.Render("Date added:"))
	b.WriteString("\n")
	b.WriteString(m.inputs[3].View())
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("tab: next field • enter: export • esc: back"))

	return boxStyle.Render(b.String())
}

func (m Model) viewProgress() string {
	var b strings.Builder

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(m.message)
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("Please wait..."))

	return boxStyle.Render(b.String())
}

func (m Model) viewResult() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(errorStyle.Render("✗ " + m.message))
	} else {
		b.WriteString(successStyle.Render("✓ " + m.message))
	}
	b.WriteString("\n\n")

	if m.stats != "" {
		b.WriteString(m.stats)
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("enter: back to menu • o: open output folder • q: quit"))

	return boxStyle.Render(b.String())
}
