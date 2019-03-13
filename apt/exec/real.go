package exec

import (
	"context"
	osexec "os/exec"
)

type realbuilder struct {
}

func RealBuilder() CmdBuilder {
	return realbuilder{}
}

func (rb realbuilder) Command(cmd string, args ...string) *osexec.Cmd {
	return osexec.Command(cmd, args...)
}

func (rb realbuilder) CommandContext(ctx context.Context, cmd string, args ...string) *osexec.Cmd {
	return osexec.CommandContext(ctx, cmd, args...)
}
