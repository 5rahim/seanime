import { cpSync, createWriteStream, existsSync, mkdirSync, readdirSync, readFileSync, rmSync } from "node:fs"
import { get } from "node:https"
import { createHash } from "node:crypto"
import { execSync } from "node:child_process"
import { basename, join, resolve } from "node:path"

const supportedTargets = new Set([
    "darwin-arm64",
    "linux-x64",
    "linux-arm64",
    "win32-x64",
])

// Check target
const target = process.argv[2] || "all"
if (target !== "all" && !supportedTargets.has(target)) {
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

// Staging preparation: Clean and recreate stagingRoot
rmSync(stagingRoot, { recursive: true, force: true, maxRetries: 5, retryDelay: 100 })
mkdirSync(stagingRoot, { recursive: true })

const targetsToSync = target === "all" ? Array.from(supportedTargets) : [target]

if (useLocal) {
    console.log(`Local mpv-prism build detected at ${sourceRoot}. Syncing from filesystem...`)

    // Verify local build artifacts
    for (const packageName of ["core", "electron", "react"]) {
        const distDirectory = join(sourceRoot, "packages", packageName, "dist")
        if (!existsSync(distDirectory)) {
            throw new Error(`mpv-prism package is not built: ${distDirectory}`)
        }
    }

    for (const t of targetsToSync) {
        const sourceRuntime = join(sourceRoot, "native-builds", t)
        if (!existsSync(sourceRuntime)) {
            if (target === "all") {
                console.log(`Skipping local native build for ${t} (not found at ${sourceRuntime})`)
                continue
            }
            throw new Error(`mpv-prism native build directory is missing: ${sourceRuntime}`)
        }

        const addonPath = join(sourceRuntime, "mpv-prism.node")
        if (!existsSync(addonPath)) {
            throw new Error(`mpv-prism native addon is missing: ${addonPath}`)
        }

        const runtimeFiles = readdirSync(sourceRuntime)
        const hasLibMpv = t === "win32-x64"
            ? runtimeFiles.some(name => /^(?:lib)?mpv(?:-2)?\.dll$/i.test(name))
            : t === "darwin-arm64"
                ? runtimeFiles.some(name => /^libmpv(?:\.2)?\.dylib$/i.test(name))
                : runtimeFiles.some(name => /^libmpv\.so\.2(?:\..+)?$/i.test(name))

        if (!hasLibMpv) {
            throw new Error(`mpv-prism libmpv sidecar is missing from local ${sourceRuntime}`)
        }

        const stagedRuntime = join(stagingRoot, t)
        mkdirSync(stagedRuntime, { recursive: true })
        cpSync(sourceRuntime, stagedRuntime, { recursive: true, dereference: true })
        console.log(`Staged mpv-prism ${t} from local ${basename(sourceRoot)} into ${stagedRuntime}`)
    }

    // Stage packages into node_modules of seanime-denshi and seanime-web to override downloaded ones
    const stagePackage = (packageName, destNodeModules) => {
        const pkgDist = join(sourceRoot, "packages", packageName, "dist")
        const targetDir = join(destNodeModules, "@mpv-prism", packageName)
        if (existsSync(targetDir)) {
            const destDist = join(targetDir, "dist")
            rmSync(destDist, { recursive: true, force: true })
            cpSync(pkgDist, destDist, { recursive: true })
            const pkgJSON = join(sourceRoot, "packages", packageName, "package.json")
            if (existsSync(pkgJSON)) {
                cpSync(pkgJSON, join(targetDir, "package.json"))
            }
            console.log(`Staged local @mpv-prism/${packageName} into ${targetDir}`)
        }
    }

    const denshiNodeModules = join(denshiRoot, "node_modules")
    const webNodeModules = join(projectRoot, "seanime-web", "node_modules")

    for (const pkg of ["core", "electron", "react"]) {
        stagePackage(pkg, denshiNodeModules)
        stagePackage(pkg, webNodeModules)
    }
} else {
    console.log(`Local mpv-prism build not found. Resolving from lockfile...`)

    for (const t of targetsToSync) {
        const nativeConfig = lockfile.native?.[t]
        if (!nativeConfig) {
            throw new Error(`Target ${t} is not defined in mpv-prism.lock.json`)
        }

        const stagedRuntime = join(stagingRoot, t)
        mkdirSync(stagedRuntime, { recursive: true })

        const { url, sha256: expectedSha } = nativeConfig
        const tempArchive = join(stagingRoot, `temp-${t}.tar.gz`)

        console.log(`Downloading native binaries for ${t} from ${url}...`)

        await downloadFile(url, tempArchive)

        console.log(`Verifying checksum...`)
        const fileBuffer = readFileSync(tempArchive)
        const actualSha = createHash("sha256").update(fileBuffer).digest("hex")

        if (actualSha !== expectedSha) {
            rmSync(tempArchive, { force: true })
            throw new Error(
                `Checksum mismatch for ${t}.\n` +
                `Expected: ${expectedSha}\n` +
                `Got:      ${actualSha}`
            )
        }

        console.log(`Extracting binaries to ${stagedRuntime}...`)
        execSync(`tar -xzf "temp-${t}.tar.gz" -C "${t}"`, { stdio: "inherit", cwd: stagingRoot })
        rmSync(tempArchive, { force: true })

        // Verify extracted files
        const addonPath = join(stagedRuntime, "mpv-prism.node")
        if (!existsSync(addonPath)) {
            throw new Error(`Extraction failed: mpv-prism.node is missing from staging`)
        }

        const runtimeFiles = readdirSync(stagedRuntime)
        const hasLibMpv = t === "win32-x64"
            ? runtimeFiles.some(name => /^(?:lib)?mpv(?:-2)?\.dll$/i.test(name))
            : t === "darwin-arm64"
                ? runtimeFiles.some(name => /^libmpv(?:\.2)?\.dylib$/i.test(name))
                : runtimeFiles.some(name => /^libmpv\.so\.2(?:\..+)?$/i.test(name))

        if (!hasLibMpv) {
            throw new Error(`Extraction failed: libmpv sidecar is missing from staging`)
        }

        console.log(`Successfully staged and verified mpv-prism ${t} into ${stagedRuntime}`)
    }
}

function downloadFile(url, destPath) {
    return new Promise((resolve, reject) => {
        const urlObj = new URL(url)
        urlObj.searchParams.set("cb", Date.now().toString())

        const options = {
            headers: {
                "Cache-Control": "no-cache",
                "Pragma": "no-cache"
            }
        }

        get(urlObj, options, (res) => {
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
