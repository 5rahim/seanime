import { createFileRoute } from "@tanstack/react-router"
import { AnimeEntryPage } from "@/app/(main)/entry/_containers/anime-entry-page"
import { z } from "zod"

const searchSchema = z.object({
    id: z.coerce.number().optional(),
})

export const Route = createFileRoute("/_main/entry/")({
    component: AnimeEntryPage,
    validateSearch: searchSchema,
})
