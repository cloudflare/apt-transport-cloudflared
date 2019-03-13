package access

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/cloudflared/apt-transport-cloudflare/apt/exec"
)

func TestHelperProcess(t *testing.T) {
	exec.MockExecHelper()
}

func testParseServiceToken(t *testing.T, val, id, secret string, errors bool) {
	tok, err := ParseServiceToken(val)
	if err != nil {
		if !errors {
			t.Errorf("Unexpected Error: %v", err)
		}
		return
	}

	if errors {
		t.Errorf("Expected error, got %v", tok)
		return
	}

	if tok.ID != id {
		t.Errorf("Expected parsed ID of '%s', got '%s'", id, tok.ID)
	}

	if tok.Secret != secret {
		t.Errorf("Expected parsed Secret of '%s', got '%s'", secret, tok.Secret)
	}
}

func TestParseServiceToken(t *testing.T) {
	testParseServiceToken(t, "Hello\nWorld", "Hello", "World", false)
	testParseServiceToken(t, "\nHello\nWorld\n", "Hello", "World", false)
	testParseServiceToken(t, "World", "", "", true)
	testParseServiceToken(t, "Hello\n", "", "", true)
	testParseServiceToken(t, "\nWorld", "", "", true)
	testParseServiceToken(t, "", "", "", true)
}

func testFindUserTokenError(ctx context.Context, t *testing.T, uri *url.URL, errmsg string) {
	out, err := FindUserToken(ctx, uri, true)
	if err == nil {
		t.Errorf(errmsg, out)
	}
}

func TestFindUserToken(t *testing.T) {
	// Mock out the exec.CommandContext
	fb := exec.NewMockBuilder("TestHelperProcess", exec.MockEntry{Sleep: time.Second})
	exec.Builder = fb

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	uri, _ := url.Parse("https://httpbin.org/get")

	// Error, hang in first command
	testFindUserTokenError(ctx, t, uri, "Expected error due to hung process, got %v")

	// Error, hang in second command
	fb.Reset(exec.MockEntry{}, exec.MockEntry{Sleep: time.Second})
	testFindUserTokenError(ctx, t, uri, "Expected error due to hung process, got %v")

	// Not testing hangs, so meh
	ctx = context.Background()

	// Error - Bad exit from `cloudflared access login`
	fb.Reset(exec.MockEntry{ExitCode: 1})
	testFindUserTokenError(ctx, t, uri, "Expected error due to bad exit code, got %v")

	// Error - Bad exit from `cloudflared access token`
	fb.Reset(exec.MockEntry{}, exec.MockEntry{ExitCode: 1})
	testFindUserTokenError(ctx, t, uri, "Expected error due to bad exit code, got %v")

	// Error - Bad output from `cloudflared access token`
	fb.Reset(exec.MockEntry{}, exec.MockEntry{Output: "Unable to fetch token"})
	testFindUserTokenError(ctx, t, uri, "Expected error due to bad output, got %v")

	// Valid, return "token-1a24fd"
	output := "token-1a24fd"
	fb.Reset(exec.MockEntry{}, exec.MockEntry{Output: output})
	out, err := FindUserToken(ctx, uri, true)
	if err != nil {
		t.Errorf("Unexpected error getting user token: %v", err)
	} else if out.JWT != output {
		t.Errorf("Bad parsed JWT; expected \"%s\", got \"%s\"", output, out.JWT)
	}
}
