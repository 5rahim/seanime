import { ElectronManager } from "@/app/(main)/_electron/electron-manager"
import { CustomBackgroundImage } from "@/app/(main)/_features/custom-ui/custom-background-image.tsx"
import { TauriManager } from "@/app/(main)/_tauri/tauri-manager"
import Template from "@/app/template"
import { AppErrorBoundary } from "@/components/shared/app-error-boundary"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo.tsx"
import { NotFound } from "@/components/shared/not-found"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { createRootRoute, Outlet } from "@tanstack/react-router"
import React from "react"

export const Route = createRootRoute({
    component: () => (
        <Template>
            {__isTauriDesktop__ && <TauriManager />}
            {__isElectronDesktop__ && <ElectronManager />}
            <CustomBackgroundImage />
            <Outlet />
            {/*<TanStackRouterDevtools /> */}
        </Template>
    ),
    pendingComponent: LoadingOverlayWithLogo,
    pendingMs: 200,
    errorComponent: AppErrorBoundary,
    notFoundComponent: NotFound,
})
