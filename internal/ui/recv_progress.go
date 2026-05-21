package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type RecvProgressMsg struct {
	ReceivedBytes int64
	TotalBytes    int64
	Done          bool
	Error         error
}

var recvStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")).
	Padding(1, 2).
	Width(50)

var recvTitleStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("39"))

var recvConnectingStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("240"))

type RecvModel struct {
	filename      string
	filesize      int64
	hasInfo       bool
	progress      progress.Model
	receivedBytes int64
	startTime     time.Time
	done          bool
	err           error
	quitting      bool
}

func NewRecvModel() RecvModel {
	p := progress.New()
	p.FullColor = "39"
	p.EmptyColor = "240"
	p.ShowPercentage = true
	p.Width = 46
	return RecvModel{
		progress:  p,
		startTime: time.Now(),
	}
}

func (m RecvModel) Init() tea.Cmd {
	return nil
}

func (m RecvModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case RecvProgressMsg:
		if msg.Error != nil {
			m.done = true
			m.err = msg.Error
			return m, tea.Sequence(tea.Tick(0, func(t time.Time) tea.Msg {
				return tea.Quit
			}))
		}
		if !m.hasInfo && msg.TotalBytes > 0 {
			m.filesize = msg.TotalBytes
			m.hasInfo = true
		}
		m.receivedBytes = msg.ReceivedBytes
		if msg.Done {
			m.done = true
			return m, tea.Sequence(tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
				return tea.Quit
			}))
		}
		ratio := float64(msg.ReceivedBytes) / float64(msg.TotalBytes)
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

func (m RecvModel) View() string {
	if m.quitting {
		return ""
	}

	speed := formatSpeed(m.receivedBytes, m.startTime)

	if !m.hasInfo {
		content := recvConnectingStyle.Render("Connecting to sender...") + "\n\n"
		content += scanHelpStyle.Render("Press 'q' or Ctrl+C to cancel")
		return content
	}

	eta := formatETA(m.receivedBytes, m.filesize, m.startTime)

	title := "Receiving file"
	if m.filename != "" {
		title = "Receiving: " + m.filename
	}
	content := fmt.Sprintf("%s\n\n", recvTitleStyle.Render(title))
	content += fmt.Sprintf("Size: %s\n", formatBytes(m.filesize))
	content += fmt.Sprintf("Received: %s / %s\n", formatBytes(m.receivedBytes), formatBytes(m.filesize))
	content += fmt.Sprintf("Speed: %s\n\n", speed)
	content += m.progress.View()

	if m.done {
		if m.err != nil {
			content += fmt.Sprintf("\n\nError: %v", m.err)
		} else {
			content += "\n\nDownload complete!"
		}
	} else {
		content += fmt.Sprintf("\n\nETA: %s", eta)
	}

	content += "\n\nPress 'q' or Ctrl+C to cancel"

	return recvStyle.Render(content)
}
