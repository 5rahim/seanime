import Page from "@/app/issue-report/page"

import { SimpleAuthWrapper } from "@/components/shared/simple-auth-wrapper"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/issue-report/")({
    component: () => <SimpleAuthWrapper><Page /></SimpleAuthWrapper>,
})
