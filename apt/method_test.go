package apt

import (
	"bufio"
	"strings"
	"testing"
)

func TestParseConfig(t *testing.T) {
	var output strings.Builder
	input := strings.NewReader("601 Configuration\nApt::Acquire::Something Blah\n\n")
	method, _ := NewCloudflaredMethod(&output, bufio.NewReader(input))
	// Replace the client with something that doesn't actually do anything
	msg := NewMessage(601, "Configuration", Field{"key", "value"})
	err := method.ParseConfig(msg)
	if err != nil {
		t.Errorf("Expected no error with valid config, got %v", err)
	}
}
