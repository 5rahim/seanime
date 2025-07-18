package nakama

import (
	"testing"
)

func TestPortForwarding(t *testing.T) {
	// Test port forwarding for port 43211
	address, err := EnablePortForwarding(43211)
	if err != nil {
		t.Logf("Port forwarding failed (expected if no UPnP router available): %v", err)
		t.Skip("No UPnP support available")
		return
	}

	t.Logf("Port forwarding enabled successfully: %s", address)

	// Clean up - disable the port forwarding
	err = disablePortForwarding(43211)
	if err != nil {
		t.Logf("Warning: Failed to clean up port forwarding: %v", err)
	}
}

func TestEncryptJoinCode(t *testing.T) {
	code, err := EncryptJoinCode("127.0.0.1", 4000, "password")
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("code: %s", code)

	addr, err := DecryptJoinCode(code, "password")
	if err != nil {
		t.Fatal(err)
	}

	if addr != "127.0.0.1:4000" {
		t.Fatal("invalid decrypted code")
	}
}
