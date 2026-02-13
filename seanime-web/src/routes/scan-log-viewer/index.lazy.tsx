import Page from "@/app/scan-log-viewer/page"

import { SimpleAuthWrapper } from "@/components/shared/simple-auth-wrapper"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/scan-log-viewer/")({
    component: () => <SimpleAuthWrapper><Page /></SimpleAuthWrapper>,
})
