package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type RecvProgressMsg struct {
	ReceivedBytes int64
	TotalBytes    int64
	Done          bool
	Error         error
}

type RecvModel struct {
	Filename      string
	Filesize      int64
	ReceivedBytes int64
	Done          bool
	Err           error
	hasInfo       bool
	startTime     time.Time
	quitting      bool
}

func NewRecvModel() RecvModel {
	return RecvModel{
		startTime: time.Now(),
	}
}

func (m RecvModel) Init() tea.Cmd {
	return nil
}

func (m RecvModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case RecvProgressMsg:
		if msg.Done {
			m.Done = true
			return m, tea.Quit
		}
		if msg.Error != nil {
			m.Done = true
			m.Err = msg.Error
			return m, tea.Quit
		}
		if !m.hasInfo && msg.TotalBytes > 0 {
			m.Filesize = msg.TotalBytes
			m.hasInfo = true
		}
		m.ReceivedBytes = msg.ReceivedBytes
		return m, nil
	}

	return m, nil
}

func (m RecvModel) View() string {
	if m.quitting {
		return ""
	}

	content := "\n"
	content += "  ╔═══════════════════════════════════════════╗\n"
	content += "  ║                SharedLink                 ║\n"
	content += "  ╚═══════════════════════════════════════════╝\n"
	content += "\n"

	if !m.hasInfo {
		content += "  Connecting to sender...\n"
		content += "\n"
		content += "Press Ctrl+C to cancel"
		return content
	}

	speed := formatSpeed(m.ReceivedBytes, m.startTime)
	title := "Receiving"
	if m.Filename != "" {
		title = "Receiving: " + m.Filename
	}

	content += fmt.Sprintf("  %s\n", title)
	content += fmt.Sprintf("  Size: %s\n", formatBytes(m.Filesize))
	content += fmt.Sprintf("  Received: %s / %s\n", formatBytes(m.ReceivedBytes), formatBytes(m.Filesize))
	content += fmt.Sprintf("  Speed: %s\n", speed)
	content += "\n"
	ratio := float64(m.ReceivedBytes) / float64(m.Filesize)
	content += "  " + renderProgressBar(ratio) + "\n"
	content += "\n"

	if m.Done {
		if m.Err != nil {
			content += fmt.Sprintf("  Error: %v\n", m.Err)
		} else {
			content += "  Download complete!\n"
		}
	} else {
		eta := formatETA(m.ReceivedBytes, m.Filesize, m.startTime)
		content += fmt.Sprintf("  ETA: %s\n", eta)
	}

	content += "Press Ctrl+C to cancel"

	return content
}
