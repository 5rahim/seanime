import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/splashscreen/crash/page"

export const Route = createFileRoute("/splashscreen/crash/")({
    component: Page,
})
