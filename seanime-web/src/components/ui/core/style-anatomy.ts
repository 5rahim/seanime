import { cva } from "class-variance-authority"
/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

/**
 * @internal UI Folder scope
 */
export type Anatomy = { [key: string]: ReturnType<typeof cva> }
/**
 * @internal
 */
export type AnatomyClassNames<T extends Anatomy> = {
    [K in keyof T as `${string & K}ClassName`]?: string
}

/**
 * @internal UI Folder scope
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
 * // const { controlClassName, ...rest }: ComponentProps = props
 * <div className={cn(ComponentAnatomy.control({ intent: "success" }, controlClassName))} />
 * @param config
 */
export function defineStyleAnatomy<A extends Anatomy = Anatomy>(config: A) {
    return config
}

/**
 * @internal UI Folder scope
 */
export type ComponentWithAnatomy<T extends Anatomy> = AnatomyClassNames<T>

