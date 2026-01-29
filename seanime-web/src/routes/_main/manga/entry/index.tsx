import { createFileRoute } from "@tanstack/react-router"
import { z } from "zod"
import Page from "@/app/(main)/manga/entry/page"

const searchSchema = z.object({
    id: z.coerce.number().optional(),
})

export const Route = createFileRoute("/_main/manga/entry/")({
    component: Page,
    validateSearch: searchSchema,
})
