import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/custom-sources/page"

const searchSchema = z.object({
    provider: z.string().optional(),
})

export const Route = createFileRoute("/_main/custom-sources/")({
    component: Page,
    validateSearch: searchSchema,
})
