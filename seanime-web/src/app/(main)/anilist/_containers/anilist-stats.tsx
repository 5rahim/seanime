import { AL_Stats } from "@/api/generated/types"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { AreaChart, BarChart, DonutChart } from "@/components/ui/charts"
import { Separator } from "@/components/ui/separator"
import { Stats } from "@/components/ui/stats"
import React from "react"
import { FaRegStar } from "react-icons/fa"
import { FiBookOpen } from "react-icons/fi"
import { LuHourglass } from "react-icons/lu"
import { PiTelevisionSimpleBold } from "react-icons/pi"
import { TbHistory } from "react-icons/tb"

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
    CURRENT: "Current",
    PLANNING: "Planning",
    COMPLETED: "Completed",
    DROPPED: "Dropped",
    PAUSED: "Paused",
    REPEATING: "Repeating",
}

export function AnilistStats(props: AnilistStatsProps) {

    const {
        stats,
        isLoading,
    } = props

    const anime_formatsStats = React.useMemo(() => {
        if (!stats?.animeStats?.formats) return []

        return stats.animeStats.formats.map((item) => {
            return {
                name: formatName[item.format as string],
                count: item.count,
                hoursWatched: Math.round(item.minutesWatched / 60),
                meanScore: Number((item.meanScore / 10).toFixed(1)),
            }
        })
    }, [stats?.animeStats?.formats])

    const anime_statusesStats = React.useMemo(() => {
        if (!stats?.animeStats?.statuses) return []

        return stats.animeStats.statuses.map((item) => {
            return {
                name: statusName[item.status as string],
                count: item.count,
                hoursWatched: Math.round(item.minutesWatched / 60),
                meanScore: Number((item.meanScore / 10).toFixed(1)),
            }
        })
    }, [stats?.animeStats?.statuses])

    const anime_genresStats = React.useMemo(() => {
        if (!stats?.animeStats?.genres) return []

        return stats.animeStats.genres.map((item) => {
            return {
                name: item.genre,
                "Count": item.count,
                hoursWatched: Math.round(item.minutesWatched / 60),
                "Average score": Number((item.meanScore / 10).toFixed(1)),
            }
        }).sort((a, b) => b["Count"] - a["Count"])
    }, [stats?.animeStats?.genres])

    const [anime_thisYearStats, anime_lastYearStats] = React.useMemo(() => {
        if (!stats?.animeStats?.startYears) return []
        const thisYear = new Date().getFullYear()
        return [
            stats.animeStats.startYears.find((item) => item.startYear === thisYear),
            stats.animeStats.startYears.find((item) => item.startYear === thisYear - 1),
        ]
    }, [stats?.animeStats?.startYears])

    const anime_releaseYearsStats = React.useMemo(() => {
        if (!stats?.animeStats?.releaseYears) return []

        return stats.animeStats.releaseYears.sort((a, b) => a.releaseYear! - b.releaseYear!).map((item) => {
            return {
                name: item.releaseYear,
                "Count": item.count,
                "Hours watched": Math.round(item.minutesWatched / 60),
                "Mean score": Number((item.meanScore / 10).toFixed(1)),
            }
        })
    }, [stats?.animeStats?.releaseYears])

    /////

    const manga_statusesStats = React.useMemo(() => {
        if (!stats?.mangaStats?.statuses) return []

        return stats.mangaStats.statuses.map((item) => {
            return {
                name: statusName[item.status as string],
                count: item.count,
                chaptersRead: item.chaptersRead,
                meanScore: Number((item.meanScore / 10).toFixed(1)),
            }
        })
    }, [stats?.mangaStats?.statuses])

    const manga_genresStats = React.useMemo(() => {
        if (!stats?.mangaStats?.genres) return []

        return stats.mangaStats.genres.map((item) => {
            return {
                name: item.genre,
                "Count": item.count,
                chaptersRead: item.chaptersRead,
                "Average score": Number((item.meanScore / 10).toFixed(1)),
            }
        }).sort((a, b) => b["Count"] - a["Count"])
    }, [stats?.mangaStats?.genres])

    const [manga_thisYearStats, manga_lastYearStats] = React.useMemo(() => {
        if (!stats?.mangaStats?.startYears) return []
        const thisYear = new Date().getFullYear()
        return [
            stats.mangaStats.startYears.find((item) => item.startYear === thisYear),
            stats.mangaStats.startYears.find((item) => item.startYear === thisYear - 1),
        ]
    }, [stats?.mangaStats?.startYears])

    const manga_releaseYearsStats = React.useMemo(() => {
        if (!stats?.mangaStats?.releaseYears) return []

        return stats.mangaStats.releaseYears.sort((a, b) => a.releaseYear! - b.releaseYear!).map((item) => {
            return {
                name: item.releaseYear,
                "Count": item.count,
                "Chapters read": item.chaptersRead,
                "Mean score": Number((item.meanScore / 10).toFixed(1)),
            }
        })
    }, [stats?.mangaStats?.releaseYears])

    return (
        <AppLayoutStack className="py-4 space-y-10" data-anilist-stats>

            <h1 className="text-center" data-anilist-stats-anime-title>Anime</h1>

            <div data-anilist-stats-anime-stats>
                <Stats
                    className="w-full"
                    size="lg"
                    items={[
                        {
                            icon: <PiTelevisionSimpleBold />,
                            name: "Total Anime",
                            value: stats?.animeStats?.count ?? 0,
                        },
                        {
                            icon: <LuHourglass />,
                            name: "Watch time",
                            value: Math.round((stats?.animeStats?.minutesWatched ?? 0) / 60),
                            unit: "hours",
                        },
                        {
                            icon: <FaRegStar />,
                            name: "Average score",
                            value: ((stats?.animeStats?.meanScore ?? 0) / 10).toFixed(1),
                        },
                    ]}
                />
                <Separator />
                <Stats
                    className="w-full"
                    size="lg"
                    items={[
                        {
                            icon: <PiTelevisionSimpleBold />,
                            name: "Anime watched this year",
                            value: anime_thisYearStats?.count ?? 0,
                        },
                        {
                            icon: <TbHistory />,
                            name: "Anime watched last year",
                            value: anime_lastYearStats?.count ?? 0,
                        },
                        {
                            icon: <FaRegStar />,
                            name: "Average score this year",
                            value: ((anime_thisYearStats?.meanScore ?? 0) / 10).toFixed(1),
                        },
                    ]}
                />
            </div>

            <h3 className="text-center" data-anilist-stats-anime-formats-title>Formats</h3>

            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 w-full" data-anilist-stats-anime-formats-container>
                <ChartContainer legend="Total" data-anilist-stats-anime-formats-container-total>
                    <DonutChart
                        data={anime_formatsStats}
                        index="name"
                        category="count"
                        variant="pie"
                    />
                </ChartContainer>
                <ChartContainer legend="Hours watched" data-anilist-stats-anime-formats-container-hours-watched>
                    <DonutChart
                        data={anime_formatsStats}
                        index="name"
                        category="hoursWatched"
                        variant="pie"
                    />
                </ChartContainer>
                <ChartContainer legend="Average score" data-anilist-stats-anime-formats-container-average-score>
                    <DonutChart
                        data={anime_formatsStats}
                        index="name"
                        category="meanScore"
                        variant="pie"
                    />
                </ChartContainer>
            </div>

            <Separator />

            <h3 className="text-center" data-anilist-stats-anime-statuses-title>Statuses</h3>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 w-full" data-anilist-stats-anime-statuses-container>
                <ChartContainer legend="Total" data-anilist-stats-anime-statuses-container-total>
                    <DonutChart
                        data={anime_statusesStats}
                        index="name"
                        category="count"
                        variant="pie"
                    />
                </ChartContainer>
                <ChartContainer legend="Hours watched" data-anilist-stats-anime-statuses-container-hours-watched>
                    <DonutChart
                        data={anime_statusesStats}
                        index="name"
                        category="hoursWatched"
                        variant="pie"
                    />
                </ChartContainer>
            </div>

            <Separator />

            <h3 className="text-center" data-anilist-stats-anime-genres-title>Genres</h3>

            <div className="grid grid-cols-1 gap-6 w-full" data-anilist-stats-anime-genres-container>
                <ChartContainer legend="Favorite genres" data-anilist-stats-anime-genres-container-favorite-genres>
                    <BarChart
                        data={anime_genresStats}
                        index="name"
                        categories={["Count", "Average score"]}
                        colors={["brand", "blue"]}
                    />
                </ChartContainer>
            </div>

            <Separator />

            <h3 className="text-center" data-anilist-stats-anime-years-title>Years</h3>

            <div className="grid grid-cols-1 gap-6 w-full" data-anilist-stats-anime-years-container>
                <ChartContainer legend="Anime watched per release year" data-anilist-stats-anime-years-container-anime-watched-per-release-year>
                    <AreaChart
                        data={anime_releaseYearsStats}
                        index="name"
                        categories={["Count"]}
                        angledLabels
                    />
                </ChartContainer>
            </div>

            {/*////////////////////////////////////////////////////*/}
            {/*////////////////////////////////////////////////////*/}
            {/*////////////////////////////////////////////////////*/}

            <h1 className="text-center pt-20" data-anilist-stats-manga-title>Manga</h1>

            <div data-anilist-stats-manga-stats>
                <Stats
                    className="w-full"
                    size="lg"
                    items={[
                        {
                            icon: <FiBookOpen />,
                            name: "Total Manga",
                            value: stats?.mangaStats?.count ?? 0,
                        },
                        {
                            icon: <LuHourglass />,
                            name: "Total chapters",
                            value: stats?.mangaStats?.chaptersRead ?? 0,
                        },
                        {
                            icon: <FaRegStar />,
                            name: "Average score",
                            value: ((stats?.mangaStats?.meanScore ?? 0) / 10).toFixed(1),
                        },
                    ]}
                />
                <Separator />
                <Stats
                    className="w-full"
                    size="lg"
                    items={[
                        {
                            icon: <FiBookOpen />,
                            name: "Manga read this year",
                            value: manga_thisYearStats?.count ?? 0,
                        },
                        {
                            icon: <TbHistory />,
                            name: "Manga read last year",
                            value: manga_lastYearStats?.count ?? 0,
                        },
                        {
                            icon: <FaRegStar />,
                            name: "Average score this year",
                            value: ((manga_thisYearStats?.meanScore ?? 0) / 10).toFixed(1),
                        },
                    ]}
                />
            </div>

            <Separator />

            <h3 className="text-center" data-anilist-stats-manga-statuses-title>Statuses</h3>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 w-full" data-anilist-stats-manga-statuses-container>
                <ChartContainer legend="Total" data-anilist-stats-manga-statuses-container-total>
                    <DonutChart
                        data={manga_statusesStats}
                        index="name"
                        category="count"
                        variant="pie"
                    />
                </ChartContainer>
                <ChartContainer legend="Chapters read" data-anilist-stats-manga-statuses-container-chapters-read>
                    <DonutChart
                        data={manga_statusesStats}
                        index="name"
                        category="chaptersRead"
                        variant="pie"
                    />
                </ChartContainer>
            </div>

            <Separator />

            <h3 className="text-center" data-anilist-stats-manga-genres-title>Genres</h3>

            <div className="grid grid-cols-1 gap-6 w-full" data-anilist-stats-manga-genres-container>
                <ChartContainer legend="Favorite genres" data-anilist-stats-manga-genres-container-favorite-genres>
                    <BarChart
                        data={manga_genresStats}
                        index="name"
                        categories={["Count", "Average score"]}
                        colors={["brand", "blue"]}
                    />
                </ChartContainer>
            </div>

            <Separator />

            <h3 className="text-center" data-anilist-stats-manga-years-title>Years</h3>

            <div className="grid grid-cols-1 gap-6 w-full" data-anilist-stats-manga-years-container>
                <ChartContainer legend="Manga read per release year" data-anilist-stats-manga-years-container-manga-read-per-release-year>
                    <AreaChart
                        data={manga_releaseYearsStats}
                        index="name"
                        categories={["Count"]}
                        angledLabels
                    />
                </ChartContainer>
            </div>

        </AppLayoutStack>
    )
}

function ChartContainer(props: { children: React.ReactNode, legend: string }) {
    return (
        <div className="text-center w-full space-y-4" data-anilist-stats-chart-container>
            {props.children}
            <p className="text-center text-lg font-semibold">{props.legend}</p>
        </div>
    )
}
