"use client"

import { Loader, LoaderOptions } from "@googlemaps/js-api-loader"
import * as React from "react"
import { useDebouncedCallback } from "use-debounce"

export type GoogleMapsAutocompletionRequest = Omit<google.maps.places.AutocompletionRequest, "input">

export type GoogleMapsAutocompleteProps = {
    /**
     * Google Maps API key
     * @see https://developers.google.com/maps/documentation/javascript/get-api-key
     */
    apiKey: string,
    /**
     * Minimum length of the input before triggering the autocompletion
     */
    minLengthAutocomplete?: number,
    /**
     * Whether to use a session token for the autocompletion
     * @see https://developers.google.com/maps/documentation/javascript/reference/places-autocomplete-service#AutocompletionRequest.sessionToken
     */
    withSessionToken?: boolean
    /**
     * Debounce time in ms
     */
    debounce?: number
    /**
     * Autocompletion request
     */
    autocompletionRequest?: GoogleMapsAutocompletionRequest
    /**
     * Loader options
     */
    loaderApiOptions?: Partial<LoaderOptions>
}

export const useGoogleMapsAutocomplete = ({
    apiKey,
    minLengthAutocomplete = 0,
    withSessionToken = false,
    debounce = 300,
    autocompletionRequest,
    loaderApiOptions = {},
}: GoogleMapsAutocompleteProps) => {

    const [autocompleteService, setAutocompleteService] = React.useState<google.maps.places.AutocompleteService | undefined>(undefined)
    const [sessionToken, setSessionToken] = React.useState<google.maps.places.AutocompleteSessionToken | undefined>(undefined)

    const initializeService = React.useCallback(() => {
        if (!window.google) throw new Error("[AddressInput]: Google script not loaded")
        if (!window.google.maps) throw new Error("[AddressInput]: Google maps script not loaded")
        if (!window.google.maps.places) throw new Error("[AddressInput]: Google maps places script not loaded")

        setAutocompleteService(new window.google.maps.places.AutocompleteService())
        setSessionToken(new google.maps.places.AutocompleteSessionToken())
    }, [])

    // Initialize service
    React.useEffect(() => {
        if (!apiKey) {
            console.warn("[AddressInput]: No API key provided")
            return
        }
        (async () => {
            try {
                if (!window.google || !window.google.maps || !window.google.maps.places) {
                    await new Loader({ apiKey: apiKey, ...{ libraries: ["places"], ...loaderApiOptions } }).load()
                }
                initializeService()
            }
            catch (error) {
                console.error(error)
            }
        })()
    }, [])


    // Fetch suggestions
    const [suggestions, setSuggestions] = React.useState<{ label: string, value: string }[]>([])
    const [isFetching, setIsFetching] = React.useState<boolean>(false)

    const fetchSuggestions = useDebouncedCallback((value: string): void => {
        if (!autocompleteService) return setSuggestions([])
        if (value.length < minLengthAutocomplete) return setSuggestions([])

        const autocompletionReq: GoogleMapsAutocompletionRequest = { ...autocompletionRequest }

        setIsFetching(true)
        autocompleteService.getPlacePredictions(
            requestBuilder(
                autocompletionReq,
                value,
                withSessionToken && sessionToken,
            ), (suggestions) => {
                setIsFetching(false)
                setSuggestions((suggestions || []).map(suggestion => ({
                    label: suggestion.description,
                    value: suggestion.place_id,
                })))
            },
        )
    }, debounce)

    return {
        suggestions,
        fetchSuggestions,
        isFetching,
    }

}

const requestBuilder = (
    autocompletionRequest: GoogleMapsAutocompletionRequest,
    input: string,
    sessionToken?: google.maps.places.AutocompleteSessionToken,
): google.maps.places.AutocompletionRequest => {
    const { location, ...rest } = autocompletionRequest

    const res: google.maps.places.AutocompletionRequest = {
        input,
        ...rest,
    }

    if (sessionToken) {
        res.sessionToken = sessionToken
    }

    if (location) {
        res.location = new google.maps.LatLng(location.toJSON())
    }

    return res
}
