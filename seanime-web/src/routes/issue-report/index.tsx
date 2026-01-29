import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/issue-report/page"

export const Route = createFileRoute("/issue-report/")({
    component: Page,
})
