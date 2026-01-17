package player

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
)

const (
	tickRate     = time.Second
	titlePadding = "        "
)

type audioPlayer struct {
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
	width               int
	height              int
}

type tickMsg time.Time

func NewAudioPlayer(
	sampleRate beep.SampleRate,
	streamer beep.StreamSeeker,
	stationName string,
	titleChan <-chan string,
	maxDisplayedTitleSize int,
) (*audioPlayer, error) {
	volume := &effects.Volume{Streamer: streamer, Base: 2, Volume: -2.0}

	if maxDisplayedTitleSize <= 0 {
		maxDisplayedTitleSize = len(titlePadding)
	}

	return &audioPlayer{sampleRate: sampleRate,
		streamer:            streamer,
		volume:              volume,
		stationName:         stationName,
		titleChan:           titleChan,
		currentTitle:        "",
		displayedTitle:      "",
		maxDisplayTitleSize: maxDisplayedTitleSize,
	}, nil
}

func (ap *audioPlayer) Init() tea.Cmd {
	ap.Play()

	return tick()
}

func (ap *audioPlayer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		ap.width = msg.Width
		ap.height = msg.Height
	case tickMsg:
		ap.titleMutex.Lock()
		if len(ap.displayedTitle)-len(titlePadding) > ap.maxDisplayTitleSize {
			ap.displayedTitle = leftShiftString(ap.displayedTitle)
		}
		ap.titleMutex.Unlock()

		if ap.titleUpdate() {
			return ap, tea.Batch(
				tick(),
				tea.SetWindowTitle("♫ "+ap.stationName+" ~ "+ap.currentTitle+" ♫"),
			)
		}

		return ap, tick()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			speaker.Close()
			return ap, tea.Quit
		case "w":
			speaker.Lock()
			ap.volumeMutex.Lock()
			ap.volume.Volume += 0.1
			ap.volumeMutex.Unlock()
			speaker.Unlock()
		case "s":
			speaker.Lock()
			ap.volumeMutex.Lock()
			ap.volume.Volume -= 0.1
			ap.volumeMutex.Unlock()
			speaker.Unlock()
		case "m", " ":
			speaker.Lock()
			ap.volumeMutex.Lock()
			ap.volume.Silent = !ap.volume.Silent
			ap.volumeMutex.Unlock()
			speaker.Unlock()
		}
	}
	return ap, nil
}

func (ap *audioPlayer) View() string {
	ap.titleMutex.Lock()
	title := ap.displayedTitle
	ap.titleMutex.Unlock()
	if len(title) > ap.maxDisplayTitleSize {
		title = title[:ap.maxDisplayTitleSize]
	}

	output := "Station: " + ap.stationName +
		"\nSong: " + title +
		"\n\nPress Space or M to mute" +
		"\nUse W and S to control Volume" +
		"\nTo exit press escape, or Ctrl+c"

	style := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFFFFF")).Padding(1, 2)

	return lipgloss.Place(
		ap.width, ap.height, lipgloss.Center, lipgloss.Center, style.Render(output))
}

func (ap *audioPlayer) titleUpdate() bool {
	select {
	case title := <-ap.titleChan:
		if len(title) > 0 {
			ap.titleMutex.Lock()
			ap.currentTitle = title
			ap.displayedTitle = ap.currentTitle + titlePadding
			ap.titleMutex.Unlock()
		}

		return true
	default:
		return false
	}
}

func (ap *audioPlayer) Play() {
	speaker.Play(ap.volume)
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
