package tui

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
	"github.com/maker2413/shellpod/internal/stack"
)

const (
	tickRate     = time.Second
	titlePadding = "        "
)

type tickMsg time.Time

type model struct {
	sampleRate          beep.SampleRate
	streamer            beep.StreamSeeker
	volume              *effects.Volume
	volumeMutex         sync.Mutex
	stationName         string
	titleChan           <-chan string
	titleMutex          sync.Mutex
	currentTitle        string
	displayedTitle      string
	maxDisplayTitleSize int
	currentPage         string
	pageHistory         stack.Stack[string]
	width               int
	height              int
}

func NewModel(
	sampleRate beep.SampleRate,
	streamer beep.StreamSeeker,
	stationName string,
	titleChan <-chan string,
	maxDisplayedTitleSize int,
) (tea.Model, error) {
	volume := &effects.Volume{Streamer: streamer, Base: 2, Volume: -2.0}

	if maxDisplayedTitleSize <= 0 {
		maxDisplayedTitleSize = len(titlePadding)
	}

	return &model{sampleRate: sampleRate,
		streamer:            streamer,
		volume:              volume,
		stationName:         stationName,
		titleChan:           titleChan,
		currentTitle:        "",
		displayedTitle:      "",
		maxDisplayTitleSize: maxDisplayedTitleSize,
		currentPage:         "home",
		pageHistory:         stack.Stack[string]{},
	}, nil
}

func (m *model) Init() tea.Cmd {
	m.Play()

	return tick()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.titleMutex.Lock()
		if len(m.displayedTitle)-len(titlePadding) > m.maxDisplayTitleSize {
			m.displayedTitle = leftShiftString(m.displayedTitle)
		}
		m.titleMutex.Unlock()

		if m.titleUpdate() {
			return m, tea.Batch(
				tick(),
				tea.SetWindowTitle("♫ "+m.stationName+" ~ "+m.currentTitle+" ♫"),
			)
		}

		return m, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			speaker.Close()
			return m, tea.Quit
		case "w":
			speaker.Lock()
			m.volumeMutex.Lock()
			m.volume.Volume += 0.1
			m.volumeMutex.Unlock()
			speaker.Unlock()
		case "s":
			speaker.Lock()
			m.volumeMutex.Lock()
			m.volume.Volume -= 0.1
			m.volumeMutex.Unlock()
			speaker.Unlock()
		case "m", " ":
			speaker.Lock()
			m.volumeMutex.Lock()
			m.volume.Silent = !m.volume.Silent
			m.volumeMutex.Unlock()
			speaker.Unlock()
		}
	}
	return m, nil
}

func (m *model) View() string {
	m.titleMutex.Lock()
	title := m.displayedTitle
	m.titleMutex.Unlock()
	if len(title) > m.maxDisplayTitleSize {
		title = title[:m.maxDisplayTitleSize]
	}

	output := "Station: " + m.stationName +
		"\nSong: " + title +
		"\n\nPress Space or M to mute" +
		"\nUse W and S to control Volume" +
		"\nTo exit press escme, or Ctrl+c"

	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFFFFF")).Padding(1, 2)

	return lipgloss.Place(
		m.width, m.height, lipgloss.Center, lipgloss.Center, style.Render(output))
}

func (m *model) titleUpdate() bool {
	select {
	case title := <-m.titleChan:
		if len(title) > 0 {
			m.titleMutex.Lock()
			m.currentTitle = title
			m.displayedTitle = m.currentTitle + titlePadding
			m.titleMutex.Unlock()
		}

		return true
	default:
		return false
	}
}

func (m *model) Play() {
	speaker.Play(m.volume)
}

func leftShiftString(s string) string {
	if len(s) <= 1 {
		return s
	}

	b := make([]byte, len(s))
	copy(b, s[1:])
	b[len(s)-1] = s[0]
	return string(b)
}

func tick() tea.Cmd {
	return tea.Tick(tickRate, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
