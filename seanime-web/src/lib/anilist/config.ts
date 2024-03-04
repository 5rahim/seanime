export const ANILIST_OAUTH_URL = `https://anilist.co/api/v2/oauth/authorize?client_id=17297&response_type=token`
export const ANILIST_PIN_URL = `https://anilist.co/api/v2/oauth/authorize?client_id=13985&response_type=token`

export const MAL_CLIENT_ID = `51cb4294feb400f3ddc66a30f9b9a00f`

export const ANILIST_API_ENDPOINT = `https://graphql.anilist.co`

export const ANILIST_BOTTLENECK_OPTIONS = {
    reservoir: 90, // initial value
    reservoirRefreshAmount: 90,
    reservoirRefreshInterval: 60 * 1000, // must be divisible by 250
    maxConcurrent: 1,
    minTime: 1000 / 90, // Minimum time (in milliseconds) between requests - 90 requests per minute
}
