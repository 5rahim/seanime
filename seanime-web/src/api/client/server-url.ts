import { __DEV_SERVER_PORT } from "@/lib/server/config"

function devOrProd(dev: string, prod: string): string {
    return process.env.NODE_ENV === "development" ? dev : prod
}

export function getServerBaseUrl(removeProtocol: boolean = false): string {
    if (process.env.NEXT_PUBLIC_PLATFORM === "desktop") {
        let ret = devOrProd(`http://127.0.0.1:${__DEV_SERVER_PORT}`, "http://127.0.0.1:43211")
        if (removeProtocol) {
            ret = ret.replace("http://", "").replace("https://", "")
        }
        return ret
    }

    let ret = typeof window !== "undefined"
        ? (`${window?.location?.protocol}//` + devOrProd(`${window?.location?.hostname}:${__DEV_SERVER_PORT}`, window?.location?.host))
        : ""
    if (removeProtocol) {
        ret = ret.replace("http://", "").replace("https://", "")
    }
    return ret
}
