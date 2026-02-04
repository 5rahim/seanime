import Page from "@/app/issue-report/page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/issue-report/")({
    component: Page,
})
