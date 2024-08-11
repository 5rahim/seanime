//go:build !windows

package util

import (
	"context"
	"os/exec"
)

func NewCmd(arg string, args ...string) *exec.Cmd {
	if len(args) == 0 {
		return exec.Command(arg)
	}
	return exec.Command(arg, args...)
}

func NewCmdCtx(ctx context.Context, arg string, args ...string) *exec.Cmd {
	if len(args) == 0 {
		return exec.CommandContext(ctx, arg)
	}
	return exec.CommandContext(ctx, arg, args...)
}
