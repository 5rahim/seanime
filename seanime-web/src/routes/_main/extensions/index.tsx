import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/extensions/page"

const searchSchema = z.object({
    tab: z.enum(["installed", "marketplace"]).optional(),
})

export const Route = createFileRoute("/_main/extensions/")({
    component: Page,
    validateSearch: searchSchema,
})
