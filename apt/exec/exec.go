package exec

import (
	"context"
	osexec "os/exec"
)

// CmdBuilder is an interface to a type which builds commands.
//
// This is used to allow mocking exec.Command().Run().
type CmdBuilder interface {
	Command(cmd string, args ...string) *osexec.Cmd
	CommandContext(ctx context.Context, cmd string, args ...string) *osexec.Cmd
}

var (
	Builder = RealBuilder()
)

// Command creates a command with the global builder.
func Command(cmd string, args ...string) *osexec.Cmd {
	return Builder.Command(cmd, args...)
}

// CommandContext creates a command with the global builder using the given
// context.
func CommandContext(ctx context.Context, cmd string, args ...string) *osexec.Cmd {
	return Builder.CommandContext(ctx, cmd, args...)
}
