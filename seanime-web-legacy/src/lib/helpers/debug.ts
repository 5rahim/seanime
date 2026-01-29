import chalk from "chalk"
import React from "react"

export const logger = (prefix: string, silence?: boolean) => {

    return {
        info: (...data: any[]) => {
            if (silence) return
            console.log(chalk.blue(`[${prefix}]`) + " ", ...data)
        },
        warning: (...data: any[]) => {
            if (silence) return
            console.log(chalk.yellow(`[${prefix}]`) + " ", ...data)
        },
        warn: (...data: any[]) => {
            if (silence) return
            console.log(chalk.yellow(`[${prefix}]`) + " ", ...data)
        },
        success: (...data: any[]) => {
            if (silence) return
            console.log(chalk.green(`[${prefix}]`) + " ", ...data)
        },
        error: (...data: any[]) => {
            if (silence) return
            console.log(chalk.red(`[${prefix}]`) + " ", ...data)
        },
        trace: (...data: any[]) => {
            if (silence || process.env.NODE_ENV !== "development") return
            console.log(chalk.bgGray(`[${prefix}]`) + " ", ...data)
        },
    }

}

export const useEffectDebugger = (
    effectHook: () => void | (() => void),
    dependencies: any[],
    dependencyNames: string[] = [],
) => {
    const previousDeps = React.useRef(dependencies)

    React.useEffect(() => {
        const changedDeps = dependencies.reduce((accum, dependency, index) => {
            if (dependency !== previousDeps.current[index]) {
                const keyName = dependencyNames[index] || `Dependency #${index}`
                return {
                    ...accum,
                    [keyName]: {
                        before: previousDeps.current[index],
                        after: dependency,
                    },
                }
            }
            return accum
        }, {})

        if (Object.keys(changedDeps).length) {
            console.log("[useEffectDebugger] Changed dependencies:", changedDeps)
        }

        previousDeps.current = dependencies

        return effectHook()
    }, dependencies) // Pass the original dependencies to useEffect
}

export function useLatestFunction<T extends (...args: any[]) => any>(callback: T): T {
    const callbackRef = React.useRef(callback)

    React.useLayoutEffect(() => {
        callbackRef.current = callback
    }, [callback])

    return React.useCallback(((...args: Parameters<T>): ReturnType<T> => {
        return callbackRef.current(...args)
    }) as T, [])
}
