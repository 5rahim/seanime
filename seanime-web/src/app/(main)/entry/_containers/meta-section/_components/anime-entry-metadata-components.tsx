import { getMediaDetailsStats } from "@/app/(main)/entry/_containers/meta-section/helpers"
import { Badge } from "@/components/ui/badge"
import { IconButton } from "@/components/ui/button"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { Tooltip } from "@/components/ui/tooltip"
import { MediaDetailsByIdQuery } from "@/lib/anilist/gql/graphql"
import capitalize from "lodash/capitalize"
import React from "react"
import { AiFillStar, AiOutlineHeart, AiOutlineStar } from "react-icons/ai"
import { BiHeart, BiHide } from "react-icons/bi"

type AnimeEntryGenresProps = {
    genres?: Array<string | null> | null | undefined
}

export function AnimeEntryGenres(props: AnimeEntryGenresProps) {

    const {
        genres,
        ...rest
    } = props

    if (!genres) return null
    return (
        <>
            <div className="items-center flex flex-wrap gap-2">
                {genres?.map(genre => {
                    return <Badge key={genre!} className="border-transparent" size="lg">{genre}</Badge>
                })}
            </div>
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AnimeEntryAudienceScoreProps = {
    meanScore?: number | null
    hideAudienceScore?: boolean
}

export function AnimeEntryAudienceScore(props: AnimeEntryAudienceScoreProps) {

    const {
        meanScore,
        hideAudienceScore,
        ...rest
    } = props

    if (!meanScore) return null

    const ScoreBadge = (
        <Badge
            className=""
            size="lg"
            intent={meanScore >= 70 ? meanScore >= 82 ? "primary" : "success" : "gray"}
            leftIcon={<BiHeart />}
        >{meanScore / 10}</Badge>
    )

    return (
        <>
            {hideAudienceScore ? <Disclosure type="single" collapsible>
                <DisclosureItem value="item-1" className="flex items-center gap-1">
                    <Tooltip
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
                        {ScoreBadge}
                    </DisclosureContent>
                </DisclosureItem>
            </Disclosure> : ScoreBadge}
        </>
    )
}


//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type AnimeEntryStudioProps = {
    details?: MediaDetailsByIdQuery["Media"]
}

export function AnimeEntryStudio(props: AnimeEntryStudioProps) {

    const {
        details,
        ...rest
    } = props

    if (!details?.studios?.nodes) return null

    return (
        <>
            <Badge
                size="lg"
                intent="gray"
                className="rounded-full border-transparent"
            >
                {details.studios?.nodes?.[0]?.name}
            </Badge>
        </>
    )
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////


type AnimeEntryRanksProps = {
    details?: MediaDetailsByIdQuery["Media"]
}

export function AnimeEntryRanks(props: AnimeEntryRanksProps) {

    const {
        details,
        ...rest
    } = props

    const {
        seasonHighestRated,
        seasonMostPopular,
        allTimeHighestRated,
    } = getMediaDetailsStats(details)

    const formatFormat = React.useCallback((format: string) => {
        return (format === "TV" ? "" : format).replace("_", " ")
    }, [])

    if (!details) return null

    return (
        <>
            {(!!allTimeHighestRated || !!seasonMostPopular) && <div className="flex-wrap gap-2 hidden md:flex">
                {allTimeHighestRated && <Badge
                    size="lg"
                    intent="gray"
                    leftIcon={<AiFillStar />}
                    iconClass="text-yellow-500"
                    className="rounded-md border-transparent px-2"
                >
                    #{String(allTimeHighestRated.rank)} Highest
                    Rated {formatFormat(allTimeHighestRated.format)} of All
                    Time
                </Badge>}
                {seasonHighestRated && <Badge
                    size="lg"
                    intent="gray"
                    leftIcon={<AiOutlineStar />}
                    iconClass="text-yellow-500"
                    className="rounded-md border-transparent px-2"
                >
                    #{String(seasonHighestRated.rank)} Highest
                    Rated {formatFormat(seasonHighestRated.format)} of {capitalize(seasonHighestRated.season!)} {seasonHighestRated.year}
                </Badge>}
                {seasonMostPopular && <Badge
                    size="lg"
                    intent="gray"
                    leftIcon={<AiOutlineHeart />}
                    iconClass="text-pink-500"
                    className="rounded-md border-transparent px-2"
                >
                    #{(String(seasonMostPopular.rank))} Most
                    Popular {formatFormat(seasonMostPopular.format)} of {capitalize(seasonMostPopular.season!)} {seasonMostPopular.year}
                </Badge>}
            </div>}
        </>
    )
}
