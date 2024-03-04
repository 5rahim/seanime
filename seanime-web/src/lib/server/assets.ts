export function getAssetUrl(path: string) {
    let p = path.replaceAll("\\", "/")

    if (p.startsWith("/")) {
        p = p.substring(1)
    }

    return process.env.NODE_ENV === "development"
        ? `http://${window?.location?.hostname}:43211/assets/${p}`
        : `http://${window?.location?.host}/assets/${p}`
}
