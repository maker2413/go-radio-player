package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gopxl/beep"
	"github.com/gopxl/beep/mp3"
	"github.com/gopxl/beep/speaker"
)

// IcyReader wraps a stream and extracts metadata at intervals
type IcyReader struct {
	body        io.ReadCloser
	interval    int
	bytesToNext int
}

func (r *IcyReader) Read(p []byte) (n int, err error) {
	if r.bytesToNext == 0 {
		// Time to read metadata!
		var lengthByte [1]byte
		if _, err := io.ReadFull(r.body, lengthByte[:]); err != nil {
			return 0, err
		}

		metaLen := int(lengthByte[0]) * 16
		if metaLen > 0 {
			metaData := make([]byte, metaLen)
			if _, err := io.ReadFull(r.body, metaData); err != nil {
				return 0, err
			}
			parseMetadata(string(metaData))
		}
		r.bytesToNext = r.interval
	}

	// Limit read to not cross into the next metadata block
	limit := len(p)
	if limit > r.bytesToNext {
		limit = r.bytesToNext
	}

	n, err = r.body.Read(p[:limit])
	r.bytesToNext -= n
	return n, err
}

func (r *IcyReader) Close() error {
	return r.body.Close()
}

func parseMetadata(meta string) {
	// Format is usually: StreamTitle='Song Name - Artist';
	if strings.Contains(meta, "StreamTitle='") {
		parts := strings.Split(meta, "StreamTitle='")
		title := strings.Split(parts[1], "';")[0]
		fmt.Printf("\n--- NOW PLAYING: %s ---\n", title)
	}
}

// bufferedReadCloser bridges bufio.Reader and io.Closer
type bufferedReadCloser struct {
	*bufio.Reader
	io.Closer
}

func main() {
	// 1. The direct stream URL (Icecast mount point)
	streamURL := ""

	// 2. Fetch the stream via HTTP
	log.Println("Connecting to stream...")
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
	log.Printf("Metadata interval: %d bytes", metaint)

	icyReader := &IcyReader{
		body:        resp.Body,
		interval:    metaint,
		bytesToNext: metaint,
	}

	wrappedReader := &bufferedReadCloser{
		Reader: bufio.NewReaderSize(icyReader, 32*1024), // 32KB buffer
		Closer: icyReader,
	}

	log.Println("Decoding MP3 stream...")
	// We wrap in bufio to ensure the decoder gets enough data to identify the format
	streamer, format, err := mp3.Decode(wrappedReader)
	if err != nil {
		log.Fatalf("Failed to decode MP3: %v", err)
	}
	defer streamer.Close()

	log.Println("Initializing speaker...")
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	log.Println("Playing! Press Ctrl+C to quit.")
	<-done
}
