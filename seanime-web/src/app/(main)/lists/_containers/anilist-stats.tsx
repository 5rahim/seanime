import type {
    AL_Stats,
    AL_UserFormatStats,
    AL_UserGenreStats,
    AL_UserReleaseYearStats,
    AL_UserScoreStats,
    AL_UserStartYearStats,
    AL_UserStatusStats,
    AL_UserStudioStats,
} from "@/api/generated/types"
import { AreaChart, BarChart, DonutChart, Legend } from "@/components/ui/charts"
import type { ChartColor } from "@/components/ui/charts/color-theme"
import { cn } from "@/components/ui/core/styling"
import { Skeleton } from "@/components/ui/skeleton"
import { StaticTabs } from "@/components/ui/tabs"
import React from "react"
import { FiBarChart2, FiBookOpen, FiClock } from "react-icons/fi"
import { LuStar, LuTrendingUp } from "react-icons/lu"
import { PiTelevisionSimpleBold } from "react-icons/pi"

type AnilistStatsProps = {
    stats?: AL_Stats
    isLoading?: boolean
}

const formatName: Record<string, string> = {
    TV: "TV",
    TV_SHORT: "TV Short",
    MOVIE: "Movie",
    SPECIAL: "Special",
    OVA: "OVA",
    ONA: "ONA",
    MUSIC: "Music",
}

const statusName: Record<string, string> = {
    CURRENT: "Watching",
    PLANNING: "Planning",
    COMPLETED: "Completed",
    DROPPED: "Dropped",
    PAUSED: "Paused",
    REPEATING: "Repeating",
}

const statusColors: Record<string, ChartColor> = {
    CURRENT: "purple",
    COMPLETED: "green",
    PLANNING: "blue",
    PAUSED: "amber",
    DROPPED: "red",
    REPEATING: "purple",
}

const statusOrder = ["CURRENT", "COMPLETED", "PLANNING", "PAUSED", "DROPPED", "REPEATING"]

export function AnilistStats(props: AnilistStatsProps) {
    const { stats, isLoading } = props
    const [activeTab, setActiveTab] = React.useState<"anime" | "manga">("anime")

    if (isLoading) {
        return (
            <div className="space-y-6 py-4" data-anilist-stats>
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                    {Array.from({ length: 4 }).map((_, i) => (
                        <Skeleton key={i} className="h-24 rounded-lg" />
                    ))}
                </div>
                <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <Skeleton className="h-72 rounded-lg" />
                    <Skeleton className="h-72 rounded-lg" />
                </div>
            </div>
        )
    }

    return (
        <div className="space-y-8 py-4" data-anilist-stats>
            <div className="flex justify-center" data-anilist-stats-tabs>
                <StaticTabs
                    className="w-fit border rounded-full py-0"
                    triggerClass="px-6 py-2 h-full rounded-full"
                    pillClass="rounded-full border-transparent"
                    items={[
                        { name: "Anime", isCurrent: activeTab === "anime", onClick: () => setActiveTab("anime") },
                        { name: "Manga", isCurrent: activeTab === "manga", onClick: () => setActiveTab("manga") },
                    ]}
                />
            </div>

            {activeTab === "anime" && <AnimeStatsView stats={stats} />}
            {activeTab === "manga" && <MangaStatsView stats={stats} />}
        </div>
    )
}

function AnimeStatsView({ stats }: { stats?: AL_Stats }) {
    const anime = stats?.animeStats
    const thisYear = new Date().getFullYear()
    const thisYearData = anime?.startYears?.find(y => y.startYear === thisYear)
    const lastYearData = anime?.startYears?.find(y => y.startYear === thisYear - 1)

    const hoursWatched = Math.round((anime?.minutesWatched ?? 0) / 60)
    const daysWatched = hoursWatched > 0 ? `${(hoursWatched / 24).toFixed(1)}d` : "0d"

    const scoreData = React.useMemo(() => toScoreData(anime?.scores), [anime?.scores])
    const statusData = React.useMemo(() => toStatusData(anime?.statuses), [anime?.statuses])
    const formatData = React.useMemo(() => toFormatRows(anime?.formats), [anime?.formats])
    const genreData = React.useMemo(() => toRankingRows(anime?.genres, "titles"), [anime?.genres])
    const startYearData = React.useMemo(() => toStartYearData(anime?.startYears), [anime?.startYears])
    const releaseYearData = React.useMemo(() => toReleaseYearData(anime?.releaseYears), [anime?.releaseYears])
    const studioData = React.useMemo(() => toStudioRows(anime?.studios), [anime?.studios])
    const highlights = React.useMemo(() => toHighlights({
        scores: anime?.scores,
        statuses: anime?.statuses,
        genres: anime?.genres,
        startYears: anime?.startYears,
        unit: "titles",
    }), [anime?.scores, anime?.statuses, anime?.genres, anime?.startYears])

    return (
        <div className="space-y-8" data-anilist-stats-anime>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4" data-anilist-stats-anime-hero>
                <MetricCard
                    icon={<PiTelevisionSimpleBold />}
                    label="Total Anime"
                    value={anime?.count ?? 0}
                />
                <MetricCard
                    icon={<FiBarChart2 />}
                    label="Episodes"
                    value={(anime?.episodesWatched ?? 0).toLocaleString()}
                />
                <MetricCard
                    icon={<FiClock />}
                    label="Watch Time"
                    value={daysWatched}
                    sub={`${hoursWatched.toLocaleString()} hours`}
                />
                <MetricCard
                    icon={<LuStar />}
                    label="Mean Score"
                    value={formatScore(anime?.meanScore)}
                    accent
                />
                <MetricCard
                    icon={<LuTrendingUp />}
                    label="Started This Year"
                    value={thisYearData?.count ?? 0}
                    sub={thisYearData ? `${thisYear}` : "No activity"}
                />
                <MetricCard
                    icon={<LuTrendingUp />}
                    label="Started Last Year"
                    value={lastYearData?.count ?? 0}
                    sub={lastYearData ? `${thisYear - 1}` : "No activity"}
                />
            </div>

            {highlights.length > 0 && (
                <HighlightsGrid highlights={highlights} data-anilist-stats-anime-insights />
            )}

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6" data-anilist-stats-anime-charts>
                {scoreData.length > 0 && (
                    <ChartSection
                        title="Score Distribution"
                        description="Titles by score"
                        data-anilist-stats-anime-scores
                    >
                        <BarChart
                            className="h-64"
                            data={scoreData}
                            index="name"
                            categories={["Titles"]}
                            colors={["blue"]}
                            showLegend={false}
                            allowDecimals={false}
                        />
                    </ChartSection>
                )}

                {statusData.length > 0 && (
                    <StatusChart
                        title="List Status"
                        data={statusData}
                        data-anilist-stats-anime-statuses
                    />
                )}

                {formatData.length > 0 && (
                    <ChartSection
                        title="Formats"
                        description="Format mix by title count"
                        className="lg:col-span-2"
                        data-anilist-stats-anime-formats
                    >
                        <RankingGrid rows={formatData} />
                    </ChartSection>
                )}

                {genreData.length > 0 && (
                    <ChartSection
                        title="Top Genres"
                        description="Most watched genres"
                        className="lg:col-span-2"
                        data-anilist-stats-anime-genres
                    >
                        <RankingGrid rows={genreData.slice(0, 10)} />
                    </ChartSection>
                )}
            </div>

            {startYearData.length > 0 && (
                <ChartSection
                    title="Started by Year"
                    description="Titles started each year"
                    data-anilist-stats-anime-activity
                >
                    <AreaChart
                        data={startYearData}
                        index="name"
                        categories={["Titles"]}
                        colors={["blue"]}
                        curveType="linear"
                        showDots={false}
                        showLegend={false}
                        allowDecimals={false}
                    />
                </ChartSection>
            )}

            {releaseYearData.length > 0 && (
                <ChartSection
                    title="Release Years"
                    description="Titles grouped by original release year"
                    data-anilist-stats-anime-years
                >
                    <BarChart
                        data={releaseYearData}
                        index="name"
                        categories={["Titles"]}
                        colors={["blue"]}
                        showLegend={false}
                        allowDecimals={false}
                    />
                </ChartSection>
            )}

            {studioData.length > 0 && (
                <ChartSection
                    title="Top Studios"
                    description="Studios with the most watched titles"
                    data-anilist-stats-anime-studios
                >
                    <RankingGrid rows={studioData.slice(0, 10)} />
                </ChartSection>
            )}
        </div>
    )
}

function MangaStatsView({ stats }: { stats?: AL_Stats }) {
    const manga = stats?.mangaStats
    const thisYear = new Date().getFullYear()
    const thisYearData = manga?.startYears?.find(y => y.startYear === thisYear)
    const lastYearData = manga?.startYears?.find(y => y.startYear === thisYear - 1)

    const scoreData = React.useMemo(() => toScoreData(manga?.scores), [manga?.scores])
    const statusData = React.useMemo(() => toStatusData(manga?.statuses, "Reading"), [manga?.statuses])
    const genreData = React.useMemo(() => toRankingRows(manga?.genres, "titles", "chapters"), [manga?.genres])
    const startYearData = React.useMemo(() => toStartYearData(manga?.startYears), [manga?.startYears])
    const releaseYearData = React.useMemo(() => toReleaseYearData(manga?.releaseYears), [manga?.releaseYears])
    const highlights = React.useMemo(() => toHighlights({
        scores: manga?.scores,
        statuses: manga?.statuses,
        genres: manga?.genres,
        startYears: manga?.startYears,
        unit: "titles",
    }), [manga?.scores, manga?.statuses, manga?.genres, manga?.startYears])

    return (
        <div className="space-y-8" data-anilist-stats-manga>
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-4" data-anilist-stats-manga-hero>
                <MetricCard
                    icon={<FiBookOpen />}
                    label="Total Manga"
                    value={manga?.count ?? 0}
                />
                <MetricCard
                    icon={<FiBarChart2 />}
                    label="Chapters"
                    value={(manga?.chaptersRead ?? 0).toLocaleString()}
                />
                <MetricCard
                    icon={<LuStar />}
                    label="Mean Score"
                    value={formatScore(manga?.meanScore)}
                    accent
                />
                <MetricCard
                    icon={<LuTrendingUp />}
                    label="Started This Year"
                    value={thisYearData?.count ?? 0}
                    sub={thisYearData ? `${thisYear}` : "No activity"}
                />
                <MetricCard
                    icon={<LuTrendingUp />}
                    label="Started Last Year"
                    value={lastYearData?.count ?? 0}
                    sub={lastYearData ? `${thisYear - 1}` : "No activity"}
                />
            </div>

            {highlights.length > 0 && (
                <HighlightsGrid highlights={highlights} data-anilist-stats-manga-insights />
            )}

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6" data-anilist-stats-manga-charts>
                {scoreData.length > 0 && (
                    <ChartSection
                        title="Score Distribution"
                        description="Titles by score"
                        data-anilist-stats-manga-scores
                    >
                        <BarChart
                            className="h-64"
                            data={scoreData}
                            index="name"
                            categories={["Titles"]}
                            colors={["blue"]}
                            showLegend={false}
                            allowDecimals={false}
                        />
                    </ChartSection>
                )}

                {statusData.length > 0 && (
                    <StatusChart
                        title="List Status"
                        data={statusData}
                        data-anilist-stats-manga-statuses
                    />
                )}

                {genreData.length > 0 && (
                    <ChartSection
                        title="Top Genres"
                        description="Most read genres"
                        className="lg:col-span-2"
                        data-anilist-stats-manga-genres
                    >
                        <RankingGrid rows={genreData.slice(0, 10)} />
                    </ChartSection>
                )}
            </div>

            {startYearData.length > 0 && (
                <ChartSection
                    title="Started by Year"
                    description="Titles started each year"
                    data-anilist-stats-manga-activity
                >
                    <AreaChart
                        data={startYearData}
                        index="name"
                        categories={["Titles"]}
                        colors={["blue"]}
                        curveType="linear"
                        showDots={false}
                        showLegend={false}
                        allowDecimals={false}
                    />
                </ChartSection>
            )}

            {releaseYearData.length > 0 && (
                <ChartSection
                    title="Release Years"
                    description="Titles grouped by original release year"
                    data-anilist-stats-manga-years
                >
                    <BarChart
                        data={releaseYearData}
                        index="name"
                        categories={["Titles"]}
                        colors={["blue"]}
                        showLegend={false}
                        allowDecimals={false}
                    />
                </ChartSection>
            )}
        </div>
    )
}

type StatusChartRow = {
    name: string
    count: number
    color: ChartColor
}

type RankingRow = {
    id: string | number
    name: string
    count: number
    maxCount: number
    valueLabel: string
    meta?: string
}

type Highlight = {
    label: string
    value: string
    detail?: string
}

function toScoreData(scores: AL_UserScoreStats[] | undefined) {
    if (!scores?.length) return []

    return [...scores]
        .filter(s => (s.score ?? 0) > 0 && s.count > 0)
        .sort((a, b) => (a.score ?? 0) - (b.score ?? 0))
        .map(s => ({
            name: formatScoreBucket(s.score),
            Titles: s.count,
        }))
}

function toStatusData(statuses: AL_UserStatusStats[] | undefined, currentLabel = "Watching"): StatusChartRow[] {
    if (!statuses?.length) return []

    return [...statuses]
        .filter(s => s.status && s.count > 0)
        .sort((a, b) => statusOrder.indexOf(a.status as string) - statusOrder.indexOf(b.status as string))
        .map(s => {
            const status = s.status as string
            return {
                name: (statusName[status] ?? status).replace("Watching", currentLabel),
                count: s.count,
                color: statusColors[status] ?? "gray",
            }
        })
}

function toHighlights({
    scores,
    statuses,
    genres,
    startYears,
    unit,
}: {
    scores?: AL_UserScoreStats[]
    statuses?: AL_UserStatusStats[]
    genres?: AL_UserGenreStats[]
    startYears?: AL_UserStartYearStats[]
    unit: string
}): Highlight[] {
    const modeScore = getModeScore(scores)
    const completion = getCompletion(statuses)
    const topGenre = getTopGenre(genres)
    const peakYear = getPeakYear(startYears)

    return [
        modeScore && {
            label: "Most Used Score",
            value: modeScore.score,
            detail: `${modeScore.count.toLocaleString()} ${unit}`,
        },
        completion && {
            label: "Completion",
            value: `${completion.rate}%`,
            detail: `${completion.completed.toLocaleString()} of ${completion.started.toLocaleString()} started`,
        },
        topGenre && {
            label: "Top Genre",
            value: topGenre.genre,
            detail: joinMeta(`${topGenre.count.toLocaleString()} ${unit}`, formatAvg(topGenre.meanScore)),
        },
        peakYear && {
            label: "Peak Start Year",
            value: String(peakYear.year),
            detail: `${peakYear.count.toLocaleString()} ${unit} started`,
        },
    ].filter(Boolean) as Highlight[]
}

function getModeScore(scores: AL_UserScoreStats[] | undefined) {
    if (!scores?.length) return null

    const score = [...scores]
        .filter(item => (item.score ?? 0) > 0 && item.count > 0)
        .sort((a, b) => b.count - a.count)[0]

    if (!score) return null

    return {
        score: formatScoreBucket(score.score),
        count: score.count,
    }
}

function getCompletion(statuses: AL_UserStatusStats[] | undefined) {
    if (!statuses?.length) return null

    const completed = statuses.find(s => s.status === "COMPLETED")?.count ?? 0
    const planning = statuses.find(s => s.status === "PLANNING")?.count ?? 0
    const total = statuses.reduce((sum, status) => sum + status.count, 0)
    const started = total - planning

    if (started <= 0) return null

    return {
        completed,
        started,
        rate: Math.round((completed / started) * 100),
    }
}

function getTopGenre(genres: AL_UserGenreStats[] | undefined) {
    if (!genres?.length) return null

    const genre = [...genres]
        .filter(item => item.count > 0)
        .sort((a, b) => b.count - a.count)[0]

    if (!genre) return null

    return {
        genre: genre.genre ?? "Unknown",
        count: genre.count,
        meanScore: genre.meanScore,
    }
}

function getPeakYear(startYears: AL_UserStartYearStats[] | undefined) {
    if (!startYears?.length) return null

    const year = [...startYears]
        .filter(item => item.startYear && item.count > 0)
        .sort((a, b) => b.count - a.count)[0]

    if (!year?.startYear) return null

    return {
        year: year.startYear,
        count: year.count,
    }
}

function toFormatRows(formats: AL_UserFormatStats[] | undefined): RankingRow[] {
    if (!formats?.length) return []

    const rows = [...formats]
        .filter(f => f.count > 0)
        .sort((a, b) => b.count - a.count)
    const maxCount = rows[0]?.count ?? 1

    return rows.map(f => {
        const name = formatName[f.format ?? ""] ?? f.format ?? "Unknown"
        const hours = Math.round(f.minutesWatched / 60)

        return {
            id: name,
            name,
            count: f.count,
            maxCount,
            valueLabel: `${f.count.toLocaleString()} titles`,
            meta: joinMeta(`${hours.toLocaleString()}h watched`, formatAvg(f.meanScore)),
        }
    })
}

function toRankingRows(
    stats: AL_UserGenreStats[] | undefined,
    unit: string,
    secondaryUnit?: "chapters",
): RankingRow[] {
    if (!stats?.length) return []

    const rows = [...stats]
        .filter(item => item.count > 0)
        .sort((a, b) => b.count - a.count)
    const maxCount = rows[0]?.count ?? 1

    return rows.map(item => {
        const countLabel = `${item.count.toLocaleString()} ${unit}`
        const secondaryLabel = secondaryUnit ? `${item.chaptersRead.toLocaleString()} chapters` : undefined

        return {
            id: item.genre ?? "Unknown",
            name: item.genre ?? "Unknown",
            count: item.count,
            maxCount,
            valueLabel: countLabel,
            meta: joinMeta(secondaryLabel, formatAvg(item.meanScore)),
        }
    })
}

function toStudioRows(studios: AL_UserStudioStats[] | undefined): RankingRow[] {
    if (!studios?.length) return []

    const rows = [...studios]
        .filter(item => item.count > 0)
        .sort((a, b) => b.count - a.count)
    const maxCount = rows[0]?.count ?? 1

    return rows.map((item, i) => ({
        id: item.studio?.id ?? i,
        name: item.studio?.name ?? "Unknown",
        count: item.count,
        maxCount,
        valueLabel: `${item.count.toLocaleString()} titles`,
        meta: formatAvg(item.meanScore),
    }))
}

function toStartYearData(years: AL_UserStartYearStats[] | undefined) {
    if (!years?.length) return []

    return [...years]
        .filter(y => y.startYear && y.count > 0)
        .sort((a, b) => (a.startYear ?? 0) - (b.startYear ?? 0))
        .map(y => ({
            name: y.startYear,
            Titles: y.count,
        }))
}

function toReleaseYearData(years: AL_UserReleaseYearStats[] | undefined) {
    if (!years?.length) return []

    return [...years]
        .filter(y => y.releaseYear && y.count > 0)
        .sort((a, b) => (a.releaseYear ?? 0) - (b.releaseYear ?? 0))
        .map(y => ({
            name: y.releaseYear,
            Titles: y.count,
        }))
}

function formatScore(score: number | undefined) {
    if (!score) return "-"
    return (score / 10).toFixed(1)
}

function formatAvg(score: number | undefined) {
    if (!score) return undefined
    return `Score Avg ${formatScore(score)}`
}

function formatScoreBucket(score: number | undefined) {
    if (!score) return "-"
    return String(score > 10 ? score / 10 : score)
}

function joinMeta(...parts: Array<string | undefined>) {
    const meta = parts.filter(Boolean).join(" - ")
    return meta || undefined
}

function MetricCard({ icon, label, value, sub, accent }: {
    icon: React.ReactNode
    label: string
    value: string | number
    sub?: string
    accent?: boolean
}) {
    return (
        <div className="rounded-lg border border-[--border] bg-[--paper] p-4 space-y-1 min-w-0">
            <div className="flex items-center gap-2 text-[--muted] text-sm min-w-0">
                <span className="text-base shrink-0">{icon}</span>
                <span className="truncate">{label}</span>
            </div>
            <p
                className={cn(
                    "text-2xl md:text-3xl font-bold tabular-nums leading-tight",
                    accent && "text-[--brand]",
                )}
            >
                {value}
            </p>
            {sub && <p className="text-xs text-[--muted] truncate">{sub}</p>}
        </div>
    )
}

function HighlightsGrid({ highlights, className, ...rest }: {
    highlights: Highlight[]
    className?: string
} & React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div
            className={cn("grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4", className)}
            {...rest}
        >
            {highlights.map(highlight => (
                <div key={highlight.label} className="rounded-lg border border-[--border] bg-[--paper] p-4 min-w-0">
                    <p className="text-xs text-[--muted] uppercase">{highlight.label}</p>
                    <p className="text-lg font-semibold truncate mt-1">{highlight.value}</p>
                    {highlight.detail && <p className="text-xs text-[--muted] truncate mt-1">{highlight.detail}</p>}
                </div>
            ))}
        </div>
    )
}

function StatusChart({ title, data, className, ...rest }: {
    title: string
    data: StatusChartRow[]
    className?: string
} & React.HTMLAttributes<HTMLDivElement>) {
    return (
        <ChartSection title={title} description="Title count by current list status" className={className} {...rest}>
            <DonutChart
                data={data}
                index="name"
                category="count"
                variant="donut"
                colors={data.map(item => item.color)}
            />
            <Legend
                categories={data.map(item => `${item.name} (${item.count.toLocaleString()})`)}
                colors={data.map(item => item.color)}
                className="mt-4 justify-center"
            />
        </ChartSection>
    )
}

function ChartSection({ title, description, children, className, ...rest }: {
    title: string
    description?: string
    children: React.ReactNode
    className?: string
} & React.HTMLAttributes<HTMLDivElement>) {
    return (
        <div
            className={cn("rounded-lg border border-[--border] bg-[--paper] p-4 space-y-4 min-w-0", className)}
            {...rest}
        >
            <div className="space-y-1">
                <h4 className="text-sm font-medium text-[--muted] uppercase">{title}</h4>
                {description && <p className="text-xs text-[--muted]">{description}</p>}
            </div>
            {children}
        </div>
    )
}

function RankingGrid({ rows }: { rows: RankingRow[] }) {
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-3">
            {rows.map((row, i) => (
                <RankedProgressRow key={row.id} rank={i + 1} row={row} />
            ))}
        </div>
    )
}

function RankingList({ rows }: { rows: RankingRow[] }) {
    return (
        <div className="space-y-3">
            {rows.map((row, i) => (
                <RankedProgressRow key={row.id} rank={i + 1} row={row} />
            ))}
        </div>
    )
}

function RankedProgressRow({ rank, row }: {
    rank: number
    row: RankingRow
}) {
    const pct = row.maxCount > 0 ? Math.round((row.count / row.maxCount) * 100) : 0

    return (
        <div className="flex items-center gap-3 text-sm min-w-0">
            <span className="text-[--muted] w-5 text-right font-medium tabular-nums shrink-0">{rank}</span>
            <div className="flex-1 min-w-0">
                <div className="flex items-baseline justify-between gap-3 mb-1">
                    <span className="truncate font-medium">{row.name}</span>
                    <span className="text-[--muted] shrink-0 text-xs tabular-nums">{row.valueLabel}</span>
                </div>
                <div className="h-1.5 rounded-full bg-[--subtle] overflow-hidden">
                    <div
                        className="h-full rounded-full bg-[--blue] transition-all duration-500"
                        style={{ width: `${pct}%` }}
                    />
                </div>
                {row.meta && <p className="mt-1 text-xs text-[--muted] truncate">{row.meta}</p>}
            </div>
        </div>
    )
}
