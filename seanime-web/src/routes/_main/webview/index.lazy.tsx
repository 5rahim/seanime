import Page from "@/app/(main)/webview/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/webview/")({
    component: Page,
})
