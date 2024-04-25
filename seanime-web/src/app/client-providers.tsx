"use client"
import { WebsocketProvider } from "@/app/websocket-provider"
import { Toaster } from "@/components/ui/toaster"
import { QueryClient } from "@tanstack/query-core"
import { QueryClientProvider } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { createStore } from "jotai"
import { Provider as JotaiProvider } from "jotai/react"
import { ThemeProvider } from "next-themes"
import { usePathname } from "next/navigation"
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
    const pathname = usePathname()

    return (
        <ThemeProvider attribute="class" defaultTheme="dark" forcedTheme={pathname === "/docs" ? "light" : "dark"}>
            <JotaiProvider store={store}>
                <QueryClientProvider client={queryClient}>
                    <WebsocketProvider>
                        {children}
                        <Toaster />
                    </WebsocketProvider>
                    {process.env.NODE_ENV === "development" && <React.Suspense fallback={null}>
                        <ReactQueryDevtools />
                    </React.Suspense>}
                </QueryClientProvider>
            </JotaiProvider>
        </ThemeProvider>
    )

}
