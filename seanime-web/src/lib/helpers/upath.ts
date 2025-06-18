/**
 * Represents a path object returned by `parse`.
 */
interface ParsedPath {
    root: string;
    dir: string;
    base: string;
    ext: string;
    name: string;
}

/**
 * Represents an object used by `format`.
 */
interface FormatInputPathObject {
    root?: string;
    dir?: string;
    base?: string;
    ext?: string;
    name?: string;
}

// Define the shape of the exported module
// This interface reflects the public API, including reimplemented core functions
// and added extra functions.
interface UPath {
    VERSION: string;
    sep: string;
    delimiter: string; // Common path property, though less critical in browser/unix

    // Reimplemented core path functions - signatures match Node.js path (posix)
    // They internally handle string arguments/results with toUnix where appropriate
    join(...paths: string[]): string;

    resolve(...pathSegments: string[]): string;

    normalize(p: string): string;

    isAbsolute(p: string): boolean;

    dirname(p: string): string;

    basename(p: string, ext?: string): string;

    extname(p: string): string;

    format(p: FormatInputPathObject): string;

    parse(p: string): ParsedPath;

    relative(from: string, to: string): string;

    // Extra functions
    toUnix(p: string): string;

    normalizeSafe(p: string): string;

    normalizeTrim(p: string): string;

    joinSafe(...p: string[]): string;

    addExt(file: string, ext?: string): string;

    trimExt(filename: string, ignoreExts?: string[], maxSize?: number): string;

    removeExt(filename: string, ext?: string): string;

    changeExt(filename: string, ext?: string, ignoreExts?: string[], maxSize?: number): string;

    defaultExt(filename: string, ext?: string, ignoreExts?: string[], maxSize?: number): string;
}

// --- Helper Functions ---

const isString = (val: any): val is string =>
    typeof val === "string" ||
    (!!val && typeof val === "object" && Object.prototype.toString.call(val) === "[object String]")

/**
 * Converts a path to a Unix-style path with forward slashes.
 * Replaces backslashes with forward slashes.
 * Collapses multiple consecutive forward slashes into a single slash,
 * except for the leading slashes in a UNC path (e.g., //server/share).
 * This replicates the behavior of the original /(?<!^)\/+/g regex without lookbehind
 * by using a different regex or checking the match offset.
 * @param p The path string.
 * @returns The Unix-style path string.
 */
const toUnix = (p: string): string => {
    if (!isString(p)) {
        return p as any // Coerce non-strings to string-like if necessary, matching original flexibility
    }
    let unixPath = p.replace(/\\/g, "/")

    // Replace sequences of 2+ forward slashes with a single slash, but only if not at the beginning
    // Replicates /(?<!^)\/+/g without lookbehind by checking the match offset.
    // Regex /\/{2,}/g matches two or more slashes.
    // The check `offset > 0` ensures we don't collapse leading `//` or `///`.
    // Example: '/a//b' -> offset=2, matches `//`, replace with `/` -> '/a/b'.
    // Example: '//a//b' -> offset=0, matches `//`, keep `//`. offset=4, matches `//`, replace with `/` -> '//a/b'.
    // Example: '///a' -> offset=0, matches `///`, keep `///`. -> '///a'. (Matches original regex behavior)
    unixPath = unixPath.replace(/\/{2,}/g, (match, offset) => offset > 0 ? "/" : match)

    return unixPath
}

/**
 * Cleans up path segments, resolving '.' and '..'.
 * @param segments Array of path segments (already split and filtered).
 * @param isAbsolute Boolean indicating if the original path was absolute.
 * @returns Array of cleaned segments.
 */
const _processSegments = (segments: string[], isAbsolute: boolean): string[] => {
    const res: string[] = []

    for (const segment of segments) {
        if (segment === "." || segment === "") {
            continue // Ignore '.' and empty segments
        }
        if (segment === "..") {
            // Pop the last segment unless we are at the effective root
            // Effective root for absolute paths is always '/', for relative paths it's the starting point.
            // We can pop if there are segments in res AND the last segment isn't '..'.
            // For absolute paths, we cannot pop if res is empty (already at root).
            if (res.length > 0 && res[res.length - 1] !== "..") {
                res.pop()
            } else if (!isAbsolute) {
                // If not absolute, allow '..' to go above the starting point
                res.push("..")
            }
            // If isAbsolute is true and res is empty or ends in '..', we cannot go up, so just ignore the '..'.
        } else {
            res.push(segment)
        }
    }

    return res
}


/**
 * Checks if a given extension is valid based on ignore list and max size.
 * @param ext The extension string (e.g., '.js').
 * @param ignoreExts Array of extensions to ignore (e.g., ['js', '.txt']).
 * @param maxSize Maximum allowed extension length.
 * @returns True if the extension is valid, false otherwise.
 */
const isValidExt = (ext: string | undefined, ignoreExts: string[] = [], maxSize = 7): boolean => {
    if (!ext) {
        return false
    }

    // Normalize ignoreExts to always start with '.' for comparison
    const normalizedIgnoreExts = ignoreExts
        .filter(e => !!e) // Filter out null/undefined/empty strings
        .map(e => (e[0] !== "." ? "." : "") + e)

    return (ext.length <= maxSize) && !normalizedIgnoreExts.includes(ext)
}


// --- Core Path Function Implementations (Browser/Unix style) ---
// These replace the dependency on Node.js 'path'.

const _isAbsolute = (p: string): boolean => {
    if (!isString(p)) return false // Or throw? Node.js path expects string.
    p = toUnix(p) // Ensure forward slashes for check
    // Check for Unix root '/', UNC path start `//`, or Windows drive letter `C:/`
    return p.length > 0 &&
        (p[0] === "/" ||
            (p.length > 1 && p[0] === "/" && p[1] === "/") ||
            /^[a-zA-Z]:\//.test(p)) // Windows drive letter check
}


const _normalize = (p: string): string => {
    if (!isString(p)) {
        return p as any
    }
    p = toUnix(p)

    if (p.length === 0) {
        return "."
    }

    const isAbsolute = p.startsWith("/")
    const isUNC = p.startsWith("//")
    // Keep track of trailing slash for non-root paths
    const hasTrailingSlash = p.endsWith("/") && p.length > 1

    // Split path into segments, filtering out empty ones.
    // This gives us ['a', 'b'] for '/a//b/'.
    const segments = p.split("/").filter(segment => segment !== "")

    // Process segments (resolve '.', '..')
    const processedSegments = _processSegments(segments, isAbsolute)

    // Reconstruct the path
    let result = processedSegments.join("/")

    // If originally absolute, prepend root
    if (isAbsolute) {
        if (isUNC) {
            result = "//" + result
        } else {
            result = "/" + result
        }
    }

    // Handle edge case: absolute path normalized to empty segments (e.g. '/../')
    // Should result in '/' or '//' for UNC paths
    if (isAbsolute && result.length === 0) {
        return isUNC ? "//" : "/"
    }

    // Add trailing slash back based on original path AND if the result wasn't normalized to '.'
    // Node.js posix trailing slash rule: preserve if original path had it AND result is not '.' AND result does not end with '/..'
    // A simplified rule: preserve if original had it AND the result is not just '.' and not just '/' and not just '//'
    if (hasTrailingSlash && result !== "." && result !== "/" && result !== "//") {
        result += "/"
    }

    // If the result is empty and it's not absolute, return '.'
    if (result.length === 0 && !isAbsolute) {
        return "."
    }

    return result
}


// Node.js posix.join is essentially normalize(path1 + '/' + path2 + '/' + ...)
const _join = (...paths: string[]): string => {
    // Convert all inputs to Unix style and filter out empty/invalid ones
    const unixPaths = paths.map(p => isString(p) ? toUnix(p) : p).filter(p => isString(p) && p.length > 0)

    if (unixPaths.length === 0) {
        return "." // Matches Node.js path.join('') -> '.'
    }

    // Join all valid non-empty strings with a single slash
    const joined = unixPaths.join("/")

    // Then normalize the result
    return _normalize(joined)
}


const _dirname = (p: string): string => {
    if (!isString(p)) return "." // Or throw? Node.js path expects string.
    p = toUnix(p)

    if (p.length === 0) {
        return "."
    }

    // Remove trailing slashes unless it's the root '/'
    let end = p.length
    while (end > 1 && p[end - 1] === "/") {
        end--
    }
    p = p.substring(0, end)

    // Find the last slash
    const lastSlash = p.lastIndexOf("/")

    if (lastSlash < 0) {
        // No slash found, directory is '.'
        return "."
    }

    // Find the first non-slash character from the start (to detect root part like '/', '//')
    let rootEnd = 0
    while (rootEnd < p.length && p[rootEnd] === "/") {
        rootEnd++
    }

    // If the last slash is within or at the end of the root sequence, the dirname is the root.
    // e.g., '/a' (lastSlash=0, rootEnd=1), '//a' (lastSlash=1, rootEnd=2)
    if (lastSlash < rootEnd) {
        return p.substring(0, rootEnd) // Return the root string (e.g., '/', '//')
    }

    // Otherwise, the dirname is the part before the last slash
    return p.substring(0, lastSlash)
}

const _basename = (p: string, ext?: string): string => {
    if (!isString(p)) return "" // Or throw? Node.js path expects string.
    p = toUnix(p)

    if (p.length === 0) {
        return ""
    }

    // Remove trailing slashes
    let end = p.length
    while (end > 0 && p[end - 1] === "/") {
        end--
    }
    p = p.substring(0, end)

    // Find the last slash to get the base name
    const lastSlash = p.lastIndexOf("/")

    let base = (lastSlash === -1) ? p : p.substring(lastSlash + 1)

    // If extension is provided and matches the end of the base name, remove it
    // Node.js rule: only remove ext if base is longer than ext and the character before ext is NOT '.'
    // A simpler rule (closer to Node.js): remove if `base` ends with `ext` AND base is not just `.`
    if (ext && base.endsWith(ext) && base !== ".") {
        // Ensure removing ext doesn't result in '.' from something like '.c' with ext '.c'
        const baseWithoutExt = base.slice(0, -ext.length)
        if (baseWithoutExt.length > 0 || ext === ".") { // Handles basename('.bashrc', '.rc') -> '.bashrc'
            base = baseWithoutExt
        } else if (ext.length === base.length) { // Handle basename('.c', '.c') -> ''
            base = ""
        }

    }

    return base
}

const _extname = (p: string): string => {
    if (!isString(p)) return "" // Or throw? Node.js path expects string.
    p = toUnix(p)

    // Special case: empty string
    if (p.length === 0) {
        return ""
    }

    // Remove trailing slashes
    let end = p.length
    while (end > 0 && p[end - 1] === "/") {
        end--
    }
    p = p.substring(0, end)

    // Find the last slash (directory separator)
    const lastSlash = p.lastIndexOf("/")
    // Find the last dot
    const lastDot = p.lastIndexOf(".")

    // No dot or dot is before the last slash -> no extension
    if (lastDot < 0 || lastDot < lastSlash) {
        return ""
    }

    // Handle dotfiles: if the dot is the first character of the basename
    // e.g., '.bashrc', '.gitignore'. extname should be ''.
    // The base name starts after the last slash.
    const baseNameStart = lastSlash + 1
    // If lastDot is the first character of the basename AND the base name is not just '.' or '..'
    if (lastDot === baseNameStart && p.substring(baseNameStart) !== "." && p.substring(baseNameStart) !== "..") {
        return ""
    }


    // Extension is from the last dot to the end
    return p.substring(lastDot)
}

const _resolve = (...pathSegments: string[]): string => {
    let resolvedPath = ""
    let resolvedAbsolute = false

    // Iterate paths from right to left to find the last absolute path
    for (let i = pathSegments.length - 1; i >= 0; i--) {
        const path = toUnix(pathSegments[i]) // Ensure unix path

        if (path.length === 0) {
            continue // Ignore empty segments
        }

        resolvedPath = path + "/" + resolvedPath // Prepend the current segment
        if (_isAbsolute(path)) {
            resolvedAbsolute = true
            break // Found the rightmost absolute path, stop combining further left
        }
    }

    // Remove trailing slash from combined path before normalization
    // Important: Check resolvedPath length before slicing
    if (resolvedPath.length > 1 && resolvedPath.endsWith("/")) {
        resolvedPath = resolvedPath.slice(0, -1)
    } else if (resolvedPath === "/") {
        // Keep single slash if that's the whole path after loop
    } else if (resolvedPath.endsWith("/") && resolvedPath.length > 0) {
        // Remove trailing slash if path is not just '/'
        resolvedPath = resolvedPath.slice(0, -1)
    }


    // If no absolute path was found, prepend the root '/' (simulating cwd)
    if (!resolvedAbsolute) {
        // Simulating resolving relative to root '/'
        resolvedPath = "/" + resolvedPath
        // Remove trailing slash again if added root results in e.g. "/path/"
        if (resolvedPath.length > 1 && resolvedPath.endsWith("/")) {
            resolvedPath = resolvedPath.slice(0, -1)
        }
    }


    // Normalize the final path
    const normalized = _normalize(resolvedPath)

    // If the result of normalizing an absolute path is '.', it should be '/'
    // Also handle if normalization of an absolute path resulted in empty string (e.g. resolve('/','..'))
    if ((normalized === "." || normalized === "") && resolvedAbsolute) {
        return "/"
    }

    // Ensure result starts with '/' if it was resolved as absolute
    // _normalize should handle this, but as a safeguard
    if (resolvedAbsolute && normalized.length > 0 && !normalized.startsWith("/")) {
        return "/" + normalized
    }


    return normalized
}

const _relative = (from: string, to: string): string => {
    if (!isString(from) || !isString(to)) {
        return "" // Return empty string for invalid inputs
    }

    // Resolve both paths first
    from = _resolve(from)
    to = _resolve(to)

    if (from === to) {
        return ""
    }

    // Split resolved paths into segments, filtering out the initial root '/'
    const fromParts = from.split("/").filter(p => p !== "")
    const toParts = to.split("/").filter(p => p !== "")

    // Find the common prefix length
    let commonLength = 0
    const maxLength = Math.min(fromParts.length, toParts.length)
    while (commonLength < maxLength && fromParts[commonLength] === toParts[commonLength]) {
        commonLength++
    }

    // Calculate 'up' moves needed from 'from' to the common prefix
    const upMoves = fromParts.length - commonLength
    const relativeParts: string[] = []

    // Add '..' for each segment we need to go up
    for (let i = 0; i < upMoves; i++) {
        relativeParts.push("..")
    }

    // Add segments from 'to' after the common prefix
    for (let i = commonLength; i < toParts.length; i++) {
        relativeParts.push(toParts[i])
    }

    // If the result is empty (e.g., relative('/a/b', '/a/b/c') when 'b' is common),
    // and 'to' is not '/', it means the target was a child, return '.' if no parts added.
    // Node.js relative('/a', '/a') -> '', relative('/a/b', '/a') -> '..', relative('/a', '/a/b') -> 'b'
    // The logic above already handles these. If relativeParts is empty, from === to.

    if (relativeParts.length === 0) {
        // This should only happen if from === to, which is handled at the beginning.
        // As a fallback, return '.' if they weren't strictly equal but resulted in no relative path (e.g., different trailing slashes resolving the
        // same)
        return "." // This is different from Node.js relative which returns '' for identical resolved paths.
        // Let's stick to the Node.js behavior and rely on from === to check.
        // If from === to, return ''. Otherwise, relativeParts should not be empty unless one is ancestor of other resolving to empty relative part.
        // E.g. relative('/a', '/a') -> '', relative('/a/', '/a') -> ''
        // Let's remove the '.' fallback and trust the resolved paths logic.
        return "" // Should not be reached if from !== to and relativeParts is empty.
    }


    return relativeParts.join("/")
}


// Helper for parse to get root, dir, base, ext, name
const _parse = (p: string): ParsedPath => {
    if (!isString(p)) {
        return { root: "", dir: "", base: "", ext: "", name: "" }
    }

    p = toUnix(p)

    const result: ParsedPath = { root: "", dir: "", base: "", ext: "", name: "" }

    if (p.length === 0) {
        return result
    }

    let rest = p

    // 1. Root
    if (rest.startsWith("//")) {
        // UNC-like path start
        // Node.js posix.parse('//server/share/dir/base') -> { root: '//', dir: '//server/share/dir', base: 'base', ... }
        // The root is just '//' in posix.
        result.root = "//"
        rest = rest.substring(2) // Remove leading '//'
    } else if (rest.startsWith("/")) {
        // Regular absolute path
        result.root = "/"
        rest = rest.substring(1) // Remove leading '/'
    }

    // Remove trailing slashes from the rest for consistent parsing of base/ext
    let end = rest.length
    while (end > 0 && rest[end - 1] === "/") {
        end--
    }
    const cleanedRest = rest.substring(0, end)


    // 2. Base and Ext
    const lastSlash = cleanedRest.lastIndexOf("/")
    result.base = (lastSlash === -1) ? cleanedRest : cleanedRest.substring(lastSlash + 1)

    // Determine extname and name from base (using internal function)
    const baseLastDot = result.base.lastIndexOf(".")
    if (baseLastDot < 0 || baseLastDot === 0 || (baseLastDot === result.base.length - 1 && result.base.length > 1)) { // No dot, dot is first char, or dot is last char (e.g., 'file.')
        result.ext = ""
        result.name = result.base
    } else {
        result.ext = result.base.substring(baseLastDot)
        result.name = result.base.substring(0, baseLastDot)
    }

    // 3. Dir
    if (lastSlash === -1) {
        // No slash in the rest means the base is the only part after the root.
        // The directory is just the root.
        result.dir = result.root
        // Node behavior: If root was empty, dir remains empty.
    } else {
        // The directory is the part of cleanedRest before the last slash, combined with the root.
        result.dir = result.root + cleanedRest.substring(0, lastSlash)
    }

    return result
}

const _format = (pathObject: FormatInputPathObject): string => {
    // Node.js path.format rules (posix):
    // 1. If pathObject.base is provided, it takes precedence over pathObject.ext and pathObject.name.
    // 2. If pathObject.base is not provided, pathObject.name and pathObject.ext are used.
    // 3. pathObject.dir takes precedence over pathObject.root.
    // 4. If pathObject.dir is provided:
    //    Result is dir + (dir ends with / ? '' : '/') + base.
    // 5. If pathObject.dir is not provided, but pathObject.root is:
    //    Result is root + base.
    // 6. If none of dir, root, or base are provided, result is '.'

    const root = pathObject.root || ""
    const dir = pathObject.dir || ""
    const base = pathObject.base ?? ((pathObject.name ?? "") + (pathObject.ext ?? ""))

    if (dir) {
        // If dir is provided, root is effectively ignored for the prefix logic.
        // Ensure dir is Unix style.
        let unixDir = toUnix(dir)

        // Node behavior: If dir ends with a slash, keep it, otherwise don't add one.

        // Join dir and base, adding a slash ONLY if dir doesn't already end with one AND base exists.
        if (base) {
            if (unixDir.endsWith("/")) {
                return unixDir + base
            } else {
                return unixDir + "/" + base
            }
        } else {
            // If no base, result is just the dir.
            return unixDir || "." // Handle empty dir becoming '.'
        }

    } else if (root) {
        // If dir is not provided, use root + base.
        // Root should already be normalized ('/' or '//').
        // If base is empty, result is just root.
        // If base is not empty, append it. Ensure base doesn't start with / if root is present.
        let cleanedBase = base
        while (cleanedBase.startsWith("/")) cleanedBase = cleanedBase.slice(1)

        return root + cleanedBase // Example: format({ root: '/', base: 'a' }) -> '/a'

    } else if (base) {
        // Neither dir nor root provided, result is just base.
        return base
    } else {
        // Nothing provided, return '.'
        return "."
    }

}


// --- Initialize UPath Object ---

// Create the internal object that will be the public API
const upath_internal: any = {
    // Define VERSION (assuming VERSION is a global or module-scoped variable injected elsewhere)
    // If VERSION is not injected, this will default to 'NO-VERSION'
    VERSION: typeof (globalThis as any).VERSION !== "undefined" ? (globalThis as any).VERSION : "NO-VERSION",
    sep: "/", // Explicitly set to Unix style
    delimiter: ":", // Standard Posix delimiter
}

// Assign the implemented core functions
upath_internal.join = _join
upath_internal.resolve = _resolve
upath_internal.normalize = _normalize
upath_internal.isAbsolute = _isAbsolute
upath_internal.dirname = _dirname
upath_internal.basename = _basename
upath_internal.extname = _extname
upath_internal.format = _format
upath_internal.parse = _parse
upath_internal.relative = _relative


// --- Extra Functions ---

const extraFunctions = {

    // Include the internal helper for external use
    toUnix: toUnix,

    /**
     * Normalizes a path safely using upath.normalize.
     * Attempts to restore './' or '//' prefixes if lost during normalization.
     * Note: The logic for restoring '//' might be specific or potentially buggy
     * depending on the exact scenarios it was designed for, but it matches
     * the original source code's behavior.
     * @param p The path string.
     * @returns The normalized path string with prefixes potentially restored.
     */
    normalizeSafe: (p: string): string => {
        // Ensure input is string for toUnix, though the interface suggests string input
        const originalP = isString(p) ? toUnix(p) : p
        if (!isString(originalP)) {
            return p // Cannot normalize non-string
        }

        // Use the implemented normalize function
        const result = upath_internal.normalize(originalP)

        // Restore './' prefix if original started with it, result doesn't, and isn't '..' or '/'
        // Check !upath_internal.isAbsolute(result) added to avoid './' for absolute paths.
        if (originalP.startsWith("./") && !result.startsWith("./") && !upath_internal.isAbsolute(result) && result !== "..") {
            return "./" + result
        }
        // Special case: Handle "//./..." paths separately to preserve the "//." prefix
        else if (originalP.startsWith("//./")) {
            // Remove the leading "//" from result before prepending "//."
            const resultWithoutLeadingSlashes = result.startsWith("//") ? result.substring(2) : result.startsWith("/") ? result.substring(1) : result
            return "//." + (resultWithoutLeadingSlashes ? "/" + resultWithoutLeadingSlashes : "")
        }
        // Restore '//' prefix if original started with it and result lost it or changed it
        else if (originalP.startsWith("//") && !result.startsWith("//")) {
            return "/" + result
        }
        // Otherwise, return the result from upath.normalize
        return result
    },

    /**
     * Normalizes a path safely using normalizeSafe, then removes a trailing slash
     * unless the path is the root ('/').
     * @param p The path string.
     * @returns The normalized path string without a trailing slash (unless root).
     */
    normalizeTrim: (p: string): string => {
        p = upath_internal.normalizeSafe(p)
        // Remove trailing slash, unless it's the root '/'
        if (p.endsWith("/") && p.length > 1) {
            return p.slice(0, p.length - 1)
        } else {
            return p
        }
    },

    /**
     * Joins path segments safely. Calls upath.join, then attempts to restore
     * './' or '//' prefix based on the first original path segment if it was lost.
     * Note: The logic for restoring '//' might be specific or potentially buggy
     * depending on the exact scenarios it was designed for, but it matches
     * the original source code's behavior.
     * @param p Path segments to join. Accepts string[] as per join signature.
     * @returns The joined path string with prefixes potentially restored.
     */
    joinSafe: (...p: string[]): string => {
        // Get the original first argument (before toUnix conversion in join)
        const p0Original = p.length > 0 ? p[0] : undefined

        // Use the implemented join function
        const result = upath_internal.join.apply(null, p)

        // Apply prefix restore logic based on original first argument (if it was a string)
        if (p.length > 0 && isString(p0Original)) {
            // Convert original first arg to Unix style for prefix checks
            const p0Unix = toUnix(p0Original)

            // Restore './' prefix if original first arg started with it, result doesn't, and isn't '..' or '/'
            if (p0Unix.startsWith("./") && !result.startsWith("./") && !upath_internal.isAbsolute(result) && result !== "..") {
                return "./" + result
            }
            // Restore '//' prefix if original first arg started with it and result lost it or changed it
            else if (p0Unix.startsWith("//") && !result.startsWith("//")) {
                if (p0Unix.startsWith("//./")) {
                    return "//." + result
                } else {
                    return "/" + result
                }
            }
        }
        // Otherwise, return the result from upath.join
        return result
    },

    /**
     * Adds an extension to a filename if it doesn't already have it.
     * Ensures the added extension starts with '.'.
     * @param file The filename.
     * @param ext The extension to add (e.g., 'js' or '.js').
     * @returns The filename with the extension added.
     */
    addExt: (file: string, ext?: string): string => {
        if (!ext) {
            return file
        } else {
            // Ensure extension starts with '.'
            ext = (ext[0] !== "." ? "." : "") + ext
            // Check if file already ends with the exact extension case-sensitively
            if (file.endsWith(ext)) {
                return file
            } else {
                return file + ext
            }
        }
    },

    /**
     * Removes the extension from a filename *only if* it is considered a valid
     * extension based on the ignore list and max size.
     * @param filename The filename.
     * @param ignoreExts Array of extensions to ignore (e.g., ['js', '.txt']).
     * @param maxSize Maximum allowed extension length.
     * @returns The filename without the valid extension, or the original filename.
     */
    trimExt: (filename: string, ignoreExts?: string[], maxSize?: number): string => {
        const oldExt = upath_internal.extname(filename) // Use the implemented extname
        if (isValidExt(oldExt, ignoreExts, maxSize)) {
            // Remove the extension part by slicing before where the extension starts
            return filename.slice(0, filename.length - oldExt.length)
        } else {
            return filename // No valid extension to trim, return original filename
        }
    },

    /**
     * Removes a *specific* extension from a filename *only if* it matches
     * the current extension exactly (case-sensitively).
     * @param filename The filename.
     * @param ext The specific extension to remove (e.g., 'js' or '.js').
     * @returns The filename without the specified extension, or the original filename.
     */
    removeExt: (filename: string, ext?: string): string => {
        if (!ext) {
            return filename
        } else {
            // Ensure the target extension starts with '.'
            ext = (ext[0] === "." ? ext : "." + ext)
            // Use implemented extname for consistency in checking
            if (upath_internal.extname(filename) === ext) {
                // If the current extension matches the target extension, remove it
                return filename.slice(0, filename.length - ext.length)
            } else {
                return filename // Current extension does not match the target extension, return original
            }
        }
    },

    /**
     * Changes the extension of a filename. It first removes the existing
     * extension (if it's considered valid by trimExt rules), then adds the new one.
     * @param filename The filename.
     * @param ext The new extension (e.g., 'js' or '.js'). Can be empty to remove extension.
     * @param ignoreExts Array of extensions to ignore for trimming the old one.
     * @param maxSize Maximum allowed extension length for trimming the old one.
     * @returns The filename with the extension changed.
     */
    changeExt: (filename: string, ext?: string, ignoreExts?: string[], maxSize?: number): string => {
        // Trim the existing valid extension first using trimExt logic
        const trimmed = upath_internal.trimExt(filename, ignoreExts, maxSize)
        // Add the new extension if specified
        if (!ext) {
            return trimmed // No new extension specified, just return trimmed filename
        } else {
            // Ensure the new extension starts with '.'
            ext = (ext[0] === "." ? ext : "." + ext)
            return trimmed + ext // Append the new extension
        }
    },

    /**
     * Adds a default extension to a filename *only if* the existing extension
     * is *not* considered valid based on ignore list and max size.
     * @param filename The filename.
     * @param ext The default extension to add (e.g., 'js' or '.js').
     * @param ignoreExts Array of extensions to ignore for checking validity.
     * @param maxSize Maximum allowed extension length for checking validity.
     * @returns The filename with the default extension added if needed, or the original filename.
     */
    defaultExt: (filename: string, ext?: string, ignoreExts?: string[], maxSize?: number): string => {
        const oldExt = upath_internal.extname(filename) // Use implemented extname
        // If existing extension is NOT valid, add the default extension
        if (!isValidExt(oldExt, ignoreExts, maxSize)) {
            return upath_internal.addExt(filename, ext) // Use addExt to ensure format is correct
        } else {
            return filename // Existing valid extension found, return original filename
        }
    },
}

// Add extra functions to upath_internal, checking for name conflicts
for (const name in extraFunctions) {
    if (Object.prototype.hasOwnProperty.call(extraFunctions, name)) {
        const extraFn = (extraFunctions as any)[name]

        if (upath_internal[name] !== undefined) {
            // Throw an error if the name already exists on upath_internal
            throw new Error(`path.${name} already exists.`)
        } else {
            // Assign the extra function
            upath_internal[name] = extraFn
        }
    }
}

// Export the internal object, casting it to the UPath interface for correct typing on export.
export const upath: UPath = upath_internal as UPath
