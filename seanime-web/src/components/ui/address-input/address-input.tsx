"use client"

import { LoaderOptions } from "@googlemaps/js-api-loader"
import _isEmpty from "lodash/isEmpty"
import React, { useId } from "react"
import { extractBasicFieldProps } from "../basic-field"
import { Combobox, ComboboxProps } from "../combobox"
import { useUILocaleConfig } from "../core"
import locales from "./locales.json"
import { GoogleMapsAutocompletionRequest, useGoogleMapsAutocomplete } from "./use-address-autocomplete"

/* -------------------------------------------------------------------------------------------------
 * AddressInput
 * -----------------------------------------------------------------------------------------------*/

export interface AddressInputProps extends Omit<ComboboxProps, "options" | "onInputChange" | "onChange"> {
    autocompletionRequest?: GoogleMapsAutocompletionRequest
    apiOptions?: Partial<LoaderOptions>
    allowedCountries?: string | string[] | null
    onChange?: (value: string | undefined) => void
    noOptionsMessage?: string
    placeholder?: string
    apiKey: string // Optionally, you could remove this parameter and get the key from environment variables
}

export const AddressInput = React.forwardRef<HTMLInputElement, AddressInputProps>((props, ref) => {

    const { locale: lng } = useUILocaleConfig()

    const [{
        children,
        className,
        autocompletionRequest,
        apiOptions,
        defaultValue,
        allowedCountries = null,
        onChange,
        apiKey,
        placeholder = locales["placeholder"][lng],
        noOptionsMessage = locales["no-address-found"][lng],
        ...rest
    }, basicFieldProps] = extractBasicFieldProps<AddressInputProps>(props, useId())

    const { suggestions, fetchSuggestions } = useGoogleMapsAutocomplete({
        apiKey: apiKey,
        minLengthAutocomplete: 0,
        withSessionToken: false,
        debounce: 300,
        autocompletionRequest: {
            componentRestrictions: { country: allowedCountries },
        },
    })

    return (
        <>
            <Combobox
                returnValueOrLabel="label" // We only return the address' text format
                allowCustomValue={false}
                withFiltering={false} // We deactivate filtering because the options are automatically filtered by the API
                options={_isEmpty(suggestions) && defaultValue ? [{
                    value: defaultValue,
                    label: defaultValue
                }] : suggestions}
                onInputChange={fetchSuggestions}
                defaultValue={defaultValue}
                onChange={onChange}
                placeholder={placeholder}
                noOptionsMessage={noOptionsMessage}
                {...basicFieldProps}
                {...rest}
                ref={ref}
            />
        </>
    )

})
