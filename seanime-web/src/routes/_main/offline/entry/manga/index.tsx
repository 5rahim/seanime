import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/(offline)/offline/entry/manga/page"

const searchSchema = z.object({
    id: z.coerce.number().optional(),
})

export const Route = createFileRoute("/_main/offline/entry/manga/")({
    component: Page,
    validateSearch: searchSchema,
})
