import { useListOnlinestreamProviderExtensions } from "@/api/hooks/extensions.hooks"
import { __onlinestream_selectedProviderAtom } from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { useAtom } from "jotai/react"
import React from "react"

export function useHandleOnlinestreamProviders() {

    const { data: providerExtensions } = useListOnlinestreamProviderExtensions()

    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)

    /**
     * Override the selected provider if it is not available
     */
    React.useLayoutEffect(() => {
        if (!!providerExtensions && !providerExtensions.find(p => p.id === provider)) {
            if (providerExtensions.length > 0) {
                setProvider(providerExtensions[0].id)
            } else {
                setProvider(null)
            }
        }
    }, [providerExtensions])

    return {
        providers: providerExtensions ?? [],
        providerOptions: (providerExtensions ?? []).map(provider => ({
            label: provider.name,
            value: provider.id,
        })).sort((a, b) => a.label.localeCompare(b.label)),
    }

}
