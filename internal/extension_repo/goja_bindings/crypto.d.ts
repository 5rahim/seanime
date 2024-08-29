declare class WordArray {
    toString(encoder?: CryptoJSEncoder): string;
}

// CryptoJS supports AES-128, AES-192, and AES-256. It will pick the variant by the size of the key you pass in. If you use a passphrase,
// then it will generate a 256-bit key.
declare class CryptoJS {
    static AES: {
        encrypt: (message: string, key: string | ArrayBuffer, cfg?: AESConfig) => WordArray;
        decrypt: (message: string | WordArray, key: string | ArrayBuffer, cfg?: AESConfig) => WordArray;
    }
    static enc: {
        Utf8: CryptoJSEncoder;
        Base64: CryptoJSEncoder;
        Hex: CryptoJSEncoder;
        Latin1: CryptoJSEncoder;
        Utf16: CryptoJSEncoder;
        Utf16LE: CryptoJSEncoder;
    }
}

declare interface AESConfig {
    iv?: ArrayBuffer;
}

declare class CryptoJSEncoder {
    stringify(input: ArrayBuffer): string;

    parse(input: string): ArrayBuffer;
}
