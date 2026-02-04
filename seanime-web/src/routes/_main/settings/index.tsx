import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

const searchSchema = z.object({
    tab: z.string().optional(),
})

export const Route = createFileRoute("/_main/settings/")({
    validateSearch: searchSchema,
})
