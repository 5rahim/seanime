import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

const searchSchema = z.object({
    provider: z.string().optional(),
})

export const Route = createFileRoute("/_main/custom-sources/")({
    validateSearch: searchSchema,
})
