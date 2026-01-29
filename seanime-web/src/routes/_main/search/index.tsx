import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/search/page"

const searchSchema = z.object({
    sorting: z.string().optional(),
    genre: z.string().optional(),
    status: z.string().optional(),
    format: z.string().optional(),
    season: z.string().optional(),
    year: z.string().optional(),
    type: z.string().optional(),
})

export const Route = createFileRoute("/_main/search/")({
    component: Page,
    validateSearch: searchSchema,
})
