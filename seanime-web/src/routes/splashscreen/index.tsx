import { createFileRoute } from "@tanstack/react-router"
import Page from "@/app/splashscreen/page"

export const Route = createFileRoute("/splashscreen/")({
    component: Page,
})
