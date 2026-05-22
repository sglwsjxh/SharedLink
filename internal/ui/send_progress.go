package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type SentProgressMsg struct {
	SentBytes  int64
	TotalBytes int64
	Done       bool
	Error      error
}

type SendModel struct {
	Filename  string
	Filesize  int64
	SentBytes int64
	Done      bool
	Err       error
	startTime time.Time
	quitting  bool
}

func NewSendModel(filename string, filesize int64) SendModel {
	return SendModel{
		Filename:  filename,
		Filesize:  filesize,
		startTime: time.Now(),
	}
}

func (m SendModel) Init() tea.Cmd {
	return nil
}

func (m SendModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case SentProgressMsg:
		if msg.Done {
			m.Done = true
			return m, tea.Quit
		}
		if msg.Error != nil {
			m.Done = true
			m.Err = msg.Error
			return m, tea.Quit
		}
		m.SentBytes = msg.SentBytes
		return m, nil
	}

	return m, nil
}

func (m SendModel) View() string {
	if m.quitting {
		return ""
	}

	speed := formatSpeed(m.SentBytes, m.startTime)

	content := "\n"
	content += "  ╔═══════════════════════════════════════════╗\n"
	content += "  ║                SharedLink                 ║\n"
	content += "  ╚═══════════════════════════════════════════╝\n"
	content += "\n"
	content += fmt.Sprintf("  Sending: %s\n", m.Filename)
	content += fmt.Sprintf("  Size:    %s\n", formatBytes(m.Filesize))
	content += fmt.Sprintf("  Sent:    %s / %s\n", formatBytes(m.SentBytes), formatBytes(m.Filesize))
	content += fmt.Sprintf("  Speed:   %s\n", speed)
	content += "\n"
	ratio := float64(m.SentBytes) / float64(m.Filesize)
	content += "  " + renderProgressBar(ratio) + "\n"
	content += "\n"

	if m.Done {
		if m.Err != nil {
			content += fmt.Sprintf("  Error: %v\n", m.Err)
		} else {
			content += "  Transfer complete!\n"
		}
	} else {
		eta := formatETA(m.SentBytes, m.Filesize, m.startTime)
		content += fmt.Sprintf("  ETA: %s\n", eta)
	}

	content += "Press Ctrl+C to cancel"

	return content
}
