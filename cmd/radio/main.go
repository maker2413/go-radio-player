package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/maker2413/go-radio-player/internal/config"
	"github.com/maker2413/go-radio-player/internal/icyreader"
	"github.com/maker2413/go-radio-player/internal/player"
)

func main() {
	config, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	streamURL := os.Getenv("STREAM_URL")
	if streamURL == "" {
		log.Fatal("STREAM_URL not set")
	}

	debug := os.Getenv("DEBUG")
	if debug == "true" {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			log.Fatal(err)
		}
		defer func() {
			err = f.Close()
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	log.Info("Connecting to stream...")
	client := &http.Client{Timeout: 0}
	req, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Icy-MetaData", "1")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// Check if the server actually supports ICY metadata
	icyIntStr := resp.Header.Get("icy-metaint")
	if icyIntStr == "" {
		log.Fatal("Server did not return icy-metaint. This might not be a direct Icecast stream.")
	}

	// Get the interval from headers
	metaint, err := strconv.Atoi(resp.Header.Get("icy-metaint"))
	if err != nil {
		log.Fatal(err)
	}
	log.Debug("Metadata interval: %d bytes", metaint)

	reader := icyreader.NewIcyReader(resp.Body, metaint)
	titleChan := make(chan string, 10)
	reader.TitleChan = titleChan

	wrappedReader := icyreader.NewWrappedReader(reader, 32*1024) // 32KB buffer

	log.Debug("Decoding MP3 stream...")
	// We wrap in bufio to ensure the decoder gets enough data to identify the format
	streamer, format, err := mp3.Decode(wrappedReader)
	if err != nil {
		log.Fatalf("Failed to decode MP3: %v", err)
	}
	defer func() {
		err = streamer.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Info("Initializing speaker...")
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal("Failed to initialize speaker:", err)
	}

	ap, err := player.NewAudioPlayer(format.SampleRate, streamer, config.StationName, titleChan)
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(ap, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
