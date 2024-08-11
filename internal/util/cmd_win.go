//go:build windows

package util

import (
	"context"
	"os/exec"
	"syscall"
)

// NewCmd creates a new exec.Cmd object with the given arguments.
// Since for Windows, the app is built as a GUI application, we need to hide the console windows launched when running commands.
func NewCmd(arg string, args ...string) *exec.Cmd {
	//cmdPrompt := "C:\\Windows\\system32\\cmd.exe"
	//cmdArgs := append([]string{"/c", arg}, args...)
	//cmd := exec.Command(cmdPrompt, cmdArgs...)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd := exec.Command(arg, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000,
		//HideWindow:    true,
	}
	return cmd
}

func NewCmdCtx(ctx context.Context, arg string, args ...string) *exec.Cmd {
	//cmdPrompt := "C:\\Windows\\system32\\cmd.exe"
	//cmdArgs := append([]string{"/c", arg}, args...)
	//cmd := exec.CommandContext(ctx, cmdPrompt, cmdArgs...)
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd := exec.CommandContext(ctx, arg, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000,
		//HideWindow:    true,
	}
	return cmd
}
