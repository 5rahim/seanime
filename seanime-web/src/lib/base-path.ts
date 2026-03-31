function normalizeBasePath(value?: string): string {
    const raw = (value || "").trim()
    if (!raw || raw === "/") return "/"

    let path = raw

    if (path.includes("://")) {
        try {
            path = new URL(path).pathname || "/"
        } catch {
            path = "/"
        }
    }

    if (!path.startsWith("/")) {
        path = `/${path}`
    }

    path = `/${path.replace(/^\/+/, "").replace(/\/+$/, "")}`
    return path === "/" ? "/" : path
}

const runtimeBasePath = typeof window !== "undefined"
    ? ((window as any).__SEANIME_BASE_URL__ as string | undefined)
    : undefined

export const APP_BASE_PATH = normalizeBasePath(runtimeBasePath || import.meta.env.SEA_PUBLIC_BASE_URL)

export function withBasePath(path: string): string {
    if (!path) return APP_BASE_PATH
    if (/^[a-zA-Z][a-zA-Z\d+\-.]*:/.test(path) || path.startsWith("//")) return path

    const normalizedPath = path.startsWith("/") ? path : `/${path}`
    if (APP_BASE_PATH === "/") return normalizedPath
    return `${APP_BASE_PATH}${normalizedPath}`
}
