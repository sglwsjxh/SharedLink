package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SentProgressMsg struct {
	SentBytes  int64
	TotalBytes int64
	Done       bool
	Error      error
}

var sentStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2).
	Width(50)

var sentTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39"))

type SendModel struct {
	filename  string
	filesize  int64
	progress  progress.Model
	sentBytes int64
	startTime time.Time
	done      bool
	err       error
	quitting  bool
}

func NewSendModel(filename string, filesize int64) SendModel {
	p := progress.New()
	p.FullColor = "39"
	p.EmptyColor = "240"
	p.ShowPercentage = true
	p.Width = 46
	return SendModel{
		filename:  filename,
		filesize:  filesize,
		progress:  p,
		startTime: time.Now(),
	}
}

func (m SendModel) Init() tea.Cmd {
	return nil
}

func (m SendModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case SentProgressMsg:
		if msg.Error != nil {
			m.done = true
			m.err = msg.Error
			return m, tea.Sequence(tea.Tick(0, func(t time.Time) tea.Msg { return tea.Quit }))
		}
		m.sentBytes = msg.SentBytes
		if msg.Done {
			m.done = true
			return m, tea.Sequence(tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				return tea.Quit
			}))
		}
		ratio := float64(msg.SentBytes) / float64(msg.TotalBytes)
		cmd := m.progress.SetPercent(ratio)
		return m, cmd

	case progress.FrameMsg:
		pm, cmd := m.progress.Update(msg)
		m.progress = pm.(progress.Model)
		return m, cmd

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 12
		if m.progress.Width < 20 {
			m.progress.Width = 20
		}
		return m, nil
	}

	return m, nil
}

func (m SendModel) View() string {
	if m.quitting {
		return ""
	}

	speed := formatSpeed(m.sentBytes, m.startTime)
	eta := formatETA(m.sentBytes, m.filesize, m.startTime)

	content := fmt.Sprintf("%s\n\n", sentTitleStyle.Render("Sending: "+m.filename))
	content += fmt.Sprintf("Size: %s\n", formatBytes(m.filesize))
	content += fmt.Sprintf("Sent: %s / %s\n", formatBytes(m.sentBytes), formatBytes(m.filesize))
	content += fmt.Sprintf("Speed: %s\n\n", speed)
	content += m.progress.View()

	if m.done {
		if m.err != nil {
			content += fmt.Sprintf("\n\nError: %v", m.err)
		} else {
			content += "\n\nTransfer complete!"
		}
	} else {
		content += fmt.Sprintf("\n\nETA: %s", eta)
	}

	content += "\n\nPress 'q' or Ctrl+C to cancel"

	return sentStyle.Render(content)
}
