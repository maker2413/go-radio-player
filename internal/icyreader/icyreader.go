package icyreader

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// bufferedReadCloser bridges bufio.Reader and io.Closer
type bufferedReadCloser struct {
	*bufio.Reader
	io.Closer
}

func NewWrappedReader(reader *IcyReader, size int) *bufferedReadCloser {
	return &bufferedReadCloser{
		Reader: bufio.NewReaderSize(reader, size),
		Closer: reader,
	}
}

func NewIcyReader(body io.ReadCloser, metaint int) *IcyReader {
	return &IcyReader{
		body:        body,
		interval:    metaint,
		bytesToNext: metaint,
	}
}

// IcyReader wraps a stream and extracts metadata at intervals
type IcyReader struct {
	body        io.ReadCloser
	interval    int
	bytesToNext int
}

func (r *IcyReader) Read(p []byte) (n int, err error) {
	if r.bytesToNext == 0 {
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
