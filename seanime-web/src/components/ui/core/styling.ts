import { cva } from "class-variance-authority"
import { ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export type Anatomy = { [key: string]: ReturnType<typeof cva> }

// export type ComponentAnatomy<T extends Anatomy> = {
//     [K in keyof T as `${string & K}Class`]?: string
// }

export type ComponentAnatomy<T extends Anatomy> = {
    [K in keyof T as K extends "root" ? never : `${string & K}Class`]?: string;
}

/**
 * @example
 * const ComponentAnatomy = defineStyleAnatomy({
 *    label: cva(null, {
 *       variants: {
 *          intent: {
 *             "success": "",
 *             "alert": "",
 *          },
 *       },
 *    }),
 *    ...
 * })
 *
 * type ComponentProps = ComponentWithAnatomy<typeof ComponentAnatomy>
 *
 * const MyComponent = React.forwardRef((props, forwardedRef) => {
 *   const { controlClass, ...rest }: ComponentProps = props
 *
 *   return (
 *      <div
 *          className={cn(ComponentAnatomy.control({ intent: "success" }, controlClass))}
 *          ref={forwardedRef}
 *      />
 *   )
 * })
 */
export function defineStyleAnatomy<A extends Anatomy = Anatomy>(config: A) {
    return config
}

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs))
}
