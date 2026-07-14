export function getProxyUrl(baseUrl: string, url: string, headers: Record<string, string>, tokenQuery: string) {
    return `${baseUrl}/api/v1/proxy?url=${encodeURIComponent(url)}&headers=${encodeURIComponent(JSON.stringify(headers))}${tokenQuery}`
}
