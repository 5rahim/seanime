import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/(offline)/offline/entry/anime/page"

const searchSchema = z.object({
    id: z.coerce.number().optional(),
})

export const Route = createFileRoute("/_main/offline/entry/anime/")({
    component: Page,
    validateSearch: searchSchema,
})
