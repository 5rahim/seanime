import { CustomBackgroundImage } from "@/app/(main)/_features/custom-ui/custom-background-image"
import Template from "@/app/template"
import { AppErrorBoundary } from "@/components/shared/app-error-boundary"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { NotFound } from "@/components/shared/not-found"
import { createRootRoute, Outlet } from "@tanstack/react-router"
import React from "react"

export const Route = createRootRoute({
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
