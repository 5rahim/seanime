import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/mediastream/page"

const searchSchema = z.object({
    id: z.coerce.number().optional(),
})

export const Route = createFileRoute("/_main/mediastream/")({
    component: Page,
    validateSearch: searchSchema,
})
