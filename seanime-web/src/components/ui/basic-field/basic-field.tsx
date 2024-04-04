import { cva } from "class-variance-authority"
import * as React from "react"
import { cn, ComponentAnatomy, defineStyleAnatomy } from "../core/styling"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const BasicFieldAnatomy = defineStyleAnatomy({
    fieldLabel: cva([
        "UI-BasicField__fieldLabel",
        "text-base w-fit font-semibold self-start",
        "data-[error=true]:text-red-500",
    ]),
    fieldAsterisk: cva("UI-BasicField__fieldAsterisk ml-1 text-red-500 text-sm"),
    fieldDetails: cva("UI-BasicField__fieldDetails"),
    field: cva("UI-BasicField__field relative w-full space-y-1"),
    fieldHelpText: cva("UI-BasicField__fieldHelpText text-sm text-[--muted]"),
    fieldErrorText: cva("UI-BasicField__fieldErrorText text-sm text-red-500"),
})

/* -------------------------------------------------------------------------------------------------
 * BasicFieldOptions
 * - Field components inherit these props
 * -----------------------------------------------------------------------------------------------*/

export type BasicFieldOptions = ComponentAnatomy<typeof BasicFieldAnatomy> & {
    /**
     * The id of the field. If not provided, a unique id will be generated.
     */
    id?: string | undefined
    /**
     * The form field name.
     */
    name?: string
    /**
     * The label of the field.
     */
    label?: React.ReactNode
    /**
     * Additional props to pass to the label element.
     */
    labelProps?: React.LabelHTMLAttributes<HTMLLabelElement>
    /**
     * Help or description text to display below the field.
     */
    help?: React.ReactNode
    /**
     * Error text to display below the field.
     */
    error?: string
    /**
     * If `true`, the field will be required.
     */
    required?: boolean
    /**
     * If `true`, the field will be disabled.
     */
    disabled?: boolean
    /**
     * If `true`, the field will be readonly.
     */
    readonly?: boolean
}

/* -------------------------------------------------------------------------------------------------
 * Extract BasicFieldProps
 * -----------------------------------------------------------------------------------------------*/

export function extractBasicFieldProps<Props extends BasicFieldOptions>(props: Props, id: string) {
    const {
        name,
        label,
        labelProps,
        help,
        error,
        required,
        disabled = false,
        readonly = false,
        fieldDetailsClass,
        fieldLabelClass,
        fieldAsteriskClass,
        fieldClass,
        fieldErrorTextClass,
        fieldHelpTextClass,
        id: _id,
        ...rest
    } = props
    return [
        rest,
        {
            id: _id || id,
            name,
            label,
            help,
            error,
            disabled,
            required,
            readonly,
            fieldAsteriskClass,
            fieldErrorTextClass,
            fieldHelpTextClass,
            fieldDetailsClass,
            fieldLabelClass,
            fieldClass,
            labelProps,
        },
    ] as [
        Omit<Props,
            "label" | "name" | "help" | "error" |
            "disabled" | "required" | "readonly" |
            "fieldDetailsClass" | "fieldLabelClass" | "fieldClass" | "fieldHelpTextClass" |
            "fieldErrorTextClass" | "id" | "labelProps" | "fieldAsteriskClass"
        >,
            Omit<BasicFieldOptions, "id"> & {
            id: string
        }
    ]
}

/* -------------------------------------------------------------------------------------------------
 * BasicField
 * -----------------------------------------------------------------------------------------------*/

export type BasicFieldProps = React.ComponentPropsWithoutRef<"div"> & BasicFieldOptions

export const BasicField = React.memo(React.forwardRef<HTMLDivElement, BasicFieldProps>((props, ref) => {

    const {
        children,
        className,
        labelProps,
        id,
        label,
        error,
        help,
        disabled,
        readonly,
        required,
        fieldClass,
        fieldDetailsClass,
        fieldLabelClass,
        fieldAsteriskClass,
        fieldErrorTextClass,
        fieldHelpTextClass,
        ...rest
    } = props

    return (
        <div
            className={cn(
                BasicFieldAnatomy.field(),
                className,
                fieldClass,
            )}
            {...rest}
            ref={ref}
        >
            {!!label &&
                <label
                    htmlFor={disabled ? undefined : id}
                    className={cn(BasicFieldAnatomy.fieldLabel(), fieldLabelClass)}
                    data-error={!!error}
                    {...labelProps}
                >
                    {label}
                    {required &&
                        <span className={cn(BasicFieldAnatomy.fieldAsterisk(), fieldAsteriskClass)}>*</span>
                    }
                </label>
            }

            {children}

            {(!!help || !!error) &&
                <div className={cn(BasicFieldAnatomy.fieldDetails(), fieldDetailsClass)}>
                    {!!help &&
                        <div className={cn(BasicFieldAnatomy.fieldHelpText(), fieldHelpTextClass)}>{help}</div>}
                    {!!error &&
                        <div className={cn(BasicFieldAnatomy.fieldErrorText(), fieldErrorTextClass)}>{error}</div>}
                </div>
            }
        </div>
    )

}))

BasicField.displayName = "BasicField"
