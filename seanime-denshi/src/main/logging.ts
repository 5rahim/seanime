import log from "electron-log/main"
import { autoUpdater } from "electron-updater"

const ANSI_REGEX =
    /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g

let loggingInitialized = false

function stripAnsi(value: string): string {
    return value.replace(ANSI_REGEX, "")
}

function stripAnsiFromData(data: unknown): unknown {
    if (Array.isArray(data)) {
        return data.map(item => (
            typeof item === "string" ? stripAnsi(item) : item
        ))
    }
    if (typeof data === "string") {
        return stripAnsi(data)
    }

    return data
}

export function setupLogging(): void {
    if (loggingInitialized) {
        return
    }

    loggingInitialized = true

    log.initialize()
    log.transports.file.level = "debug"

    log.transports.file.transforms.push(({ data }) => {
        return stripAnsiFromData(data)
    })

    autoUpdater.logger = log

    // electron-log's supported console redirection mechanism
    Object.assign(console, log.functions)

    process.on("warning", warning => {
        if (
            warning.name === "DeprecationWarning"
            && warning.message.includes("fs.Stats")
        ) {
            return
        }

        log.warn(`${warning.name}: ${warning.message}`, warning.stack ?? "")
    })
}

export { log }
