import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"

const searchSchema = z.object({
    tab: z.enum(["installed", "marketplace"]).optional(),
})

export const Route = createFileRoute("/_main/extensions/")({
    validateSearch: searchSchema,
})
