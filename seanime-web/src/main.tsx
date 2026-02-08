import { ClientProviders, queryClient, store } from "@/app/client-providers"
import "./app/globals.css"
import { createRouter, RouterProvider } from "@tanstack/react-router"
import React from "react"
import ReactDOM from "react-dom/client"
import { routeTree } from "./routeTree.gen"
import "@fontsource-variable/inter"

const router = createRouter({
    routeTree,
    // defaultPreload: import.meta.env.PROD ? "intent" : false,
    defaultPreload: "intent",
    context: {
        queryClient,
        store,
    },
    scrollRestoration: true,
    defaultPreloadStaleTime: 0,
})

declare module "@tanstack/react-router" {
    interface Register {
        router: typeof router
    }
}

if (import.meta.env.DEV) {
    const script = document.createElement("script")
    script.src = "https://unpkg.com/react-scan/dist/auto.global.js"
    script.crossOrigin = "anonymous"
    document.head.appendChild(script)
}

ReactDOM.createRoot(document.getElementById("root")!).render(
    <ClientProviders>
        <RouterProvider router={router} />
    </ClientProviders>,
)
