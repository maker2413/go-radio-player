package player

import (
	"fmt"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/effects"
	"github.com/gopxl/beep/speaker"
)

type audioPlayer struct {
	sampleRate   beep.SampleRate
	streamer     beep.StreamSeeker
	volume       *effects.Volume
	stationName  string
	titleChan    <-chan string
	currentTitle string
}

func NewAudioPlayer(
	sampleRate beep.SampleRate,
	streamer beep.StreamSeeker,
	stationName string,
	titleChan <-chan string,
) (*audioPlayer, error) {
	volume := &effects.Volume{Streamer: streamer, Base: 2, Volume: -2.0}

	return &audioPlayer{sampleRate, streamer, volume, stationName, titleChan, ""}, nil
}

func drawTextLine(screen tcell.Screen, x, y int, s string, style tcell.Style) {
	for _, r := range s {
		screen.SetContent(x, y, r, nil, style)
		x++
	}
}

func (ap *audioPlayer) Play() {
	speaker.Play(ap.volume)
}

func (ap *audioPlayer) Draw(screen tcell.Screen) {
	mainStyle := tcell.StyleDefault.
		Background(tcell.NewHexColor(0x473437)).
		Foreground(tcell.NewHexColor(0xD7D8A2))
	statusStyle := mainStyle.
		Foreground(tcell.NewHexColor(0xDDC074)).
		Bold(true)

	screen.Fill(' ', mainStyle)

	drawTextLine(screen, 0, 0, "Welcome to the Speedy Player!", mainStyle)
	drawTextLine(screen, 0, 1, "Press [ESC] to quit.", mainStyle)
	drawTextLine(screen, 0, 2, "Press [SPACE] to pause/resume.", mainStyle)
	drawTextLine(screen, 0, 3, "Use keys W or S to control volume.", mainStyle)

	speaker.Lock()
	volume := ap.volume.Volume
	speaker.Unlock()

	songTitle := ap.currentTitle
	volumeStatus := fmt.Sprintf("%.1f", volume+2.0)

	drawTextLine(screen, 0, 5, "Station:", mainStyle)
	drawTextLine(screen, 9, 5, ap.stationName, statusStyle)

	drawTextLine(screen, 0, 7, "Now Playing:", mainStyle)
	drawTextLine(screen, 13, 7, songTitle, statusStyle)

	drawTextLine(screen, 0, 8, "Volume:", mainStyle)
	drawTextLine(screen, 13, 8, volumeStatus, statusStyle)
}

func (ap *audioPlayer) Handle(event tcell.Event) (changed, quit bool) {
	switch event := event.(type) {
	case *tcell.EventKey:
		if event.Key() == tcell.KeyESC {
			return false, true
		}

		if event.Key() != tcell.KeyRune {
			return false, false
		}

		switch unicode.ToLower(event.Rune()) {
		case 's':
			speaker.Lock()
			ap.volume.Volume -= 0.1
			speaker.Unlock()
			return true, false

		case 'w':
			speaker.Lock()
			ap.volume.Volume += 0.1
			speaker.Unlock()
			return true, false
		}
		return false, false
	}
	return false, false
}

func (ap *audioPlayer) CheckForTitleUpdate() bool {
	select {
	case title := <-ap.titleChan:
		ap.currentTitle = title
		return true
	default:
		return false
	}
}
