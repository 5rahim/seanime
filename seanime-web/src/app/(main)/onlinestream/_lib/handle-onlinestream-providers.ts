import { useListOnlinestreamProviderExtensions } from "@/api/hooks/extensions.hooks"
import { __onlinestream_selectedProviderAtom } from "@/app/(main)/onlinestream/_lib/onlinestream.atoms"
import { logger } from "@/lib/helpers/debug"
import { useAtom } from "jotai/react"
import React from "react"

export function useHandleOnlinestreamProviderExtensions() {

    const { data: providerExtensions } = useListOnlinestreamProviderExtensions()

    const [provider, setProvider] = useAtom(__onlinestream_selectedProviderAtom)

    /**
     * Override the selected provider if it is not available
     */
    React.useLayoutEffect(() => {
        logger("ONLINESTREAM").info("extensions", providerExtensions)

        if (!providerExtensions) return

        if (provider === null || !providerExtensions.find(p => p.id === provider)) {
            if (providerExtensions.length > 0) {
                setProvider(providerExtensions[0].id)
            } else {
                setProvider(null)
            }
        }
    }, [providerExtensions])

    return {
        providerExtensions: providerExtensions ?? [],
        providerExtensionOptions: (providerExtensions ?? []).map(provider => ({
            label: provider.name,
            value: provider.id,
        })).sort((a, b) => a.label.localeCompare(b.label)),
    }

}
