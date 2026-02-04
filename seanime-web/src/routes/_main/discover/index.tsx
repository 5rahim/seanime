import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

const searchSchema = z.object({
    type: z.enum(["anime", "schedule", "manga"]).optional(),
})

export const Route = createFileRoute("/_main/discover/")({
    validateSearch: searchSchema,
})
