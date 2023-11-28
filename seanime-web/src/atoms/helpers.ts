import { atom, Atom, PrimitiveAtom } from "jotai"
import { useAtomValue, useSetAtom } from "jotai/react"
import { focusAtom } from "jotai-optics"
import { useCallback } from "react"
import { OpticFor_ } from "optics-ts"
import { selectAtom } from "jotai/utils"
import deepEquals from "fast-deep-equal"


export function useSelectAtom<T, R>(anAtom: PrimitiveAtom<T> | Atom<T>, keyFn: (v: T) => R) {
    return useAtomValue(
        selectAtom(
            anAtom,
            useCallback(keyFn, []),
            deepEquals,
        ),
    )
}

const _dummy = atom(null)

/**
 * Select from an atom that might be undefined.
 * Ensures that the condition that determines the callback's definition remains stable across renders
 *
 * @example
 * const anAtom = condition ? getAtom() : undefined
 * const value = useStableSelectAtom(anAtom, n => n.property)
 *
 * @param anAtom
 * @param keyFn
 */
export function useStableSelectAtom<T, R>(anAtom: PrimitiveAtom<T> | Atom<T> | undefined, keyFn: (v: T) => R | null) {
    return (anAtom ? useAtomValue(
        selectAtom(
            anAtom,
            useCallback(keyFn, []),
            deepEquals,
        ),
    ) : useAtomValue(
        selectAtom(
            _dummy,
            useCallback(() => null, []),
            deepEquals,
        ),
    )) as R | null
}

export const useFocusSetAtom = <T>(anAtom: PrimitiveAtom<T>, prop: keyof T) => {
    return useSetAtom(
        focusAtom(
            anAtom,
            useCallback((optic: OpticFor_<T>) => optic.prop(prop), []),
        ),
    )
}
