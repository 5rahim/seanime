export function shouldRecoverStartup(currentTime: number) {
    return currentTime < 1
}

export function getRefreshKey(provider: string | null, server: string | undefined, episodeNumber: number | null) {
    return `${provider ?? ""}:${server ?? ""}:${episodeNumber ?? ""}`
}

export function markSourceRefreshed(refreshed: Set<string>, key: string) {
    if (refreshed.has(key)) return false
    refreshed.add(key)
    return true
}

export function orderProviders(providers: Array<{ id: string }>, currentProvider: string | null) {
    const ids = providers.map(provider => provider.id)
    if (!currentProvider || !ids.includes(currentProvider)) return ids
    return [currentProvider, ...ids.filter(id => id !== currentProvider)]
}
