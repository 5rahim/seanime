import { AL_AnimeDetailsById_Media_Rankings, AL_MangaDetailsById_Media_Rankings } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { SeaLink } from "@/components/shared/sea-link"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { Tooltip } from "@/components/ui/tooltip"
import { getScoreColor } from "@/lib/helpers/score"
import capitalize from "lodash/capitalize"
import React from "react"
import { AiFillStar, AiOutlineHeart, AiOutlineStar } from "react-icons/ai"
import { BiHeart, BiHide } from "react-icons/bi"

type MediaEntryGenresListProps = {
    genres?: Array<string | null> | null | undefined
    className?: string
}

export function MediaEntryGenresList(props: MediaEntryGenresListProps) {

    const {
        genres,
        className,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    if (!genres) return null

    if (serverStatus?.isOffline) {
        return (
            <>
                <div className={cn("items-center flex flex-wrap gap-0", className)}>
                    {genres?.map(genre => {
                        return <Badge
                            key={genre!}
                            className={cn(
                                "opacity-75 hover:opacity-100 transition-all border-transparent bg-transparent hover:bg-black/30 hover:text-white")}
                            size="lg"
                        >
                            {genre}
                        </Badge>
                    })}
                </div>
            </>
        )
    } else {
        return (
            <>
                <div className={cn("items-center flex flex-wrap gap-0", className)}>
                    {genres?.map(genre => {
                        return <SeaLink href={`/search?genre=${genre}&sorting=TRENDING_DESC`} key={genre!}>
                            <Badge
                                className={cn(
                                    "opacity-75 hover:opacity-100 transition-all border-transparent bg-transparent hover:bg-black/30 hover:text-white")}
                                size="lg"
                            >
                                {genre}
                            </Badge>
                        </SeaLink>
                    })}
                </div>
            </>
        )
    }
}

type MediaEntryAudienceScoreProps = {
    meanScore?: number | null
    badgeClass?: string
}

export function MediaEntryAudienceScore(props: MediaEntryAudienceScoreProps) {

    const {
        meanScore,
        badgeClass,
        ...rest
    } = props

    const status = useServerStatus()
    const hideAudienceScore = React.useMemo(() => status?.settings?.anilist?.hideAudienceScore ?? false,
        [status?.settings?.anilist?.hideAudienceScore])

    if (!meanScore) return null

    return (
        <>
            {hideAudienceScore ? <Disclosure type="single" collapsible>
                <DisclosureItem value="item-1" className="flex items-center gap-0">
                    <Tooltip
                        side="right"
                        trigger={<DisclosureTrigger>
                            <IconButton
                                intent="gray-basic"
                                icon={<BiHide className="text-sm" />}
                                rounded
                                size="sm"
                            />
                        </DisclosureTrigger>}
                    >Show audience score</Tooltip>
                    <DisclosureContent>
                        <Badge
                            intent="unstyled"
                            size="lg"
                            className={cn(getScoreColor(meanScore, "audience"), badgeClass)}
                            leftIcon={<BiHeart className="text-xs" />}
                        >{meanScore / 10}</Badge>
                    </DisclosureContent>
                </DisclosureItem>
            </Disclosure> : <Badge
                intent="unstyled"
                size="lg"
                className={cn(getScoreColor(meanScore, "audience"), badgeClass)}
                leftIcon={<BiHeart className="text-xs" />}
            >{meanScore / 10}</Badge>}
        </>
    )
}

type AnimeEntryRankingsProps = {
    rankings?: AL_AnimeDetailsById_Media_Rankings[] | AL_MangaDetailsById_Media_Rankings[]
}

export function AnimeEntryRankings(props: AnimeEntryRankingsProps) {

    const {
        rankings,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    const seasonMostPopular = rankings?.find(r => (!!r?.season || !!r?.year) && r?.type === "POPULAR" && r.rank <= 10)
    const allTimeHighestRated = rankings?.find(r => !!r?.allTime && r?.type === "RATED" && r.rank <= 100)
    const seasonHighestRated = rankings?.find(r => (!!r?.season || !!r?.year) && r?.type === "RATED" && r.rank <= 5)
    const allTimeMostPopular = rankings?.find(r => !!r?.allTime && r?.type === "POPULAR" && r.rank <= 100)

    const formatFormat = React.useCallback((format: string) => {
        if (format === "MANGA") return ""
        return (format === "TV" ? "" : format).replace("_", " ")
    }, [])

    const Link = React.useCallback((props: { children: React.ReactNode, href: string }) => {
        if (serverStatus?.isOffline) {
            return <>{props.children}</>
        }

        return <SeaLink href={props.href}>{props.children}</SeaLink>
    }, [serverStatus])

    if (!rankings) return null

    return (
        <>
            {(!!allTimeHighestRated || !!seasonMostPopular) && <div className="flex-wrap gap-2 hidden md:flex">
                {allTimeHighestRated && <Link
                    href={`/search?sorting=SCORE_DESC${allTimeHighestRated.format ? `&format=${allTimeHighestRated.format}` : ""}`}
                >
                    <Badge
                        size="lg"
                        intent="gray"
                        leftIcon={<AiFillStar className="text-lg" />}
                        iconClass="text-yellow-500"
                        className="opacity-75 transition-all hover:opacity-100 rounded-full bg-transparent border-transparent px-2 hover:bg-black/30 hover:text-white"
                    >
                        #{String(allTimeHighestRated.rank)} Highest
                        Rated {formatFormat(allTimeHighestRated.format)} of All
                        Time
                    </Badge>
                </Link>}
                {seasonHighestRated && <Link
                    href={`/search?sorting=SCORE_DESC${seasonHighestRated.format
                        ? `&format=${seasonHighestRated.format}`
                        : ""}${seasonHighestRated.season ? `&season=${seasonHighestRated.season}` : ""}&year=${seasonHighestRated.year}`}
                >
                    <Badge
                        size="lg"
                        intent="gray"
                        leftIcon={<AiOutlineStar />}
                        iconClass="text-yellow-500"
                        className="opacity-75 transition-all hover:opacity-100 rounded-full border-transparent bg-transparent px-2 hover:bg-black/30 hover:text-white"
                    >
                        #{String(seasonHighestRated.rank)} Highest
                        Rated {formatFormat(seasonHighestRated.format)} of {capitalize(seasonHighestRated.season!)} {seasonHighestRated.year}
                    </Badge>
                </Link>}
                {seasonMostPopular && <Link
                    href={`/search?sorting=POPULARITY_DESC${seasonMostPopular.format
                        ? `&format=${seasonMostPopular.format}`
                        : ""}${seasonMostPopular.year ? `&year=${seasonMostPopular.year}` : ""}${seasonMostPopular.season
                        ? `&season=${seasonMostPopular.season}`
                        : ""}`}
                >
                    <Badge
                        size="lg"
                        intent="gray"
                        leftIcon={<AiOutlineHeart />}
                        iconClass="text-pink-500"
                        className="opacity-75 transition-all hover:opacity-100 rounded-full border-transparent bg-transparent px-2 hover:bg-black/30 hover:text-white"
                    >
                        #{(String(seasonMostPopular.rank))} Most
                        Popular {formatFormat(seasonMostPopular.format)} of {capitalize(seasonMostPopular.season!)} {seasonMostPopular.year}
                    </Badge>
                </Link>}
            </div>}
        </>
    )
}
