import React from "react"
import ReactDOM from "react-dom/client"
import { RouterProvider, createRouter } from "@tanstack/react-router"
import { routeTree } from "./routeTree.gen"
import { ClientProviders } from "@/app/client-providers"
import "./app/globals.css"
import "@fontsource/inter"

const router = createRouter({ routeTree })

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
