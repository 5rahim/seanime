import { cn, ComponentWithAnatomy, defineStyleAnatomy } from "../core"
import { cva } from "class-variance-authority"
import React from "react"

/* -------------------------------------------------------------------------------------------------
 * Anatomy
 * -----------------------------------------------------------------------------------------------*/

export const BasicFieldAnatomy = defineStyleAnatomy({
    fieldLabel: cva([
        "UI-BasicField__fieldLabel",
        "block text-md sm:text-lg font-semibold self-start"
    ], {
        variants: {
            hasError: {
                true: "text-red-500",
                false: null,
            },
        },
    }),
    fieldAsterisk: cva("UI-BasicField__fieldAsterisk ml-1 text-red-500 text-sm"),
    fieldDetails: cva("UI-BasicField__fieldDetails"),
    field: cva("UI-BasicField__field w-full space-y-1"),
    fieldHelpText: cva("UI-BasicField__fieldHelpText text-sm text-gray-500"),
    fieldErrorText: cva("UI-BasicField__fieldErrorText text-sm text-red-500"),
})

/* -------------------------------------------------------------------------------------------------
 * BasicFieldOptions
 * - Field components inherit these props
 * -----------------------------------------------------------------------------------------------*/

export interface BasicFieldOptions extends ComponentWithAnatomy<typeof BasicFieldAnatomy> {
    id?: string | undefined
    name?: string
    label?: React.ReactNode
    labelProps?: object
    help?: React.ReactNode
    error?: string
    isRequired?: boolean
    isDisabled?: boolean
    isReadOnly?: boolean
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
        isRequired,
        isDisabled = false,
        isReadOnly = false,
        fieldDetailsClassName,
        fieldLabelClassName,
        fieldClassName,
        fieldErrorTextClassName,
        fieldHelpTextClassName,
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
            isDisabled,
            isRequired,
            isReadOnly,
            fieldErrorTextClassName,
            fieldHelpTextClassName,
            fieldDetailsClassName,
            fieldLabelClassName,
            fieldClassName,
            labelProps,
        },
    ] as [
        Omit<Props,
            "label" | "name" | "help" | "error" |
            "isDisabled" | "isRequired" | "isReadOnly" |
            "fieldDetailsClassName" | "fieldLabelClassName" | "fieldClassName" | "fieldHelpTextClassName" |
            "fieldErrorTextClassName" | "id" | "labelProps"
        >,
            Omit<BasicFieldOptions, "id"> & {
            id: string
        }
    ]
}

/* -------------------------------------------------------------------------------------------------
 * BasicField
 * -----------------------------------------------------------------------------------------------*/

export interface BasicFieldProps extends React.ComponentPropsWithRef<"div">, BasicFieldOptions {
}

export const BasicField: React.FC<BasicFieldProps> = React.memo(React.forwardRef<HTMLDivElement, BasicFieldProps>((props, ref) => {

    const {
        children,
        className,
        labelProps,
        id,
        label,
        error,
        help,
        isDisabled,
        isReadOnly,
        isRequired,
        fieldClassName,
        fieldDetailsClassName,
        fieldLabelClassName,
        fieldAsteriskClassName,
        fieldErrorTextClassName,
        fieldHelpTextClassName,
        ...rest
    } = props

    return (
        <>
            <div
                className={cn(
                    BasicFieldAnatomy.field(),
                    className,
                    fieldClassName,
                )}
                {...rest}
                ref={ref}
            >
                {!!label &&
                    <label
                        htmlFor={isDisabled ? undefined : id}
                        className={cn(BasicFieldAnatomy.fieldLabel({ hasError: !!error }), fieldLabelClassName)}
                        {...labelProps}
                    >
                        {label}
                        {isRequired &&
                            <span className={cn(BasicFieldAnatomy.fieldAsterisk(), fieldAsteriskClassName)}>*</span>
                        }
                    </label>
                }

                {children}

                {(!!help || !!error) &&
                    <div className={cn(BasicFieldAnatomy.fieldDetails(), fieldDetailsClassName)}>
                        {!!help &&
                            <p className={cn(BasicFieldAnatomy.fieldHelpText(), fieldHelpTextClassName)}>{help}</p>}
                        {!!error &&
                            <p className={cn(BasicFieldAnatomy.fieldErrorText(), fieldErrorTextClassName)}>{error}</p>}
                    </div>
                }
            </div>
        </>
    )

}))

BasicField.displayName = "BasicField"
