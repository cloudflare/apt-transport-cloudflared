package main

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Fake Exec stuff here - lets us test code which uses os/exec

// An entry in the FakeExec list. This specifies how the process should execute
type fakeExecEntry struct {
	ExitCode int
	Sleep    time.Duration
	Output   string
	Extra    []string
}

func (f *fakeExecEntry) GetEnvVars() []string {
	if f == nil {
		return []string{"GO_HELPER_EXIT_CODE=0"}
	}

	var s []string

	// Add exit code if non-zero
	if f.ExitCode != 0 {
		s = append(s, fmt.Sprintf("GO_HELPER_EXIT_CODE=%d", f.ExitCode))
	}

	// Add sleep time if greater than 0
	if f.Sleep > 0 {
		s = append(s, fmt.Sprintf("GO_HELPER_SLEEP=%d", f.Sleep))
	}

	// Add os.Stdout output
	s = append(s, "GO_HELPER_OUTPUT="+f.Output)
	//if f.Output != "" {
	//    s = append(s, "GO_HELPER_OUTPUT=" + f.Output)
	//}

	// Any extra environment variables
	if len(f.Extra) > 0 {
		s = append(s, f.Extra...)
	}
	return s
}

// Controlling variables for the fakeExec function
type fakeExecData struct {
	Index   int
	Entries []fakeExecEntry
	Helper  string
	T       *testing.T
}

// The global fakeExec data
var fakeExec fakeExecData

// Replacement for exec.CommandContext
func fakeExecCtxCommand(ctx context.Context, command string, args ...string) *exec.Cmd {
	cs := []string{fmt.Sprintf("-test.run=%s", fakeExec.Helper), "--", command}
	cs = append(cs, args...)
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}

	var entry *fakeExecEntry = nil

	if fakeExec.Index >= 0 && fakeExec.Index < len(fakeExec.Entries) {
		entry = &fakeExec.Entries[fakeExec.Index]
		fakeExec.Index += 1
	}
	cmd.Env = append(cmd.Env, entry.GetEnvVars()...)
	return cmd
}

// Replacement for exec.Command
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	return fakeExecCtxCommand(context.Background(), command, args...)
}

// Initialize the global fakeExec structure
// TODO: Accept a pointer to the variable holding the function and overwrite it
func fakeExecInit(t *testing.T, helper string, entries ...fakeExecEntry) {
	fakeExec.T = t
	fakeExec.Helper = helper
	fakeExec.Index = 0
	fakeExec.Entries = entries
}

// Reset the fakeExec entries with those given here
func fakeExecReset(entries ...fakeExecEntry) {
	fakeExec.Index = 0
	fakeExec.Entries = entries
}

// The helper process - the fakeExec replacement functions call this test
// handler to implement the subprocess replacement.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	sleep_str := os.Getenv("GO_HELPER_SLEEP")
	if sleep_str != "" {
		dur, err := strconv.ParseInt(sleep_str, 10, 32)
		if err == nil && dur > 0 {
			time.Sleep(time.Duration(dur))
		}
	}

	output_str := os.Getenv("GO_HELPER_OUTPUT")
	if output_str != "" {
		fmt.Print(output_str)
	}

	exitcodestr := os.Getenv("GO_HELPER_EXIT_CODE")
	var exitcode int64 = 0
	if exitcodestr != "" {
		exitcode, _ = strconv.ParseInt(exitcodestr, 10, 32)
	}
	os.Exit(int(exitcode))
}

func TestParseConfig(t *testing.T) {
	var output strings.Builder
	input := strings.NewReader("601 Configuration\nApt::Acquire::Something Blah\n\n")
	method, _ := NewCloudflaredMethod(&output, bufio.NewReader(input), "")
	// Replace the client with something that doesn't actually do anything
	msg := NewMessage(601, "Configuration", Field{"key", "value"})
	err := method.ParseConfig(msg)
	if err != nil {
		t.Errorf("Expected no error with valid config, got %v", err)
	}
}

func TestGetToken(t *testing.T) {
	// Mock out the exec.CommandContext
	fakeExecInit(t, "TestHelperProcess", fakeExecEntry{Sleep: time.Second})
	makeCommand = fakeExecCtxCommand

	var output strings.Builder
	input := strings.NewReader("601 Configuration\nApt::Acquire::Something: Blah\n\n")
	method, _ := NewCloudflaredMethod(&output, bufio.NewReader(input), "")

	// Error - Hang in first command
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	uri, _ := url.Parse("https://httpbin.org/get")
	out, err := method.GetToken(ctx, uri)
	if err == nil {
		// It should have hung
		t.Errorf("Expected error due to hung process `cloudflared access login %s`, got %v", uri.String(), out)
	}

	// Error - Hang second command
	fakeExecReset(fakeExecEntry{ExitCode: 0}, fakeExecEntry{Sleep: time.Second})
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	out, err = method.GetToken(ctx, uri)
	if err == nil {
		t.Errorf("Expected error due to hung process `cloudflared access token --app %s`, got %v", uri.String(), out)
	}

	// We're guaranteed not to hang now, so just use Background()
	ctx = context.Background()

	// Error - Bad exit from `cloudflared access login`
	fakeExecReset(fakeExecEntry{ExitCode: 1})
	out, err = method.GetToken(ctx, uri)
	if err == nil {
		t.Errorf("Expected error due to bad exit code, got %v", out)
	}

	// Error - Bad exit from `cloudflared access token --app ...`
	fakeExecReset(fakeExecEntry{}, fakeExecEntry{ExitCode: 1})
	out, err = method.GetToken(ctx, uri)
	if err == nil {
		t.Errorf("Expected error due to bad exit code, got %v", out)
	}

	fakeExecReset(fakeExecEntry{}, fakeExecEntry{Output: "Unable to fetch token"})

	// Valid: return "token"
	fakeExecReset(fakeExecEntry{}, fakeExecEntry{Output: "token-1a24fd"})
	h, err := method.GetToken(ctx, uri)
	if err != nil {
		t.Errorf("Expected valid headers, got error: %v", err)
	}

	if len(h) != 1 {
		t.Errorf("Expected exactly 1 header addition, got %v", h)
	}
}
