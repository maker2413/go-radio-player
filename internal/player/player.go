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

type audioPlayer struct {
	sampleRate   beep.SampleRate
	streamer     beep.StreamSeeker
	volume       *effects.Volume
	volumeMutex  sync.Mutex
	stationName  string
	titleChan    <-chan string
	currentTitle string
	width        int
	height       int
}

type tickMsg time.Time

func NewAudioPlayer(
	sampleRate beep.SampleRate,
	streamer beep.StreamSeeker,
	stationName string,
	titleChan <-chan string,
) (*audioPlayer, error) {
	volume := &effects.Volume{Streamer: streamer, Base: 2, Volume: -2.0}

	return &audioPlayer{sampleRate: sampleRate,
		streamer:     streamer,
		volume:       volume,
		stationName:  stationName,
		titleChan:    titleChan,
		currentTitle: ""}, nil
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
		if ap.titleUpdate() {
			return ap, tea.Batch(
				tick(),
				tea.SetWindowTitle(ap.stationName),
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
	output := "Song: " + ap.currentTitle +
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
		ap.currentTitle = title
		return true
	default:
		return false
	}
}

func (ap *audioPlayer) Play() {
	speaker.Play(ap.volume)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
