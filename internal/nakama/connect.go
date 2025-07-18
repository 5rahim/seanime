package nakama

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/huin/goupnp/dcps/internetgateway1"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

type UPnPClient interface {
	GetExternalIPAddress() (string, error)
	AddPortMapping(string, uint16, string, uint16, string, bool, string, uint32) error
	DeletePortMapping(string, uint16, string) error
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Port forwarding
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EnablePortForwarding(port int) (string, error) {
	return enablePortForwarding(port)
}

// enablePortForwarding enables port forwarding for a given port and returns the address.
func enablePortForwarding(port int) (string, error) {
	// Try IGDv2 first, then fallback to IGDv1
	ip, err := addPortMappingIGD(func() ([]UPnPClient, error) {
		clients, _, err := internetgateway2.NewWANIPConnection1Clients()
		if err != nil {
			return nil, err
		}
		upnpClients := make([]UPnPClient, len(clients))
		for i, client := range clients {
			upnpClients[i] = client
		}
		return upnpClients, nil
	}, port)
	if err != nil {
		ip, err = addPortMappingIGD(func() ([]UPnPClient, error) {
			clients, _, err := internetgateway1.NewWANIPConnection1Clients()
			if err != nil {
				return nil, err
			}
			upnpClients := make([]UPnPClient, len(clients))
			for i, client := range clients {
				upnpClients[i] = client
			}
			return upnpClients, nil
		}, port)
		if err != nil {
			return "", fmt.Errorf("failed to add port mapping: %w", err)
		}
	}

	return fmt.Sprintf("http://%s:%d", ip, port), nil
}

func disablePortForwarding(port int) error {
	// Try to remove port mapping from both IGDv2 and IGDv1
	err1 := removePortMappingIGD(func() ([]UPnPClient, error) {
		clients, _, err := internetgateway2.NewWANIPConnection1Clients()
		if err != nil {
			return nil, err
		}
		upnpClients := make([]UPnPClient, len(clients))
		for i, client := range clients {
			upnpClients[i] = client
		}
		return upnpClients, nil
	}, port)
	err2 := removePortMappingIGD(func() ([]UPnPClient, error) {
		clients, _, err := internetgateway1.NewWANIPConnection1Clients()
		if err != nil {
			return nil, err
		}
		upnpClients := make([]UPnPClient, len(clients))
		for i, client := range clients {
			upnpClients[i] = client
		}
		return upnpClients, nil
	}, port)

	// Return error only if both failed
	if err1 != nil && err2 != nil {
		return fmt.Errorf("failed to remove port mapping from IGDv2: %v, IGDv1: %v", err1, err2)
	}

	return nil
}

// addPortMappingIGD adds a port mapping using the provided client factory and returns the external IP
func addPortMappingIGD(clientFactory func() ([]UPnPClient, error), port int) (string, error) {
	clients, err := clientFactory()
	if err != nil {
		return "", err
	}

	for _, client := range clients {
		// Get external IP address
		externalIP, err := client.GetExternalIPAddress()
		if err != nil {
			continue // Try next client
		}

		// Add port mapping
		err = client.AddPortMapping(
			"",               // NewRemoteHost (empty for any)
			uint16(port),     // NewExternalPort
			"TCP",            // NewProtocol
			uint16(port),     // NewInternalPort
			"127.0.0.1",      // NewInternalClient (localhost)
			true,             // NewEnabled
			"Seanime Nakama", // NewPortMappingDescription
			uint32(3600),     // NewLeaseDuration (1 hour)
		)
		if err != nil {
			continue // Try next client
		}

		return externalIP, nil // Success
	}

	return "", fmt.Errorf("no working UPnP clients found")
}

// removePortMappingIGD removes a port mapping using the provided client factory
func removePortMappingIGD(clientFactory func() ([]UPnPClient, error), port int) error {
	clients, err := clientFactory()
	if err != nil {
		return err
	}

	for _, client := range clients {
		err = client.DeletePortMapping(
			"",           // NewRemoteHost (empty for any)
			uint16(port), // NewExternalPort
			"TCP",        // NewProtocol
		)
		if err != nil {
			continue // Try next client
		}

		return nil // Success
	}

	return fmt.Errorf("no working UPnP clients found")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Join code (shelved)
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EncryptJoinCode(ip string, port int, password string) (string, error) {
	plainText := fmt.Sprintf("%s:%d", ip, port)

	// Derive 256-bit key from password
	key := sha256.Sum256([]byte(password))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecryptJoinCode(code, password string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil {
		return "", err
	}

	key := sha256.Sum256([]byte(password))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
