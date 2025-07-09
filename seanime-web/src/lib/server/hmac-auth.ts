interface TokenClaims {
    endpoint: string
    iat: number // issued at (unix timestamp)
    exp: number // expires at (unix timestamp)
}

class HMACAuth {
    private secret: string
    private ttl: number

    constructor(secret: string, ttl: number) {
        this.secret = secret
        this.ttl = ttl
    }

    async generateToken(endpoint: string): Promise<string> {
        const now = Math.floor(Date.now() / 1000)
        const claims: TokenClaims = {
            endpoint,
            iat: now,
            exp: now + this.ttl,
        }

        const claimsJSON = JSON.stringify(claims)

        // Encode claims as base64
        const claimsB64 = btoa(claimsJSON)
            .replace(/\+/g, "-")
            .replace(/\//g, "_")
            .replace(/=/g, "")

        // Generate HMAC signature
        const signature = await this.generateHMACSignature(claimsB64)

        // Return token in format: claims.signature
        return `${claimsB64}.${signature}`
    }

    generateQueryParam(endpoint: string, symbol?: string): Promise<string> {
        return this.generateToken(endpoint).then(token => {
            const sym = symbol || "?"
            return `${sym}token=${encodeURIComponent(token)}`
        })
    }

    private async generateHMACSignature(data: string): Promise<string> {
        // Convert secret to ArrayBuffer
        const encoder = new TextEncoder()
        const secretBuffer = encoder.encode(this.secret)
        const dataBuffer = encoder.encode(data)

        // Import key for HMAC
        const key = await crypto.subtle.importKey(
            "raw",
            secretBuffer,
            { name: "HMAC", hash: "SHA-256" },
            false,
            ["sign"],
        )

        const signatureBuffer = await crypto.subtle.sign("HMAC", key, dataBuffer)

        // Convert to base64url
        const signatureArray = new Uint8Array(signatureBuffer)
        return btoa(String.fromCharCode(...signatureArray))
            .replace(/\+/g, "-")
            .replace(/\//g, "_")
            .replace(/=/g, "")
    }
}

// HMAC auth instance using server password (for server endpoints)
export function createServerPasswordHMACAuth(password: string): HMACAuth {
    return new HMACAuth(password, 24 * 60 * 60)
}

// HMAC auth instance using Nakama password (for Nakama endpoints)
export function createNakamaHMACAuth(nakamaPassword: string): HMACAuth {
    return new HMACAuth(nakamaPassword, 24 * 60 * 60)
}
