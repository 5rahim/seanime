import { useAnilistListRecentAiringAnime } from "@/api/hooks/anilist.hooks"
import { SeaLink } from "@/components/shared/sea-link"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Separator } from "@/components/ui/separator"
import { format, isSameMonth, isToday, subDays } from "date-fns"
import { addDays } from "date-fns/addDays"
import { isSameDay } from "date-fns/isSameDay"
import { SeaContextMenu } from "@/app/(main)/_features/context-menu/sea-context-menu"
import { ContextMenuGroup, ContextMenuItem, ContextMenuLabel, ContextMenuTrigger } from "@/components/ui/context-menu"
import { useMediaPreviewModal } from "@/app/(main)/_features/media/_containers/media-preview-modal"
import { useRouter } from "next/navigation"
import Image from "next/image"
import React from "react"


export function DiscoverAiringSchedule() {
    const { data, isLoading } = useAnilistListRecentAiringAnime({
        page: 1,
        perPage: 50,
        airingAt_lesser: Math.floor(addDays(new Date(), 14).getTime() / 1000),
        airingAt_greater: Math.floor(subDays(new Date(), 2).getTime() / 1000),
        notYetAired: true,
        sort: ["TIME"],
    })
    const { data: data2, isLoading: isLoading2 } = useAnilistListRecentAiringAnime({
        page: 2,
        perPage: 50,
        airingAt_lesser: Math.floor(addDays(new Date(), 14).getTime() / 1000),
        airingAt_greater: Math.floor(subDays(new Date(), 2).getTime() / 1000),
        notYetAired: true,
        sort: ["TIME"],
    })

    const media = React.useMemo(() => [...(data?.Page?.airingSchedules?.filter(item => item?.media?.isAdult === false
        && item?.media?.type === "ANIME"
        && item?.media?.countryOfOrigin === "JP"
        && item?.media?.format !== "TV_SHORT",
    ).filter(Boolean) || []),
    ...(data2?.Page?.airingSchedules?.filter(item => item?.media?.isAdult === false
        && item?.media?.type === "ANIME"
        && item?.media?.countryOfOrigin === "JP"
        && item?.media?.format !== "TV_SHORT",
    ).filter(Boolean) || []),
    ], [isLoading, isLoading2])

    const router = useRouter()
    const { setPreviewModalMediaId } = useMediaPreviewModal()

    // State for the current displayed month
    const [currentDate, setCurrentDate] = React.useState(new Date())

    const days = React.useMemo(() => {

        // Ensure startOfWeek aligns with the correct day
        const start = subDays(new Date(), 1)

        const daysArray = []
        let day = new Date(start.setHours(0, 0, 0, 0))  // Ensure the day starts at midnight
        const endDate = addDays(day, 14)  // 14-day range from the current start

        while (day <= endDate) {
            const upcomingMedia = media.filter((item) => !!item?.airingAt && isSameDay(new Date(item.airingAt * 1000), day)).map((item) => {
                if (item.media?.id === 162804) console.log(item.airingAt)
                return {
                    id: item.id + item?.episode!,
                    name: item.media?.title?.userPreferred,
                    time: format(new Date(item?.airingAt! * 1000), "h:mm a"),
                    datetime: format(new Date(item?.airingAt! * 1000), "yyyy-MM-dd'T'HH:mm"),
                    href: `/entry?id=${item.id}`,
                    media: item.media,
                    episode: item?.media?.nextAiringEpisode?.episode || 1,
                }
            })

            daysArray.push({
                date: format(day, "yyyy-MM-dd'T'HH:mm"),
                isCurrentMonth: isSameMonth(day, currentDate),
                isToday: isToday(day),
                isSelected: false,
                events: upcomingMedia,
            })
            day = addDays(day, 1)
        }
        return daysArray
    }, [media, currentDate])

    if (isLoading || isLoading2) return <LoadingSpinner />

    if (!data?.Page?.airingSchedules?.length) return null

    return (
        <div className="space-y-4 z-[5] relative" data-discover-airing-schedule-container>
            <h2 className="text-center">Airing schedule</h2>
            <div className="space-y-6">
                {days.map((day, index) => {
                    if (day.events.length === 0) return null
                    return (
                        <React.Fragment key={day.date}>
                            <div className="flex flex-col gap-2">
                                <div className="flex items-center gap-2">
                                    <h3 className="font-semibold">{format(new Date(day.date), "EEEE, PP")}</h3>
                                    {day.isToday && <span className="text-[--muted]">Today</span>}
                                </div>
                                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
                                    {day.events?.toSorted((a, b) => a.datetime.localeCompare(b.datetime))?.map((event, index) => {
                                        return (
                                            <React.Fragment key={event.id}>
                                                <SeaContextMenu
                                                    content={<ContextMenuGroup>
                                                        <ContextMenuLabel className="text-[--muted] line-clamp-2 py-0 my-2">
                                                            {event.media?.title?.userPreferred}
                                                        </ContextMenuLabel>
                                                        <ContextMenuItem
                                                            onClick={() => {
                                                                router.push(`/entry?id=${event.media?.id}`)
                                                            }}
                                                        >
                                                            Open page
                                                        </ContextMenuItem>
                                                        <ContextMenuItem
                                                            onClick={() => {
                                                                setPreviewModalMediaId(event.media?.id || 0, "anime")
                                                            }}
                                                        >
                                                            Preview
                                                        </ContextMenuItem>
                                                    </ContextMenuGroup>}
                                                >
                                                    <ContextMenuTrigger>
                                                        <div key={String(`${event.id}${index}`)}
                                                            className="flex gap-3 bg-[--background] rounded-[--radius-md] p-2"
                                                        >
                                                            <div
                                                                className="w-[5rem] h-[5rem] rounded-[--radius] flex-none object-cover object-center overflow-hidden relative"
                                                            >
                                                                <Image
                                                                    src={event.media?.coverImage?.large || event.media?.bannerImage || "/no-cover.png"}
                                                                    alt="banner"
                                                                    fill
                                                                    quality={80}
                                                                    priority
                                                                    sizes="20rem"
                                                                    className="object-cover object-center"
                                                                />
                                                            </div>

                                                            <div className="space-y-1">
                                                                <SeaLink
                                                                    href={`/entry?id=${event.media?.id}`}
                                                                    className="font-medium tracking-wide line-clamp-1"
                                                                >{event.media?.title?.userPreferred}</SeaLink>

                                                                <p className="text-[--muted]">
                                                                    Ep {event.episode} airing at {event.time}
                                                                </p>
                                                            </div>
                                                        </div>
                                                    </ContextMenuTrigger>
                                                </SeaContextMenu>
                                            </React.Fragment>
                                        )
                                    })}
                                </div>
                            </div>
                            {index < 7 && <Separator />}
                        </React.Fragment>
                    )
                })}
            </div>
        </div>
    )
}
