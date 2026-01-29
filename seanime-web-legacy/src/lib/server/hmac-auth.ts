import * as CryptoJS from "crypto-js"

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
        const signature = CryptoJS.HmacSHA256(data, this.secret)

        const base64 = CryptoJS.enc.Base64.stringify(signature)
        return base64
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
