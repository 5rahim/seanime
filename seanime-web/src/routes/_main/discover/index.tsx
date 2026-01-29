import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/discover/page"

const searchSchema = z.object({
    type: z.enum(["anime", "schedule", "manga"]).optional(),
})

export const Route = createFileRoute("/_main/discover/")({
    component: Page,
    validateSearch: searchSchema,
})
