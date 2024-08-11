package extension_repo

import (
	"seanime/internal/util"
	"strings"
	"testing"
)

func TestExtensionID(t *testing.T) {

	tests := []struct {
		id       string
		expected bool
	}{
		{"my-extension", true},
		{"my-extension-", false},
		{"-my-extension", false},
		{"my-extension-1", true},
		{"my.extension", false},
		{"my_extension", false},
	}

	for _, test := range tests {
		if isValidExtensionIDString(test.id) != test.expected {
			t.Errorf("isValidExtensionID(%v) != %v", test.id, test.expected)
		}
	}

}

func TestReplacePackageName(t *testing.T) {
	extensionPackageName := "ext_" + util.GenerateCryptoID()

	payload := `package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"`

	newPayload := ReplacePackageName(payload, extensionPackageName)

	if strings.Contains(newPayload, "package main") {
		t.Errorf("ReplacePackageName failed")
	}

	t.Log(newPayload)
}
