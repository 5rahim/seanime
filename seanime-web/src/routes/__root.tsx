import { CustomBackgroundImage } from "@/app/(main)/_features/custom-ui/custom-background-image.tsx"
import { createRootRoute, Outlet } from "@tanstack/react-router"
import Template from "@/app/template"
import React from "react"
import { TauriManager } from "@/app/(main)/_tauri/tauri-manager"
import { ElectronManager } from "@/app/(main)/_electron/electron-manager"
import { __isElectronDesktop__, __isTauriDesktop__ } from "@/types/constants"
import { AppErrorBoundary } from "@/components/shared/app-error-boundary"
import { NotFound } from "@/components/shared/not-found"

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
    errorComponent: AppErrorBoundary,
    notFoundComponent: NotFound,
})
