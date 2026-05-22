package ui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"SharedLink/internal/discover"
)

type scanDoneMsg struct {
	entries []discover.Entry
	err     error
}

type ScanModel struct {
	ctx      context.Context
	entries  []discover.Entry
	cursor   int
	scanning bool
	done     bool
	err      error
	selected *discover.Entry
	quitting bool
}

var scanTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39"))

var scanHelpStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240"))

var scanCursorStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("39")).
	Bold(true)

var scanEntryStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("252"))

func NewScanModel(ctx context.Context) ScanModel {
	return ScanModel{
		ctx:      ctx,
		scanning: true,
	}
}

func (m ScanModel) Init() tea.Cmd {
	return m.scan
}

func (m ScanModel) scan() tea.Msg {
	entries, err := discover.Scan(m.ctx, 3*time.Second)
	return scanDoneMsg{entries: entries, err: err}
}

func (m ScanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "r", "R":
			m.scanning = true
			m.entries = nil
			m.cursor = 0
			m.done = false
			m.err = nil
			return m, m.scan
		case "up", "k":
			if m.done && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.done && m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "enter":
			if m.done && len(m.entries) > 0 {
				m.selected = &m.entries[m.cursor]
				m.quitting = true
				return m, tea.Quit
			}
		}
		return m, nil

	case scanDoneMsg:
		m.scanning = false
		m.done = true
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.entries = msg.entries
		}
		return m, nil
	}

	return m, nil
}

func (m ScanModel) View() string {
	if m.quitting {
		return ""
	}

	content := scanTitleStyle.Render("Scanning LAN for senders...") + "\n\n"

	if m.scanning {
		content += "Scanning...\n"
		content += scanHelpStyle.Render("Press Ctrl+C to quit")
		return content
	}

	if m.err != nil {
		content += fmt.Sprintf("Error: %v\n", m.err)
		content += scanHelpStyle.Render("Press 'r' to retry, Ctrl+C to quit")
		return content
	}

	if len(m.entries) == 0 {
		content += "No senders found on the network.\n"
		content += scanHelpStyle.Render("Press 'r' to rescan, Ctrl+C to quit")
		return content
	}

	content += fmt.Sprintf("Found %d sender(s):\n\n", len(m.entries))
	for i, entry := range m.entries {
		cursor := "  "
		if i == m.cursor {
			cursor = scanCursorStyle.Render("▸ ")
		}
		line := fmt.Sprintf("%s%s  %s  %s",
			cursor,
			scanEntryStyle.Render(entry.Hostname),
			entry.Addr,
			formatBytes(entry.FileSize),
		)
		content += line + "\n"
		if i == m.cursor {
			content += fmt.Sprintf("   File: %s\n", m.entries[i].FileName)
		}
	}

	content += "\n" + scanHelpStyle.Render("↑↓ select  •  enter confirm  •  r rescan  •  Ctrl+C quit")
	return content
}

func (m ScanModel) SelectedAddr() string {
	if m.selected == nil {
		return ""
	}
	return m.selected.Addr
}

func (m ScanModel) SelectedEntry() *discover.Entry {
	return m.selected
}
