import { AL_MediaFormat } from "@/api/generated/types"

export const ADVANCED_SEARCH_MEDIA_GENRES = [
    "Action",
    "Adventure",
    "Comedy",
    "Drama",
    "Ecchi",
    "Fantasy",
    "Horror",
    "Mahou Shoujo",
    "Mecha",
    "Music",
    "Mystery",
    "Psychological",
    "Romance",
    "Sci-Fi",
    "Slice of Life",
    "Sports",
    "Supernatural",
    "Thriller",
]

export const ADVANCED_SEARCH_SEASONS = [
    "Winter",
    "Spring",
    "Summer",
    "Fall",
]

export const ADVANCED_SEARCH_FORMATS: { value: AL_MediaFormat, label: string }[] = [
    { value: "TV", label: "TV" },
    { value: "MOVIE", label: "Movie" },
    { value: "ONA", label: "ONA" },
    { value: "OVA", label: "OVA" },
    { value: "TV_SHORT", label: "TV Short" },
    { value: "SPECIAL", label: "Special" },
]

export const ADVANCED_SEARCH_FORMATS_MANGA: { value: AL_MediaFormat, label: string }[] = [
    { value: "MANGA", label: "Manga" },
    { value: "ONE_SHOT", label: "One Shot" },
]


export const ADVANCED_SEARCH_COUNTRIES_MANGA: { value: string, label: string }[] = [
    { value: "JP", label: "Japan" },
    { value: "KR", label: "South Korea" },
    { value: "CN", label: "China" },
    { value: "TW", label: "Taiwan" },
]

export const ADVANCED_SEARCH_STATUS = [
    { value: "FINISHED", label: "Finished" },
    { value: "RELEASING", label: "Releasing" },
    { value: "NOT_YET_RELEASED", label: "Upcoming" },
    { value: "HIATUS", label: "Hiatus" },
    { value: "CANCELLED", label: "Cancelled" },
]

export const ADVANCED_SEARCH_SORTING = [
    { value: "TRENDING_DESC", label: "Trending" },
    { value: "START_DATE_DESC", label: "Release date" },
    { value: "SCORE_DESC", label: "Highest score" },
    { value: "POPULARITY_DESC", label: "Most popular" },
    { value: "EPISODES_DESC", label: "Number of episodes" },
]

export const ADVANCED_SEARCH_SORTING_MANGA = [
    { value: "TRENDING_DESC", label: "Trending" },
    { value: "START_DATE_DESC", label: "Release date" },
    { value: "SCORE_DESC", label: "Highest score" },
    { value: "POPULARITY_DESC", label: "Most popular" },
    { value: "CHAPTERS_DESC", label: "Number of chapters" },
]

export const ADVANCED_SEARCH_TYPE = [
    { value: "anime", label: "Anime" },
    { value: "manga", label: "Manga" },
]
