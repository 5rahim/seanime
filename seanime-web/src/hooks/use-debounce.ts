import React, { useCallback, useEffect, useRef, useState } from "react"

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

export function useDebounceWithTrigger<T>(initialValue: T, delay?: number): {
    value: T
    debouncedValue: T
    setValue: React.Dispatch<React.SetStateAction<T>>
    triggerImmediate: (val?: T) => void
} {
    const actualDelay = delay ?? 500
    const [value, setValue] = useState<T>(initialValue)
    const [debouncedValue, setDebouncedValue] = useState<T>(initialValue)
    const timerRef = useRef<any>(null)

    // Sync state if initialValue changes from parent
    const lastInitialValue = useRef(initialValue)
    if (initialValue !== lastInitialValue.current) {
        setValue(initialValue)
        setDebouncedValue(initialValue)
        lastInitialValue.current = initialValue
    }

    const clearTimer = useCallback(() => {
        if (timerRef.current) {
            clearTimeout(timerRef.current)
            timerRef.current = null
        }
    }, [])

    useEffect(() => {
        clearTimer()
        timerRef.current = setTimeout(() => {
            setDebouncedValue(value)
        }, actualDelay)

        return () => clearTimer()
    }, [value, actualDelay, clearTimer])

    const triggerImmediate = useCallback((val?: T) => {
        clearTimer()
        const targetValue = val !== undefined ? val : value
        setDebouncedValue(targetValue)
    }, [value, clearTimer])

    return {
        value,
        debouncedValue,
        setValue,
        triggerImmediate,
    }
}
