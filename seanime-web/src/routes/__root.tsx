import { CustomBackgroundImage } from "@/app/(main)/_features/custom-ui/custom-background-image"
import Template from "@/app/template"
import { AppErrorBoundary } from "@/components/shared/app-error-boundary"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { NotFound } from "@/components/shared/not-found"

import { QueryClient } from "@tanstack/react-query"
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router"

import { createStore } from "jotai"
import React from "react"

export const Route = createRootRouteWithContext<{
    queryClient: QueryClient
    store: ReturnType<typeof createStore>
}>()({
    component: () => (
        <Template>
            <CustomBackgroundImage />
            <Outlet />
            {/*<TanStackRouterDevtools />*/}
        </Template>
    ),
    pendingComponent: LoadingOverlayWithLogo,
    pendingMs: 200,
    errorComponent: AppErrorBoundary,
    notFoundComponent: NotFound,
})
