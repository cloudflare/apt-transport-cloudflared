package apt

import (
	"bytes"
	"net/url"
	"strings"
)

// URLWriter is a io.Writer which only writes URLS
type URLWriter struct {
	writer *MessageWriter
	buffer bytes.Buffer
	prefix string
}

// NewURLWriter creates a new URLWriter instance.
func NewURLWriter(w *MessageWriter, prefix string) *URLWriter {
	return &URLWriter{
		writer: w,
		buffer: bytes.Buffer{},
		prefix: prefix,
	}
}

// Write implements the io.Writer interface.
//
// Write buffers data until it hits a newline, at which point it checks if the
// buffered data is a URL. If it is, it writes that URL to the MessageWriter
// instance it was created with, along with the prefix prepended.
func (uw *URLWriter) Write(data []byte) (int, error) {
	start := 0
	for i, b := range data {
		// If we hit a newline, copy data[start:i] into the buffer and commit
		if b == byte('\n') {
			uw.buffer.Write(data[start:i])
			uw.commit()
			start = i + 1
		}
	}

	uw.buffer.Write(data[start:])

	return len(data), nil
}

// commit takes a line from the buffer and checks if it is a url. If it is,
// then it writes the URL using the message writer.
func (uw *URLWriter) commit() {
	// Pull data from buffer, trim off spaces
	line := strings.TrimSpace(uw.buffer.String())
	if strings.HasPrefix(line, "http") {
		// Check if we have a URL, and if so print it
		_, err := url.Parse(line)
		if err == nil {
			uw.writer.Log(uw.prefix + line)
		}
	}

	// Clear the buffer regardless
	uw.buffer.Reset()
}
