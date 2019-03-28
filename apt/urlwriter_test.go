package apt

import (
	"bytes"
	"testing"
)

func TestURLWriter(t *testing.T) {
	output := &bytes.Buffer{}
	urlw := NewURLWriter(output, "URL: ")

	urlw.Write([]byte("Hello World\n"))
	if output.String() != "" {
		t.Errorf("Expected no output from non-url input")
	}

	urlw.Write([]byte("Header line\nhttps://httpbin.org/get\nTrailing line\n"))
	if output.String() != "\rURL: https://httpbin.org/get\n" {
		t.Errorf("Unexpected output: %q\n", output.String())
	}
}
