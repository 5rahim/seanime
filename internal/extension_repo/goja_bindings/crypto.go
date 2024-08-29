package goja_bindings

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/dop251/goja"
	"io"
)

type wordArray struct {
	vm   *goja.Runtime
	iv   []byte
	data []byte
}

func BindCrypto(vm *goja.Runtime) error {

	err := vm.Set("CryptoJS", map[string]interface{}{
		"AES": map[string]interface{}{
			"encrypt": cryptoAESEncryptFunc(vm),
			"decrypt": cryptoAESDecryptFunc(vm),
		},
		"enc": map[string]interface{}{
			"Utf8": map[string]interface{}{
				"parse":     cryptoEncUtf8ParseFunc(vm),
				"stringify": cryptoEncUtf8StringifyFunc(vm),
			},
			"Base64": map[string]interface{}{
				"parse":     cryptoEncBase64ParseFunc(vm),
				"stringify": cryptoEncBase64StringifyFunc(vm),
			},
			"Hex": map[string]interface{}{
				"parse":     cryptoEncHexParseFunc(vm),
				"stringify": cryptoEncHexStringifyFunc(vm),
			},
			"Latin1": map[string]interface{}{
				"parse":     cryptoEncLatin1ParseFunc(vm),
				"stringify": cryptoEncLatin1StringifyFunc(vm),
			},
			"Utf16": map[string]interface{}{
				"parse":     cryptoEncUtf16ParseFunc(vm),
				"stringify": cryptoEncUtf16StringifyFunc(vm),
			},
			"Utf16LE": map[string]interface{}{
				"parse":     cryptoEncUtf16LEParseFunc(vm),
				"stringify": cryptoEncUtf16LEStringifyFunc(vm),
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func newWordArrayGojaValue(vm *goja.Runtime, data []byte, iv []byte) goja.Value {
	wa := &wordArray{
		vm:   vm,
		iv:   iv,
		data: data,
	}
	// WordArray // Utf8
	// WordArray.toString(): string // Uses Base64.stringify
	// WordArray.toString(encoder: Encoder): string
	obj := vm.NewObject()
	obj.Prototype().Set("toString", wa.toStringFunc)
	obj.Set("toString", wa.toStringFunc)
	obj.Set("iv", iv)
	return obj
}

func (wa *wordArray) toStringFunc(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) == 0 {
		return wa.vm.ToValue(base64Stringify(wa.data))
	}

	encoder, ok := call.Argument(0).Export().(map[string]interface{})
	if !ok {
		panic(wa.vm.ToValue("TypeError: encoder parameter must be a CryptoJS.enc object"))
	}

	var ret string
	if f, ok := encoder["stringify"]; ok {
		if stringify, ok := f.(func(functionCall goja.FunctionCall) goja.Value); ok {
			ret = stringify(goja.FunctionCall{Arguments: []goja.Value{wa.vm.ToValue(wa.data)}}).String()
		} else {
			panic(wa.vm.ToValue("TypeError: encoder.stringify must be a function"))
		}
	} else {
		ret = string(wa.data)
	}

	return wa.vm.ToValue(ret)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// CryptoJS.AES.encrypt(message: string, key: string, cfg?: { iv: ArrayBuffer }): WordArray
func cryptoAESEncryptFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) (ret goja.Value) {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("TypeError: AES.encrypt requires at least 2 arguments"))
		}

		message := call.Argument(0).String()

		var keyBytes []byte
		switch call.Argument(1).Export().(type) {
		case string:
			key := call.Argument(1).String()
			keyBytes = adjustKeyLength([]byte(key))
		case []byte:
			keyBytes = call.Argument(1).Export().([]byte)
			keyBytes = adjustKeyLength(keyBytes)
		default:
			panic(vm.ToValue("TypeError: key parameter must be a string or an ArrayBuffer"))
		}

		usedRandomIV := false
		// Check if IV is provided
		var ivBytes []byte
		if len(call.Arguments) > 2 {
			cfg := call.Argument(2).Export().(map[string]interface{})
			var ok bool
			iv, ok := cfg["iv"].([]byte)
			if !ok {
				panic(vm.ToValue("TypeError: iv parameter must be an ArrayBuffer"))
			}
			ivBytes = iv
			if len(ivBytes) != aes.BlockSize {
				panic(vm.ToValue("TypeError: IV length must be equal to block size (16 bytes for AES)"))
			}
		} else {
			// Generate a random IV
			ivBytes = make([]byte, aes.BlockSize)
			if _, err := io.ReadFull(rand.Reader, ivBytes); err != nil {
				panic(vm.ToValue(fmt.Sprintf("Failed to generate IV: %v", err)))
			}
			usedRandomIV = true
		}

		defer func() {
			if r := recover(); r != nil {
				ret = vm.ToValue(fmt.Sprintf("Encryption failed: %v", r))
			}
		}()

		// Encrypt the message
		encryptedMessage := encryptAES(vm, message, keyBytes, ivBytes)

		if usedRandomIV {
			// Prepend the IV to the encrypted message
			encryptedMessage = append(ivBytes, encryptedMessage...)
		}

		return newWordArrayGojaValue(vm, encryptedMessage, ivBytes)
	}
}

// CryptoJS.AES.decrypt(encryptedMessage: string | WordArray, key: string, cfg?: { iv: ArrayBuffer }): WordArray
func cryptoAESDecryptFunc(vm *goja.Runtime) func(call goja.FunctionCall) goja.Value {
	return func(call goja.FunctionCall) (ret goja.Value) {
		if len(call.Arguments) < 2 {
			panic(vm.ToValue("TypeError: AES.decrypt requires at least 2 arguments"))
		}

		// Can be string or WordArray
		// If WordArray, String() will call WordArray.toString() which will return the base64 encoded string
		encryptedMessage := call.Argument(0).String()

		var keyBytes []byte
		switch call.Argument(1).Export().(type) {
		case string:
			key := call.Argument(1).String()
			keyBytes = adjustKeyLength([]byte(key))
		case []byte:
			keyBytes = call.Argument(1).Export().([]byte)
			keyBytes = adjustKeyLength(keyBytes)
		default:
			panic(vm.ToValue("TypeError: key parameter must be a string or an ArrayBuffer"))
		}

		var ivBytes []byte
		var cipherText []byte

		// If IV is provided in the third argument
		if len(call.Arguments) > 2 {
			cfg := call.Argument(2).Export().(map[string]interface{})
			var ok bool
			iv, ok := cfg["iv"].([]byte)
			if !ok {
				panic(vm.ToValue("TypeError: iv parameter must be an ArrayBuffer"))
			}
			ivBytes = iv
			if len(ivBytes) != aes.BlockSize {
				panic(vm.ToValue("TypeError: IV length must be equal to block size (16 bytes for AES)"))
			}
			var err error
			decodedMessage, err := base64.StdEncoding.DecodeString(encryptedMessage)
			if err != nil {
				panic(vm.ToValue(fmt.Sprintf("Failed to decode ciphertext: %v", err)))
			}
			cipherText = decodedMessage
		} else {
			// Decode the base64 encoded string
			decodedMessage, err := base64.StdEncoding.DecodeString(encryptedMessage)
			if err != nil {
				panic(vm.ToValue(fmt.Sprintf("Failed to decode ciphertext: %v", err)))
			}

			// Extract the IV from the beginning of the message
			ivBytes = decodedMessage[:aes.BlockSize]
			cipherText = decodedMessage[aes.BlockSize:]
		}

		// Decrypt the message
		decrypted := decryptAES(vm, cipherText, keyBytes, ivBytes)

		return newWordArrayGojaValue(vm, decrypted, ivBytes)
	}
}

// Adjusts the key length to match AES key length requirements (16, 24, or 32 bytes).
// If the key length is not 16, 24, or 32 bytes, it is hashed using SHA-256 and truncated to 32 bytes (AES-256).
func adjustKeyLength(keyBytes []byte) []byte {
	switch len(keyBytes) {
	case 16, 24, 32:
		// Valid AES key lengths: 16 bytes (AES-128), 24 bytes (AES-192), 32 bytes (AES-256)
		return keyBytes
	default:
		// Hash the key to 32 bytes (AES-256)
		hash := sha256.Sum256(keyBytes)
		return hash[:]
	}
}

func encryptAES(vm *goja.Runtime, message string, key []byte, iv []byte) (ret []byte) {
	defer func() {
		if r := recover(); r != nil {
			ret = nil
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(vm.ToValue(fmt.Sprintf("%v", err)))
	}

	messageBytes := []byte(message)
	messageBytes = pkcs7Padding(messageBytes, aes.BlockSize)

	cipherText := make([]byte, len(messageBytes))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, messageBytes)

	return cipherText
}

func decryptAES(vm *goja.Runtime, cipherText []byte, key []byte, iv []byte) (ret []byte) {
	defer func() {
		if r := recover(); r != nil {
			ret = nil
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(vm.ToValue(fmt.Sprintf("%v", err)))
	}

	plainText := make([]byte, len(cipherText))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plainText, cipherText)

	plainText = pkcs7Trimming(plainText)

	return plainText
}

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7Trimming(data []byte) []byte {
	length := len(data)
	up := int(data[length-1])
	return data[:(length - up)]
}
