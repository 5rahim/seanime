package goja_bindings

import (
	"seanime/internal/util"
	"testing"
	"time"

	"github.com/dop251/goja"
	gojabuffer "github.com/dop251/goja_nodejs/buffer"
	gojarequire "github.com/dop251/goja_nodejs/require"
	"github.com/stretchr/testify/require"
)

func TestGojaCrypto(t *testing.T) {
	vm := goja.New()
	defer vm.ClearInterrupt()

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindCrypto(vm)
	BindConsole(vm, util.NewLogger())

	_, err := vm.RunString(`
async function run() {

    try {

        console.log("\nTesting Buffer encoding/decoding")

        const originalString = "Hello, this is a string to encode!"
        const base64String = Buffer.from(originalString).toString("base64")

        console.log("Original String:", originalString)
        console.log("Base64 Encoded:", base64String)

        const decodedString = Buffer.from(base64String, "base64").toString("utf-8")

        console.log("Base64 Decoded:", decodedString)

    }
    catch (e) {
        console.error(e)
    }

    try {

        console.log("\nTesting AES")

        let message = "seanime"
        let key = CryptoJS.enc.Utf8.parse("secret key")


        console.log("Message:", message)

        let encrypted = CryptoJS.AES.encrypt(message, key)
        console.log("Encrypted without IV:", encrypted) // map[iv toString]
        console.log("Encrypted.toString():", encrypted.toString()) // AoHrnhJfbRht2idLHM82WdkIEpRbXufnA6+ozty9fbk=
        console.log("Encrypted.toString(CryptoJS.enc.Base64):", encrypted.toString(CryptoJS.enc.Base64)) // AoHrnhJfbRht2idLHM82WdkIEpRbXufnA6+ozty9fbk=

        let decrypted = CryptoJS.AES.decrypt(encrypted, key)
        console.log("Decrypted:", decrypted.toString(CryptoJS.enc.Utf8))

        let iv = CryptoJS.enc.Utf8.parse("3134003223491201")
        encrypted = CryptoJS.AES.encrypt(message, key, { iv: iv })
        console.log("Encrypted with IV:", encrypted) // map[iv toString]

        decrypted = CryptoJS.AES.decrypt(encrypted, key)
        console.log("Decrypted without IV:", decrypted.toString(CryptoJS.enc.Utf8))

        decrypted = CryptoJS.AES.decrypt(encrypted, key, { iv: iv })
        console.log("Decrypted with IV:", decrypted.toString(CryptoJS.enc.Utf8)) // seanime

    }
    catch (e) {
        console.error(e)
    }

    try {

        console.log("\nTesting encoders")

        console.log("")
        let a = CryptoJS.enc.Utf8.parse("Hello, World!")
        console.log("Base64 Parsed:", a)
        let b = CryptoJS.enc.Base64.stringify(a)
        console.log("Base64 Stringified:", b)
        let c = CryptoJS.enc.Base64.parse(b)
        console.log("Base64 Parsed:", c)
        let d = CryptoJS.enc.Utf8.stringify(c)
        console.log("Base64 Stringified:", d)
        console.log("")

        let words = CryptoJS.enc.Latin1.parse("Hello, World!")
        console.log("Latin1 Parsed:", words)
        let latin1 = CryptoJS.enc.Latin1.stringify(words)
        console.log("Latin1 Stringified", latin1)

        words = CryptoJS.enc.Hex.parse("48656c6c6f2c20576f726c6421")
        console.log("Hex Parsed:", words)
        let hex = CryptoJS.enc.Hex.stringify(words)
        console.log("Hex Stringified", hex)

        words = CryptoJS.enc.Utf8.parse("𔭢")
        console.log("Utf8 Parsed:", words)
        let utf8 = CryptoJS.enc.Utf8.stringify(words)
        console.log("Utf8 Stringified", utf8)

        words = CryptoJS.enc.Utf16.parse("Hello, World!")
        console.log("Utf16 Parsed:", words)
        let utf16 = CryptoJS.enc.Utf16.stringify(words)
        console.log("Utf16 Stringified", utf16)

        words = CryptoJS.enc.Utf16LE.parse("Hello, World!")
        console.log("Utf16LE Parsed:", words)
        utf16 = CryptoJS.enc.Utf16LE.stringify(words)
        console.log("Utf16LE Stringified", utf16)
    }
    catch (e) {
        console.error("Error:", e)
    }
}
`)
	require.NoError(t, err)

	runFunc, ok := goja.AssertFunction(vm.Get("run"))
	require.True(t, ok)

	ret, err := runFunc(goja.Undefined())
	require.NoError(t, err)

	promise := ret.Export().(*goja.Promise)

	for promise.State() == goja.PromiseStatePending {
		time.Sleep(10 * time.Millisecond)
	}

	if promise.State() == goja.PromiseStateRejected {
		err := promise.Result()
		t.Fatal(err)
	}
}

func TestGojaCryptoOpenSSL(t *testing.T) {
	vm := goja.New()
	defer vm.ClearInterrupt()

	registry := new(gojarequire.Registry)
	registry.Enable(vm)
	gojabuffer.Enable(vm)
	BindCrypto(vm)
	BindConsole(vm, util.NewLogger())

	_, err := vm.RunString(`
async function run() {

    try {

        console.log("\nTesting Buffer encoding/decoding")

        const payload = "U2FsdGVkX19ZanX9W5jQGgNGOIOBGxhY6gxa1EHnRi3yHL8Ml4cMmQeryf9p04N12VuOjiBas21AcU0Ypc4dB4AWOdc9Cn1wdA2DuQhryUonKYHwV/XXJ53DBn1OIqAvrIAxrN8S2j9Rk5z/F/peu1Kk/d3m82jiKvhTWQcxDeDW8UzCMZbbFnm4qJC3k19+PD5Pal5sBcVTGRXNCpvSSpYb56FcP9Xs+3DyBWhNUqJuO+Wwm3G1J5HhklxCWZ7tcn7TE5Y8d5ORND7t51Padrw4LgEOootqHtfHuBVX6EqlvJslXt0kFgcXJUIO+hw0q5SJ+tiS7o/2OShJ7BCk4XzfQmhFJdBJYGjQ8WPMHYzLuMzDkf6zk2+m7YQtUTXx8SVoLXFOt8gNZeD942snGrWA5+CdYveOfJ8Yv7owoOueMzzYqr5rzG7GVapVI0HzrA24LR4AjRDICqTsJEy6Yg=="
		const key = "6315b93606d60f48c964b67b14701f3848ef25af01296cf7e6a98c9460e1d2ac"
        console.log("Original String:", payload)

        const decrypted = CryptoJS.AES.decrypt(payload, key)

		console.log("Decrypted:", decrypted.toString(CryptoJS.enc.Utf8))

    }
    catch (e) {
        console.error(e)
    }

}
`)
	require.NoError(t, err)

	runFunc, ok := goja.AssertFunction(vm.Get("run"))
	require.True(t, ok)

	ret, err := runFunc(goja.Undefined())
	require.NoError(t, err)

	promise := ret.Export().(*goja.Promise)

	for promise.State() == goja.PromiseStatePending {
		time.Sleep(10 * time.Millisecond)
	}

	if promise.State() == goja.PromiseStateRejected {
		err := promise.Result()
		t.Fatal(err)
	}
}
