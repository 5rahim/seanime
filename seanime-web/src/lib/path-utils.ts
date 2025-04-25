import path from "path-browserify"

/**
 * A replacement for upath that doesn't use lookbehind regex
 * This utility provides cross-platform path manipulation functions
 */

/**
 * Normalizes a path, converting backslashes to forward slashes
 */
export function normalize(pathStr: string): string {
    return path.normalize(pathStr)
}

export function normalizeSafe(pathStr: string): string {
    if (!pathStr) return ""
    return path.normalize(pathStr)
}

/**
 * Joins path segments together and normalizes the result
 */
export function join(...paths: string[]): string {
    return path.join(...paths)
}

/**
 * Returns the directory name of a path
 */
export function dirname(pathStr: string): string {
    return path.dirname(pathStr)
}

/**
 * Returns the last portion of a path
 */
export function basename(pathStr: string): string {
    return path.basename(pathStr)
}
