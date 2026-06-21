import { createWriteStream, readFileSync, existsSync, mkdirSync, rmSync, readdirSync, cpSync } from "node:fs"
import { get } from "node:https"
import { createHash } from "node:crypto"
import { execSync } from "node:child_process"
import { homedir } from "node:os"
import { resolve, join, basename } from "node:path"

const supportedTargets = new Set([
    "darwin-arm64",
    "linux-x64",
    "linux-arm64",
    "win32-x64",
])

// Check target
const target = process.argv[2] || `${process.platform}-${process.arch}`
if (!supportedTargets.has(target)) {
    throw new Error(`Unsupported target: ${target}`)
}

const denshiRoot = resolve(import.meta.dirname, "..")
const projectRoot = resolve(denshiRoot, "..")
const lockfilePath = join(projectRoot, "mpv-prism.lock.json")

// Load lockfile
if (!existsSync(lockfilePath)) {
    throw new Error(`mpv-prism.lock.json not found at ${lockfilePath}`)
}
const lockfile = JSON.parse(readFileSync(lockfilePath, "utf-8"))

const packageJsonPath = join(denshiRoot, "package.json")
if (existsSync(packageJsonPath)) {
    const packageJson = JSON.parse(readFileSync(packageJsonPath, "utf-8"))
    const targetElectronVersion = packageJson.devDependencies?.electron || packageJson.dependencies?.electron
    const cleanVersion = (v) => v ? v.replace(/^[~^]/, "") : ""
    const appElectronVersion = cleanVersion(targetElectronVersion)
    const lockElectronVersion = cleanVersion(lockfile.electronVersion)

    if (appElectronVersion && lockElectronVersion) {
        const getMajor = (v) => v.split(".")[0]
        const appMajor = getMajor(appElectronVersion)
        const lockMajor = getMajor(lockElectronVersion)

        if (appMajor !== lockMajor) {
            throw new Error(
                `Electron version mismatch: seanime-denshi requires ${appElectronVersion}, ` +
                `but mpv-prism.lock.json is pinned to ${lockElectronVersion}. Native binaries are incompatible.`
            )
        } else if (appElectronVersion !== lockElectronVersion) {
            console.warn(
                `Warning: Electron mismatch (seanime-denshi: ${appElectronVersion}, ` +
                `lockfile: ${lockElectronVersion}). Proceeding as major version matches.`
            )
        }
    }
}

// Check for local override
const sourceRoot = process.env.MPV_PRISM_SOURCE ? resolve(process.env.MPV_PRISM_SOURCE) : null
const useLocal = sourceRoot ? existsSync(sourceRoot) : false

const stagingRoot = join(denshiRoot, "native-builds")
const stagedRuntime = join(stagingRoot, target)

// Staging preparation: Clean and recreate target dir
rmSync(stagingRoot, { recursive: true, force: true, maxRetries: 5, retryDelay: 100 })
mkdirSync(stagedRuntime, { recursive: true })

if (useLocal) {
    console.log(`Local mpv-prism build detected at ${sourceRoot}. Syncing from filesystem...`)

    // Verify local build artifacts
    for (const packageName of ["core", "electron", "react"]) {
        const distDirectory = join(sourceRoot, "packages", packageName, "dist")
        if (!existsSync(distDirectory)) {
            throw new Error(`mpv-prism package is not built: ${distDirectory}`)
        }
    }

    const sourceRuntime = join(sourceRoot, "native-builds", target)
    const addonPath = join(sourceRuntime, "mpv-prism.node")
    if (!existsSync(addonPath)) {
        throw new Error(`mpv-prism native addon is missing: ${addonPath}`)
    }

    const runtimeFiles = readdirSync(sourceRuntime)
    const hasLibMpv = target === "win32-x64"
        ? runtimeFiles.some(name => /^(?:lib)?mpv(?:-2)?\.dll$/i.test(name))
        : target === "darwin-arm64"
            ? runtimeFiles.some(name => /^libmpv(?:\.2)?\.dylib$/i.test(name))
            : runtimeFiles.some(name => /^libmpv\.so\.2(?:\..+)?$/i.test(name))

    if (!hasLibMpv) {
        throw new Error(`mpv-prism libmpv sidecar is missing from local ${sourceRuntime}`)
    }

    // Stage runtime
    cpSync(sourceRuntime, stagedRuntime, { recursive: true, dereference: true })
    console.log(`Staged mpv-prism ${target} from local ${basename(sourceRoot)} into ${stagedRuntime}`)
} else {
    console.log(`Local mpv-prism build not found. Resolving from lockfile...`)

    const nativeConfig = lockfile.native?.[target]
    if (!nativeConfig) {
        throw new Error(`Target ${target} is not defined in mpv-prism.lock.json`)
    }

    const { url, sha256: expectedSha } = nativeConfig
    const tempArchive = join(stagingRoot, `temp-${target}.tar.gz`)

    console.log(`Downloading native binaries for ${target} from ${url}...`)

    await downloadFile(url, tempArchive)

    console.log(`Verifying checksum...`)
    const fileBuffer = readFileSync(tempArchive)
    const actualSha = createHash("sha256").update(fileBuffer).digest("hex")

    if (actualSha !== expectedSha) {
        rmSync(tempArchive, { force: true })
        throw new Error(
            `Checksum mismatch for ${target}.\n` +
            `Expected: ${expectedSha}\n` +
            `Got:      ${actualSha}`
        )
    }

    console.log(`Extracting binaries to ${stagedRuntime}...`)
    execSync(`tar -xzf "${tempArchive}" -C "${stagedRuntime}"`, { stdio: "inherit" })
    rmSync(tempArchive, { force: true })

    // Verify extracted files
    const addonPath = join(stagedRuntime, "mpv-prism.node")
    if (!existsSync(addonPath)) {
        throw new Error(`Extraction failed: mpv-prism.node is missing from staging`)
    }

    const runtimeFiles = readdirSync(stagedRuntime)
    const hasLibMpv = target === "win32-x64"
        ? runtimeFiles.some(name => /^(?:lib)?mpv(?:-2)?\.dll$/i.test(name))
        : target === "darwin-arm64"
            ? runtimeFiles.some(name => /^libmpv(?:\.2)?\.dylib$/i.test(name))
            : runtimeFiles.some(name => /^libmpv\.so\.2(?:\..+)?$/i.test(name))

    if (!hasLibMpv) {
        throw new Error(`Extraction failed: libmpv sidecar is missing from staging`)
    }

    console.log(`Successfully staged and verified mpv-prism ${target} into ${stagedRuntime}`)
}

function downloadFile(url, destPath) {
    return new Promise((resolve, reject) => {
        get(url, (res) => {
            if (res.statusCode === 301 || res.statusCode === 302) {
                downloadFile(res.headers.location, destPath).then(resolve).catch(reject)
                return
            }
            if (res.statusCode !== 200) {
                reject(new Error(`Server responded with status ${res.statusCode}`))
                return
            }
            const fileStream = createWriteStream(destPath)
            res.pipe(fileStream)
            fileStream.on("finish", () => {
                fileStream.close()
                resolve()
            })
            fileStream.on("error", (err) => {
                reject(err)
            })
        }).on("error", (err) => {
            reject(err)
        })
    })
}
