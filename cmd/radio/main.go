package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
	"github.com/joho/godotenv"
	"github.com/maker2413/go-radio/internal/icyreader"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 1. The direct stream URL (Icecast mount point)
	streamURL := os.Getenv("STREAM_URL")
	if streamURL == "" {
		log.Fatal("STREAM_URL not set")
	}

	debug := os.Getenv("DEBUG")
	if debug == "" {
		log.Fatal("DEBUG not set")
	}

	logCaller := false
	logLevel := 0
	if debug == "true" {
		logCaller = true
		logLevel = -4
	}

	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    logCaller,
		ReportTimestamp: true,
		TimeFormat:      time.RFC3339,
		Level:           log.Level(logLevel),
	})

	// 2. Fetch the stream via HTTP
	logger.Info("Connecting to stream...")
	client := &http.Client{Timeout: 0}
	req, _ := http.NewRequest("GET", streamURL, nil)
	req.Header.Set("Icy-MetaData", "1")

	resp, err := client.Do(req)
	if err != nil {
		logger.Fatalf("Failed to connect: %v", err)
	}
	defer resp.Body.Close()

	// Check if the server actually supports ICY metadata
	icyIntStr := resp.Header.Get("icy-metaint")
	if icyIntStr == "" {
		logger.Fatal("Server did not return icy-metaint. This might not be a direct Icecast stream.")
	}

	// Get the interval from headers
	metaint, _ := strconv.Atoi(resp.Header.Get("icy-metaint"))
	logger.Debug("Metadata interval: %d bytes", metaint)

	reader := icyreader.NewIcyReader(resp.Body, metaint)

	wrappedReader := icyreader.NewWrappedReader(reader, 32*1024) // 32KB buffer

	logger.Debug("Decoding MP3 stream...")
	// We wrap in bufio to ensure the decoder gets enough data to identify the format
	streamer, format, err := mp3.Decode(wrappedReader)
	if err != nil {
		logger.Fatalf("Failed to decode MP3: %v", err)
	}
	defer streamer.Close()

	logger.Info("Initializing speaker...")
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	logger.Print("Playing! Press Ctrl+C to quit.")
	<-done
}
