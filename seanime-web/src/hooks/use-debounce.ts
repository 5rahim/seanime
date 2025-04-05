import React, { useEffect, useState } from "react"

export function useDebounce<T>(value: T, delay?: number): T {
    const [debouncedValue, setDebouncedValue] = useState<T>(value)

    useEffect(() => {
        const timer = setTimeout(() => setDebouncedValue(value), delay || 500)

        return () => {
            clearTimeout(timer)
        }
    }, [value, delay])

    return debouncedValue
}

export function useDebounceWithSet<T>(value: T, delay?: number): [T, T, React.Dispatch<React.SetStateAction<T>>] {
    const [actualValue, setActualValue] = useState<T>(value)
    const debouncedValue = useDebounce(actualValue, delay)

    return [actualValue, debouncedValue, setActualValue]
}