import { ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"
import React from "react"

/* -------------------------------------------------------------------------------------------------
 * Tailwind
 * -----------------------------------------------------------------------------------------------*/

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}

/* -------------------------------------------------------------------------------------------------
 * Polymorphic component
 * -----------------------------------------------------------------------------------------------*/

type ExtendedProps<Props = {}, OverrideProps = {}> = OverrideProps &
    Omit<Props, keyof OverrideProps>;
type ElementType = keyof JSX.IntrinsicElements | React.JSXElementConstructor<any>;
type PropsOf<C extends ElementType> = JSX.LibraryManagedAttributes<C,
    React.ComponentPropsWithoutRef<C>>;
type ComponentProp<C> = {
    component?: C;
};
type InheritedProps<C extends ElementType, Props = {}> = ExtendedProps<PropsOf<C>, Props>;
export type PolymorphicRef<C> = C extends React.ElementType
    ? React.ComponentPropsWithRef<C>["ref"]
    : never;
export type PolymorphicComponentProps<C, Props = {}> = C extends React.ElementType
    ? InheritedProps<C, Props & ComponentProp<C>> & { ref?: PolymorphicRef<C> }
    : Props & { component: React.ElementType };

/**
 * @example
 * const _Accordion = React.forwardRef<HTMLDivElement, AccordionProps>((props, ref) => {})
 * _Accordion.Item = AccordionItem
 *
 * export const Accordion = createPolymorphicComponent<'div', AccordionProps, {
 *    Item: typeof AccordionItem,
 * }>(_Accordion)
 * @param component
 */
export function createPolymorphicComponent<ComponentDefaultType,
    Props,
    StaticComponents = Record<string, never>>(component: any) {
    type ComponentProps<C> = PolymorphicComponentProps<C, Props>;

    type _PolymorphicComponent = <C = ComponentDefaultType>(
        props: ComponentProps<C>,
    ) => React.ReactElement;

    type ComponentProperties = Omit<React.FunctionComponent<ComponentProps<any>>, never>;

    type PolymorphicComponent = _PolymorphicComponent & ComponentProperties & StaticComponents;

    return component as PolymorphicComponent
}

/* -------------------------------------------------------------------------------------------------
 * Display Name
 * -----------------------------------------------------------------------------------------------*/

export const getChildDisplayName = (child: string | number | boolean | React.ReactElement<any, string | React.JSXElementConstructor<any>> | React.ReactFragment | React.ReactPortal | null | undefined) => {
    return (child as any)?.type?.displayName as (string | undefined)
}

/* -------------------------------------------------------------------------------------------------
 * Refs
 * -----------------------------------------------------------------------------------------------*/

export function mergeRefs<T>(...refs: React.ForwardedRef<T>[]) {
    const targetRef = React.useRef<T>(null)

    React.useLayoutEffect(() => {
        refs.forEach(ref => {
            if (!ref) return
            if (typeof ref === "function") {
                ref(targetRef.current)
            } else {
                ref.current = targetRef.current
            }
        })
    }, [refs])

    return targetRef
}
