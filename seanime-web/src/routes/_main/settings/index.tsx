import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/settings/page"

const searchSchema = z.object({
    tab: z.string().optional(),
})

export const Route = createFileRoute("/_main/settings/")({
    component: Page,
    validateSearch: searchSchema,
})
