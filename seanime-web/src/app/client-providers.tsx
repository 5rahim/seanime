
import { WebsocketProvider } from "@/app/websocket-provider"
import { CustomCSSProvider } from "@/components/shared/custom-css-provider"
import { CustomThemeProvider } from "@/components/shared/custom-theme-provider"
import { Toaster } from "@/components/ui/toaster"
import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createStore } from "jotai"
import { Provider as JotaiProvider } from "jotai/react"
import { ThemeProvider } from "next-themes"
import React from "react"
import { CookiesProvider } from "react-cookie"

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
        <ThemeProvider attribute="class" defaultTheme="dark" forcedTheme={"dark"}>
            <CookiesProvider>
                <JotaiProvider store={store}>
                    <QueryClientProvider client={queryClient}>
                        <WebsocketProvider>
                            {children}
                            <CustomThemeProvider />
                            <Toaster />
                        </WebsocketProvider>
                        <CustomCSSProvider />
                        {/*{import.meta.env.MODE === "development" && <React.Suspense fallback={null}>*/}
                        {/*    <ReactQueryDevtools />*/}
                        {/*</React.Suspense>}*/}
                    </QueryClientProvider>
                </JotaiProvider>
            </CookiesProvider>
        </ThemeProvider>
    )

}
