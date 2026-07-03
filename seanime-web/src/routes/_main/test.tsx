import TestPage from "@/app/(main)/test/page"
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_main/test")({
    component: TestPage,
})
