import { Models_HomeItem, Nullish } from "@/api/generated/types"
import { ADVANCED_SEARCH_COUNTRIES_MANGA, ADVANCED_SEARCH_MEDIA_GENRES } from "@/app/(main)/search/_lib/advanced-search-constants"

export const MAX_HOME_ITEMS = 10

export const DEFAULT_HOME_ITEMS: Models_HomeItem[] = [
    {
        id: "anime-continue-watching",
        type: "anime-continue-watching",
        schemaVersion: 1,
    },
    {
        id: "anime-library",
        type: "anime-library",
        schemaVersion: 1,
        options: {
            statuses: ["CURRENT", "PAUSED", "PLANNING", "COMPLETED", "DROPPED"],
            layout: "grid",
        },
    },
]

export function isAnimeLibraryItemsOnly(items: Nullish<Models_HomeItem[]>) {
    if (!items) return true

    for (const item of items) {
        if (![
            "anime-continue-watching",
            "anime-library",
            "anime-continue-watching-header",
            "local-anime-library",
            "local-anime-library-stats",
            "library-upcoming-episodes",
        ].includes(item.type)) {
            return false
        }
    }
    return true
}

type HomeItemSchema = {
    name: string
    kind: ("row" | "header")[]
    options?: { label: string, name: string, type: string, options?: any[] }[]
    schemaVersion: number
    description?: string
}

const _carouselOptions = [
    {
        label: "Name",
        type: "text",
        name: "name",
    },
    {
        label: "Sorting",
        type: "select",
        name: "sorting",
        options: [
            {
                label: "Popular",
                value: "POPULARITY_DESC",
            },
            {
                label: "Trending",
                value: "TRENDING_DESC",
            },
            {
                label: "Romaji Title (A-Z)",
                value: "TITLE_ROMAJI_ASC",
            },
            {
                label: "Romaji Title (Z-A)",
                value: "TITLE_ROMAJI_DESC",
            },
            {
                label: "English title (A-Z)",
                value: "TITLE_ENGLISH_ASC",
            },
            {
                label: "English title (Z-A)",
                value: "TITLE_ENGLISH_DESC",
            },
            {
                label: "Score (0-10)",
                value: "SCORE",
            },
            {
                label: "Score (10-0)",
                value: "SCORE_DESC",
            },
        ],
    },
    {
        label: "Status",
        type: "multi-select",
        name: "status",
        options: [
            {
                label: "Releasing",
                value: "RELEASING",
            },
            {
                label: "Finished",
                value: "FINISHED",
            },
            {
                label: "Not yet released",
                value: "NOT_YET_RELEASED",
            },
        ],
    },
    {
        label: "Format",
        type: "select",
        name: "format",
        options: [
            {
                label: "TV",
                value: "TV",
            },
            {
                label: "Movie",
                value: "MOVIE",
            },
            {
                label: "OVA",
                value: "OVA",
            },
            {
                label: "ONA",
                value: "ONA",
            },
            {
                label: "Special",
                value: "SPECIAL",
            },
        ],
    },
    {
        label: "Genres",
        type: "multi-select",
        options: ADVANCED_SEARCH_MEDIA_GENRES.map(n => ({ value: n, label: n })),
        name: "genres",
    },
    {
        label: "Season",
        type: "select",
        name: "season",
        options: [
            { value: "WINTER", label: "Winter" },
            { value: "SPRING", label: "Spring" },
            { value: "SUMMER", label: "Summer" },
            { value: "FALL", label: "Fall" },
        ],
    },
    {
        label: "Year",
        type: "number",
        name: "year",
        min: 0,
        max: 2100,
    },
    {
        label: "Country of Origin",
        type: "select",
        name: "countryOfOrigin",
        options: ADVANCED_SEARCH_COUNTRIES_MANGA,
    },
]

export const HOME_ITEMS = {
    "centered-title": {
        name: "Centered title",
        kind: ["row"],
        schemaVersion: 1,
        description: "Display a centered title text.",
        options: [{
            label: "Text",
            type: "text",
            name: "text",
        }],
    },
    "anime-continue-watching": {
        name: "Continue Watching",
        kind: ["row", "header"],
        schemaVersion: 1,
        description: "Display a list of episodes you are currently watching.",
    },
    "anime-continue-watching-header": {
        name: "Continue Watching Header",
        kind: ["header"],
        schemaVersion: 1,
        description: "Display a header with a carousel of anime you are currently watching.",
    },
    "anime-library": {
        name: "Anime Library",
        kind: ["row"],
        schemaVersion: 2,
        description: "Display anime you have in your library by status.",
        options: [
            {
                label: "Statuses",
                name: "statuses",
                type: "multi-select",
                options: [
                    {
                        value: "CURRENT",
                        label: "Currently Watching",
                    },
                    {
                        value: "PAUSED",
                        label: "Paused",
                    },
                    {
                        value: "PLANNING",
                        label: "Planning",
                    },
                    {
                        value: "COMPLETED",
                        label: "Completed",
                    },
                    {
                        value: "DROPPED",
                        label: "Dropped",
                    },
                ],
            },
            {
                label: "Layout",
                name: "layout",
                type: "select",
                options: [
                    {
                        label: "Grid",
                        value: "grid",
                    },
                    {
                        label: "Carousel",
                        value: "carousel",
                    },
                ],
            },
        ],
    },
    "local-anime-library": {
        name: "Local Anime Library",
        kind: ["row"],
        schemaVersion: 2,
        description: "Display a complete grid of anime you have in your local library.",
        options: [
            {
                label: "Layout",
                name: "layout",
                type: "select",
                options: [
                    {
                        label: "Grid",
                        value: "grid",
                    },
                    {
                        label: "Carousel",
                        value: "carousel",
                    },
                ],
            },
        ],
    },
    "library-upcoming-episodes": {
        name: "Upcoming Library Episodes",
        kind: ["row"],
        schemaVersion: 1,
        description: "Display a carousel of upcoming episodes from anime you have in your library.",
    },
    "aired-recently": {
        name: "Aired Recently (Global)",
        kind: ["row"],
        schemaVersion: 1,
        description: "Display a carousel of anime episodes that aired recently.",
    },
    "anime-schedule-calendar": {
        name: "Anime Schedule Calendar",
        kind: ["row"],
        schemaVersion: 2,
        description: "Display a calendar of anime episodes based on their airing schedule.",
        options: [
            {
                label: "Type",
                name: "type",
                type: "select",
                options: [
                    {
                        label: "My lists",
                        value: "my-lists",
                    },
                    {
                        label: "Global",
                        value: "global",
                    },
                ],
            },
        ],
    },
    "local-anime-library-stats": {
        name: "Local Anime Library Stats",
        kind: ["row"],
        schemaVersion: 1,
        description: "Display the stats for your local anime library.",
    },
    "discover-header": {
        name: "Discover Header",
        kind: ["header"],
        schemaVersion: 1,
        description: "Display a header with a carousel of anime that are trending.",
    },
    "anime-carousel": {
        name: "Anime Carousel",
        kind: ["row"],
        schemaVersion: 3,
        options: _carouselOptions,
        description: "Display a carousel of anime based on the selected options.",
    },
    "manga-carousel": {
        name: "Manga Carousel",
        kind: ["row"],
        schemaVersion: 1,
        description: "Display a carousel of manga based on the selected options.",
        options: _carouselOptions.map(n => {
            if (n.name === "format") {
                return {
                    ...n,
                    options: [
                        {
                            label: "Manga",
                            value: "MANGA",
                        },
                        {
                            label: "One Shot",
                            value: "ONE_SHOT",
                        },
                    ],
                }
            }
            return n
        }),
    },
    "manga-library": {
        name: "Manga Library",
        kind: ["row", "header"],
        schemaVersion: 2,
        description: "Display a list of manga you have in your library by status.",
        options: [
            {
                label: "Statuses",
                name: "statuses",
                type: "multi-select",
                options: [
                    {
                        value: "CURRENT",
                        label: "Currently Reading",
                    },
                    {
                        value: "PAUSED",
                        label: "Paused",
                    },
                ],
            },
            {
                label: "Layout",
                name: "layout",
                type: "select",
                options: [
                    {
                        label: "Grid",
                        value: "grid",
                    },
                    {
                        label: "Carousel",
                        value: "carousel",
                    },
                ],
            },
        ],
    },
} as Record<string, HomeItemSchema>

export const HOME_ITEM_IDS = Object.keys(HOME_ITEMS) as (keyof typeof HOME_ITEMS)[]

// export type HomeItemID = (keyof typeof HOME_ITEMS)
