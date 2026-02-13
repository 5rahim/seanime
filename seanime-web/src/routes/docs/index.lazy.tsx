import Page from "@/app/docs/page"

import { SimpleAuthWrapper } from "@/components/shared/simple-auth-wrapper"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/docs/")({
    component: () => <SimpleAuthWrapper><Page /></SimpleAuthWrapper>,
})
