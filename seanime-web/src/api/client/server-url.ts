import { __DEV_SERVER_PORT, TESTONLY__DEV_SERVER_PORT_NAKAMA } from "@/lib/server/config"
import { __isDesktop__ } from "@/types/constants"

function devOrProd(dev: string, prod: string): string {
    return process.env.NODE_ENV === "development" ? dev : prod
}

export function getServerBaseUrl(removeProtocol: boolean = false): string {
    if (__isDesktop__) {
        let ret = devOrProd(`http://127.0.0.1:${__DEV_SERVER_PORT}`, "http://127.0.0.1:43211")
        if (removeProtocol) {
            ret = ret.replace("http://", "").replace("https://", "")
        }
        return ret
    }

    // DEV ONLY: Hack to allow 2 development servers to run with the same development web server
    // - og web server: 127.0.0.1:43210 -> 127.0.0.1:43000
    // - nakama web server: localhost:43210 -> 127.0.0.1:43001
    if (process.env.NODE_ENV === "development" && window.location.host.includes("localhost")) {
        let ret = `http://127.0.0.1:${TESTONLY__DEV_SERVER_PORT_NAKAMA}`
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
