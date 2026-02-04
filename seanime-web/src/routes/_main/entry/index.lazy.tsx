import { AnimeEntryPage } from "@/app/(main)/entry/_containers/anime-entry-page"
import { createLazyFileRoute } from "@tanstack/react-router"

export const Route = createLazyFileRoute("/_main/entry/")({
    component: AnimeEntryPage,
})
