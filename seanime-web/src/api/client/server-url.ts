import { __DEV_SERVER_PORT } from "@/lib/server/config"
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

    // // DEV ONLY: Hack to allow multiple development servers for the same web server
    // // localhost:43210 -> 127.0.0.1:43001
    // // 192.168.1.100:43210 -> 127.0.0.1:43002
    // if (process.env.NODE_ENV === "development" && window.location.host.includes("localhost")) {
    //     let ret = `http://127.0.0.1:${TESTONLY__DEV_SERVER_PORT2}`
    //     if (removeProtocol) {
    //         ret = ret.replace("http://", "").replace("https://", "")
    //     }
    //     return ret
    // }
    // if (process.env.NODE_ENV === "development" && window.location.host.startsWith("192.168")) {
    //     let ret = `http://127.0.0.1:${TESTONLY__DEV_SERVER_PORT3}`
    //     if (removeProtocol) {
    //         ret = ret.replace("http://", "").replace("https://", "")
    //     }
    //     return ret
    // }

    let ret = typeof window !== "undefined"
        ? (`${window?.location?.protocol}//` + devOrProd(`${window?.location?.hostname}:${__DEV_SERVER_PORT}`, window?.location?.host))
        : ""
    if (removeProtocol) {
        ret = ret.replace("http://", "").replace("https://", "")
    }
    return ret
}
