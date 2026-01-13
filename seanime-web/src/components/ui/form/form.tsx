"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { NoInfer } from "@tanstack/react-query"
import * as React from "react"
import { FieldErrors, FieldValues, FormProvider, useForm, UseFormProps, UseFormReturn, WatchObserver } from "react-hook-form"
import { z } from "zod"
import { cn } from "../core/styling"
import { isEmpty } from "../core/utils"
import { getZodDefaults } from "./zod-resolver"

/* -------------------------------------------------------------------------------------------------
 * Context
 * -----------------------------------------------------------------------------------------------*/

/**
 * @internal
 */
const __FormSchemaContext = React.createContext<{
    shape: z.ZodRawShape,
    schema: z.ZodObject<z.ZodRawShape>
} | undefined>(undefined)

export const useFormSchema = (): { shape: z.ZodRawShape, schema: z.ZodObject<z.ZodRawShape> } => {
    return React.useContext(__FormSchemaContext)!
}

export type SubmitHandler<T> = (data: T, event?: React.BaseSyntheticEvent) => any
export type SubmitErrorHandler<TFieldValues extends FieldValues> = (errors: FieldErrors<TFieldValues>, event?: React.BaseSyntheticEvent) => any

/* -------------------------------------------------------------------------------------------------
 * Form
 * -----------------------------------------------------------------------------------------------*/

export type FormProps<Schema extends z.ZodObject<z.ZodRawShape> = z.ZodObject<z.ZodRawShape>> =
    UseFormProps<NoInfer<z.infer<Schema>>> &
    Omit<React.ComponentPropsWithRef<"form">, "children" | "onChange" | "onSubmit" | "onError" | "ref"> & {
    /**
     * The schema of the form.
     */
    schema: Schema
    /**
     * Callback invoked when the form is submitted.
     */
    onSubmit: SubmitHandler<NoInfer<z.infer<Schema>>>
    /**
     * Callback invoked when any of the field change.
     */
    onChange?: WatchObserver<NoInfer<z.infer<Schema>>>
    /**
     * Callback invoked when there are validation errors.
     */
    onError?: SubmitErrorHandler<any>
    /**
     * Ref to the form element.
     */
    formRef?: React.RefObject<HTMLFormElement>

    children?: MaybeRenderProp<UseFormReturn<NoInfer<z.infer<Schema>>>>
    /**
     * @default w-full space-y-3
     */
    stackClass?: string
    /**
     * Ref to the form methods.
     */
    mRef?: React.Ref<UseFormReturn<NoInfer<z.infer<Schema>>>>
}

export const Form = <Schema extends z.ZodObject<z.ZodRawShape>>(props: FormProps<Schema>) => {

    const {
        mode = "onSubmit",
        resolver,
        reValidateMode,
        shouldFocusError,
        shouldUnregister,
        shouldUseNativeValidation,
        criteriaMode,
        delayError,
        schema,
        defaultValues: _defaultValues,
        onChange,
        onSubmit,
        onError,
        formRef,
        children,
        mRef,
        /**/
        stackClass,
        ...rest
    } = props

    const defaultValues = React.useMemo(() => {
        if (isEmpty(getZodDefaults(schema)) && isEmpty(_defaultValues)) return undefined
        return {
            ...getZodDefaults(schema),
            ..._defaultValues,
        } as any
    }, [])

    const form = {
        mode,
        resolver,
        defaultValues,
        reValidateMode,
        shouldFocusError,
        shouldUnregister,
        shouldUseNativeValidation,
        criteriaMode,
        delayError,
    }

    form.resolver = zodResolver(schema)

    const methods = useForm(form)
    const { handleSubmit } = methods

    React.useImperativeHandle(mRef, () => methods, [mRef, methods])

    React.useEffect(() => {
        let subscription: ReturnType<typeof methods.watch> | undefined
        if (onChange) {
            subscription = methods.watch(onChange)
        }
        return () => subscription?.unsubscribe()
    }, [methods, onChange])

    return (
        <FormProvider {...methods}>
            <__FormSchemaContext.Provider value={{ schema, shape: schema.shape }}>
                <form
                    ref={formRef}
                    onSubmit={handleSubmit(onSubmit, onError)}
                    {...rest}
                >
                    <div className={cn("w-full space-y-3", stackClass)}>
                        {runIfFn(children, methods)}
                    </div>
                </form>
            </__FormSchemaContext.Provider>
        </FormProvider>
    )

}

Form.displayName = "Form"

/* -------------------------------------------------------------------------------------------------
 * Utils
 * -----------------------------------------------------------------------------------------------*/

type MaybeRenderProp<P> =
    | React.ReactNode
    | ((props: P) => React.ReactNode)

const isFunction = <T extends Function = Function>(value: any): value is T => typeof value === "function"

function runIfFn<T, U>(
    valueOrFn: T | ((...fnArgs: U[]) => T),
    ...args: U[]
): T {
    return isFunction(valueOrFn) ? valueOrFn(...args) : valueOrFn
}
