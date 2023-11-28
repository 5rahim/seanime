"use client"

import { zodResolver } from "@hookform/resolvers/zod"
import { cn } from "../core"
import _isEmpty from "lodash/isEmpty"
import React, { createContext, useContext, useEffect, useMemo } from "react"
import {
    FormProvider,
    SubmitErrorHandler,
    SubmitHandler,
    useForm,
    UseFormProps,
    UseFormReturn,
    WatchObserver,
} from "react-hook-form"
import { z, ZodObject } from "zod"
import { getZodDefaults } from "./zod-resolver"

/* -------------------------------------------------------------------------------------------------
 * Context
 * -----------------------------------------------------------------------------------------------*/

/**
 * @internal
 */
const __FormSchemaContext = createContext<{
    shape: z.ZodRawShape,
    schema: z.ZodObject<z.ZodRawShape>
} | undefined>(undefined)

export const useFormSchema = (): { shape: z.ZodRawShape, schema: z.ZodObject<z.ZodRawShape> } => {
    return useContext(__FormSchemaContext)!
}

/* -------------------------------------------------------------------------------------------------
 * TypesafeForm
 * -----------------------------------------------------------------------------------------------*/

export interface TypesafeFormProps<Schema extends z.ZodObject<z.ZodRawShape> = ZodObject<any>>
    extends UseFormProps<z.infer<Schema>>,
        Omit<React.ComponentPropsWithRef<"form">, "children" | "onChange" | "onSubmit" | "onError" | "ref"> {
    schema: Schema
    onSubmit: SubmitHandler<z.infer<Schema>>
    onChange?: WatchObserver<z.infer<Schema>> // Triggers when any of the field change.
    onError?: SubmitErrorHandler<z.infer<Schema>> // Triggers when there are validation errors.
    formRef?: React.RefObject<HTMLFormElement>
    children?: MaybeRenderProp<UseFormReturn<z.infer<Schema>>>
    /**
     * @default w-full space-y-3
     */
    stackClassName?: string
    mRef?: React.Ref<UseFormReturn<z.infer<Schema>>>
}

/**
 * @example
 * <TypesafeForm
 *     schema={definedSchema}
 *     onSubmit={console.log}
 *     onError={console.log}
 *     onChange={console.log}
 *     defaultValues={undefined}
 *  >
 *     <Field.Submit role="create" />
 *  </TypesafeForm>
 * @param props
 * @constructor
 */
export const TypesafeForm = <Schema extends z.ZodObject<z.ZodRawShape>>(props: TypesafeFormProps<Schema>) => {

    const {
        mode = "onTouched",
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

        stackClassName,
        ...rest
    } = props

    const defaultValues = useMemo(() => {
        if (_isEmpty(getZodDefaults(schema)) && _isEmpty(_defaultValues)) return undefined
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

    useEffect(() => {
        let subscription: any
        if (onChange) {
            subscription = methods.watch(onChange)
        }
        return () => subscription?.unsubscribe()
    }, [methods, onChange])

    return (
        <>
            <FormProvider {...methods}>
                <__FormSchemaContext.Provider value={{ schema, shape: schema.shape }}>
                    <form
                        ref={formRef}
                        onSubmit={handleSubmit(onSubmit, onError)}
                        {...rest}
                    >
                        <div className={cn("w-full space-y-3", stackClassName)}>
                            {runIfFn(children, methods)}
                        </div>
                    </form>
                </__FormSchemaContext.Provider>
            </FormProvider>
        </>
    )

}

TypesafeForm.displayName = "TypesafeForm"

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
