//go:build windows

package util

import (
	"syscall"
)

func HideFile(path string) (string, error) {
	defer HandlePanicInModuleThen("HideFile", func() {})

	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}
	err = syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return "", err
	}
	return path, nil
}
