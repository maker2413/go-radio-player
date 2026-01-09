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
	"github.com/joho/godotenv"
	"github.com/maker2413/go-radio-player/internal/icyreader"
	"github.com/maker2413/go-radio-player/internal/player"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	stationName := os.Getenv("STATION_NAME")
	if stationName == "" {
		stationName = "Unknown"
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
		defer f.Close()
	}

	log.Info("Connecting to stream...")
	client := &http.Client{Timeout: 0}
	req, _ := http.NewRequest("GET", streamURL, nil)
	req.Header.Set("Icy-MetaData", "1")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	// Check if the server actually supports ICY metadata
	icyIntStr := resp.Header.Get("icy-metaint")
	if icyIntStr == "" {
		log.Fatal("Server did not return icy-metaint. This might not be a direct Icecast stream.")
	}

	// Get the interval from headers
	metaint, _ := strconv.Atoi(resp.Header.Get("icy-metaint"))
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
	defer streamer.Close()

	log.Info("Initializing speaker...")
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	ap, err := player.NewAudioPlayer(format.SampleRate, streamer, stationName, titleChan)
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(ap, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
