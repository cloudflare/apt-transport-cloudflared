package exec

import (
	"context"
	"fmt"
	"os"
	osexec "os/exec"
	"strconv"
	"time"
)

// MockEntry describes the execution of one mock subprocess.
type MockEntry struct {
	// ExitCode is the exit code the process should exit with
	ExitCode int

	// Sleep is how long the process should sleep to simulate a hang.
	Sleep time.Duration

	// Output is a string to write to os.Stdout
	Output string

	// Extra is an array of extra environment variables to pass.
	Extra []string
}

// GetEnvVars returns environment variables for the test process.
func (me *MockEntry) GetEnvVars() []string {
	if me == nil {
		return []string{"MOCK_EXEC_EXIT_CODE=0"}
	}

	var s []string
	if me.ExitCode != 0 {
		s = append(s, fmt.Sprintf("MOCK_EXEC_EXIT_CODE=%d", me.ExitCode))
	}

	if me.Sleep > 0 {
		s = append(s, fmt.Sprintf("MOCK_EXEC_SLEEP=%d", me.Sleep))
	}

	s = append(s, "MOCK_EXEC_OUTPUT="+me.Output)

	if len(me.Extra) > 0 {
		s = append(s, me.Extra...)
	}

	return s
}

// MockBuilder is a CmdBuilder which builds test Cmd instances.
type MockBuilder struct {
	Index   int
	Entries []MockEntry
	Helper  string
}

// NewMockBuilder creates a new builder which creates mock Commands.
func NewMockBuilder(helper string, entries ...MockEntry) *MockBuilder {
	return &MockBuilder{
		Helper:  helper,
		Entries: entries,
	}
}

// Command returns a test command with no context.
func (mb *MockBuilder) Command(cmd string, args ...string) *osexec.Cmd {
	return mb.CommandContext(context.Background(), cmd, args...)
}

// CommandContext returns a test command with the given context.
func (mb *MockBuilder) CommandContext(ctx context.Context, cmd string, args ...string) *osexec.Cmd {
	cs := []string{"-test.run=" + mb.Helper, "--", "cmd"}
	cs = append(cs, args...)

	command := osexec.CommandContext(ctx, os.Args[0], cs...)
	command.Env = []string{"MOCK_EXEC_HELPER_PROCESS=1"}

	var entry *MockEntry

	if mb.Index >= 0 && mb.Index < len(mb.Entries) {
		entry = &mb.Entries[mb.Index]
		mb.Index++
	}
	command.Env = append(command.Env, entry.GetEnvVars()...)
	return command
}

// Reset sets the index of the MockBuilder to 0 and changes the entry list.
func (mb *MockBuilder) Reset(entries ...MockEntry) {
	mb.Index = 0
	mb.Entries = entries
}

// MockExecHelper implements the logic for the helper process.
//
// This function should be called from a test stub in your program used
// exclusively for creating processes. The test stub shouldn't do anything
// else.
func MockExecHelper() {
	if os.Getenv("MOCK_EXEC_HELPER_PROCESS") != "1" {
		return
	}

	sleepstr := os.Getenv("MOCK_EXEC_SLEEP")
	if sleepstr != "" {
		dur, err := strconv.ParseInt(sleepstr, 10, 32)
		if err == nil && dur > 0 {
			time.Sleep(time.Duration(dur))
		}
	}

	out := os.Getenv("MOCK_EXEC_OUTPUT")
	if out != "" {
		fmt.Print(out)
	}

	var exitcode int64
	ecs := os.Getenv("MOCK_EXEC_EXIT_CODE")
	if ecs != "" {
		exitcode, _ = strconv.ParseInt(ecs, 10, 32)
	}
	os.Exit(int(exitcode))
}
