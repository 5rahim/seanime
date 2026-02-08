import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

const searchSchema = z.object({
    sorting: z.string().optional(),
    genre: z.string().optional(),
    status: z.string().optional(),
    format: z.string().optional(),
    season: z.string().optional(),
    year: z.coerce.number().optional(),
    type: z.string().optional(),
})

export const Route = createFileRoute("/_main/search/")({
    validateSearch: searchSchema,
})
