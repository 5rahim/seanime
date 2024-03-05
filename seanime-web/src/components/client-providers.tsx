"use client"
import { AuthWrapper } from "@/components/application/auth-wrapper"
import { WebsocketProvider } from "@/components/application/websocket-provider"
import { Toaster } from "@/components/ui/toaster"
import { QueryClient } from "@tanstack/query-core"
import { QueryClientProvider } from "@tanstack/react-query"
import { createStore } from "jotai"
import { Provider as JotaiProvider } from "jotai/react"
import { ThemeProvider } from "next-themes"
import React from "react"

interface ClientProvidersProps {
    children?: React.ReactNode
}

const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            refetchOnWindowFocus: false,
            retry: 0,
        },
    },
})

export const ClientProviders: React.FC<ClientProvidersProps> = ({ children }) => {
    const [store] = React.useState(createStore())

    return (
        <ThemeProvider attribute="class" defaultTheme="dark" forcedTheme="dark">
            <JotaiProvider store={store}>
                <QueryClientProvider client={queryClient}>
                    <WebsocketProvider>
                        <AuthWrapper>
                            {children}
                            <Toaster />
                        </AuthWrapper>
                    </WebsocketProvider>
                </QueryClientProvider>
            </JotaiProvider>
        </ThemeProvider>
    )

}
