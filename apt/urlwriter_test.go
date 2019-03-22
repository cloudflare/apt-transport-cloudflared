package apt

import (
	"bytes"
	"testing"
)

func TestURLWriter(t *testing.T) {
	output := &bytes.Buffer{}
	mwriter := NewMessageWriter(output)
	urlw := NewURLWriter(mwriter, "URL: ")

	urlw.Write([]byte("Hello World\n"))
	if output.String() != "" {
		t.Errorf("Expected no output from non-url input")
	}

	urlw.Write([]byte("Header line\nhttps://httpbin.org/get\nTrailing line\n"))
	if output.String() != "101 Log\nMessage: URL: https://httpbin.org/get\n\n" {
		t.Errorf("Unexpected output: %s\n", output.String())
	}
}
