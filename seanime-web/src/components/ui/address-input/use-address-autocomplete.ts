"use client"

import { Loader, LoaderOptions } from "@googlemaps/js-api-loader"
import { useCallback, useEffect, useState } from "react"
import { useDebouncedCallback } from "use-debounce"

export type GoogleMapsAutocompletionRequest = Omit<google.maps.places.AutocompletionRequest, "input">

export interface GoogleMapsAutocompleteProps {
    apiKey: string,
    minLengthAutocomplete?: number,
    withSessionToken?: boolean
    debounce?: number
    autocompletionRequest?: GoogleMapsAutocompletionRequest
    loaderApiOptions?: Partial<LoaderOptions>
}

export const useGoogleMapsAutocomplete = ({
                                              apiKey,
                                              minLengthAutocomplete = 0,
                                              withSessionToken = false,
                                              debounce = 300,
                                              autocompletionRequest,
                                              loaderApiOptions = {},
                                          }:
                                              GoogleMapsAutocompleteProps) => {

    const [autocompleteService, setAutocompleteService] = useState<google.maps.places.AutocompleteService | undefined>(undefined)
    const [sessionToken, setSessionToken] = useState<google.maps.places.AutocompleteSessionToken | undefined>(undefined)

    const initializeService = useCallback(() => {
        if (!window.google) throw new Error("[AddressInput]: Google script not loaded")
        if (!window.google.maps) throw new Error("[AddressInput]: Google maps script not loaded")
        if (!window.google.maps.places) throw new Error("[AddressInput]: Google maps places script not loaded")

        setAutocompleteService(new window.google.maps.places.AutocompleteService())
        setSessionToken(new google.maps.places.AutocompleteSessionToken())
    }, [window])

    /**
     * Initialize
     */
    useEffect(() => {
        const init = async () => {
            try {
                if (!window.google || !window.google.maps || !window.google.maps.places) {
                    await new Loader({ apiKey: apiKey, ...{ libraries: ["places"], ...loaderApiOptions } }).load()
                }
                initializeService()
            } catch (error) {
                console.log(error)
            }
        }

        if (apiKey) init()
        else initializeService()
    }, [])


    /**
     * Fetch suggestions
     */
    const [suggestions, setSuggestions] = useState<{ label: string, value: string }[]>([])

    const fetchSuggestions = useDebouncedCallback((value: string): void => {
        if (!autocompleteService) return setSuggestions([])
        if (value.length < minLengthAutocomplete) return setSuggestions([])

        const autocompletionReq: GoogleMapsAutocompletionRequest = { ...autocompletionRequest }

        autocompleteService.getPlacePredictions(
            requestBuilder(
                autocompletionReq,
                value,
                withSessionToken && sessionToken,
            ), (suggestions) => {
                setSuggestions((suggestions || []).map(suggestion => ({
                    label: suggestion.description,
                    value: suggestion.place_id
                })))
            },
        )
    }, debounce)

    return {
        suggestions,
        fetchSuggestions,
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
        res.location = new google.maps.LatLng(location)
    }

    return res
}
