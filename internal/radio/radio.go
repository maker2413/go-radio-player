package radio

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// A model can be more or less any type of data. It holds all the data for a
// program, so often it's a struct. For this simple example, however, all
// we'll need is a simple string.
type radio struct {
	Station          string
	CurrentSongTitle string
}

// Init optionally returns an initial command we should run. In this case we
// want to start the timer.
func (r radio) Init() tea.Cmd {
	return tick
}

// Update is called when messages are received. The idea is that you inspect the
// message and send back an updated model accordingly. You can also return
// a command, which is a function that performs I/O and returns a message.
func (r radio) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return r, tea.Quit
		case "ctrl+z":
			return r, tea.Suspend
		}

	case tickMsg:
		return r, tick
	}
	return r, nil
}

// View returns a string based on data in the model. That string which will be
// rendered to the terminal.
func (r radio) View() string {
	return fmt.Sprintf("Station: %s\nSongTitle: %s\nPress Ctrl+c to quit\n", r.Station, r.CurrentSongTitle)
}

// Messages are events that we respond to in our Update function. This
// particular one indicates that the timer has ticked.
type tickMsg time.Time

func tick() tea.Msg {
	time.Sleep(time.Second)
	return tickMsg{}
}
