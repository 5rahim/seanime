/// <reference path="../crypto.d.ts" />
/// <reference path="../buffer.d.ts" />

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
