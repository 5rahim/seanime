import { AL_BaseMedia, Anime_MediaEntryEpisode } from "@/api/generated/types"
import { useServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Tooltip } from "@/components/ui/tooltip"
import { addMonths, endOfMonth, endOfWeek, format, isSameMonth, isToday, startOfMonth, startOfWeek, subMonths } from "date-fns"
import { addDays } from "date-fns/addDays"
import { isSameDay } from "date-fns/isSameDay"
import { enUS } from "date-fns/locale"
import Image from "next/image"
import Link from "next/link"
import React, { Fragment } from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight } from "react-icons/ai"

type WeekCalendarProps = {
    children?: React.ReactNode
    media: AL_BaseMedia[]
    missingEpisodes: Anime_MediaEntryEpisode[]
}

export function MonthCalendar(props: WeekCalendarProps) {

    const {
        children,
        media,
        missingEpisodes,
        ...rest
    } = props

    const serverStatus = useServerStatus()

    // State for the current displayed month
    const [currentDate, setCurrentDate] = React.useState(new Date())

    // Function to go to the previous month
    const goToPreviousMonth = () => {
        setCurrentDate(prevDate => subMonths(prevDate, 1))
    }

    // Function to go to the next month
    const goToNextMonth = () => {
        setCurrentDate(prevDate => addMonths(prevDate, 1))
    }

    const days = React.useMemo(() => {
        const startOfCurrentMonth = startOfMonth(currentDate)
        const endOfCurrentMonth = endOfMonth(currentDate)

        const startOfCalendar = startOfWeek(startOfCurrentMonth, { weekStartsOn: 1, locale: enUS })
        const endOfCalendar = endOfWeek(endOfCurrentMonth, { weekStartsOn: 1, locale: enUS })

        const daysArray = []
        let day = startOfCalendar

        while (day <= endOfCalendar) {
            const upcomingMedia = media.filter((item) => !!item.nextAiringEpisode?.airingAt && isSameDay(new Date(item.nextAiringEpisode?.airingAt * 1000),
                day)).map((item) => {
                return {
                    id: item.id + item.nextAiringEpisode?.episode!,
                    name: item.title?.userPreferred,
                    time: format(new Date(item.nextAiringEpisode?.airingAt! * 1000), "h:mm a"),
                    datetime: format(new Date(item.nextAiringEpisode?.airingAt! * 1000), "yyyy-MM-dd'T'HH:mm"),
                    href: `/entry?id=${item.id}`,
                    image: item.bannerImage ?? item.coverImage?.extraLarge ?? item.coverImage?.large ?? item.coverImage?.medium,
                    episode: item.nextAiringEpisode?.episode || 1,
                }
            })

            const pastMedia = missingEpisodes.filter((item) => !!item.episodeMetadata?.airDate && isSameDay(item.episodeMetadata?.airDate,
                day)).map((item) => {
                return {
                    id: item.baseMedia?.id! + item.fileMetadata?.episode!,
                    name: item.baseMedia?.title?.userPreferred,
                    time: "",
                    datetime: format(new Date(item.episodeMetadata?.airDate!), "yyyy-MM-dd'T'HH:mm"),
                    href: `/entry?id=${item.baseMedia?.id}`,
                    image: item.baseMedia?.bannerImage ?? item.baseMedia?.coverImage?.extraLarge ?? item.baseMedia?.coverImage?.large ?? item.baseMedia?.coverImage?.medium,
                    episode: item.episodeNumber || 1,
                }
            })

            daysArray.push({
                date: format(day, "yyyy-MM-dd"),
                isCurrentMonth: isSameMonth(day, currentDate),
                isToday: isToday(day),
                isSelected: false,
                events: [...upcomingMedia, ...pastMedia],
            })
            day = addDays(day, 1)
        }
        return daysArray
    }, [currentDate, media, missingEpisodes])

    if (media?.length === 0 && missingEpisodes?.length === 0) return null

    // const selectedDay = days.find((day) => day.isSelected)

    return (
        <>
            <div className="hidden lg:flex lg:h-full lg:flex-col rounded-md border">
                <header className="flex items-center justify-center py-4 px-6 gap-4 lg:flex-none border-b">
                    <IconButton icon={<AiOutlineArrowLeft />} onClick={goToPreviousMonth} rounded intent="gray-outline" size="sm" />
                    <h1
                        className={cn(
                            "text-lg font-semibold text-gray-100 text-center w-[200px]",
                            isSameMonth(currentDate, new Date()) && "text-brand-200",
                        )}
                    >
                        <time dateTime={format(currentDate, "yyyy-MM")}>
                            {format(currentDate, "MMMM yyyy")}
                        </time>
                    </h1>
                    <IconButton icon={<AiOutlineArrowRight />} onClick={goToNextMonth} rounded intent="gray-outline" size="sm" />
                </header>
                <div className="lg:flex lg:flex-auto lg:flex-col">
                    <div className="grid grid-cols-7 gap-px border-b bg-gray-900 text-center text-base font-semibold leading-6 text-gray-200 lg:flex-none">
                        <div className="bg-gray-950 py-2">
                            M<span className="sr-only sm:not-sr-only">on</span>
                        </div>
                        <div className="bg-gray-950 py-2">
                            T<span className="sr-only sm:not-sr-only">ue</span>
                        </div>
                        <div className="bg-gray-950 py-2">
                            W<span className="sr-only sm:not-sr-only">ed</span>
                        </div>
                        <div className="bg-gray-950 py-2">
                            T<span className="sr-only sm:not-sr-only">hu</span>
                        </div>
                        <div className="bg-gray-950 py-2">
                            F<span className="sr-only sm:not-sr-only">ri</span>
                        </div>
                        <div className="bg-gray-950 py-2">
                            S<span className="sr-only sm:not-sr-only">at</span>
                        </div>
                        <div className="bg-gray-950 py-2">
                            S<span className="sr-only sm:not-sr-only">un</span>
                        </div>
                    </div>
                    <div className="flex bg-gray-900 text-xs leading-6 text-gray-200 lg:flex-auto">
                        <div className="hidden w-full lg:grid lg:grid-cols-7 lg:grid-rows-6 lg:gap-px">
                            {days.map((day, index) => (
                                <div
                                    key={day.date + index}
                                    className={cn(
                                        day.isCurrentMonth ? "bg-gray-950" : "opacity-30",
                                        "relative py-2 px-3 lg:min-h-24 overflow-hidden",
                                        // "hover:bg-gray-900",
                                        "flex flex-col justify-between",
                                    )}
                                >
                                    {day.events[0] && (
                                        <>
                                            <div className="absolute top-0 left-0 z-[0] w-full h-full overflow-hidden opacity-30">
                                                <Image src={day.events[0]?.image || ""} alt="banner" fill className="object-cover" />
                                            </div>
                                            <div className="absolute left-0 bottom-0 z-[1] w-full h-full bg-gradient-to-t from-gray-950 to-transparent" />
                                        </>
                                    )}
                                    <time
                                        dateTime={day.date}
                                        className={
                                            day.isToday
                                                ? "z-[1] relative flex h-7 w-7 text-xs items-center justify-center rounded-full bg-brand font-semibold text-white"
                                                : "text-xs md:text-base"
                                        }
                                    >
                                        {day.date.split("-")?.pop()?.replace(/^0/, "")}
                                    </time>
                                    {day.events.length > 0 && (
                                        <ol className="mt-2 relative z-[1]">
                                            {day.events.slice(0, 4).map((event) => (
                                                <Tooltip
                                                    key={event.id}
                                                    trigger={
                                                        <li key={event.id}>
                                                            <Link className="group flex" href={event.href}>
                                                                <p className="flex-auto truncate font-medium text-gray-100 group-hover:text-gray-200">
                                                                    {event.name}
                                                                </p>
                                                                <time
                                                                    dateTime={event.datetime}
                                                                    className="ml-3 hidden flex-none text-[--muted] group-hover:text-gray-200 xl:block"
                                                                >
                                                                    {event.time}
                                                                </time>
                                                            </Link>
                                                        </li>
                                                    }
                                                >
                                                    Episode {event.episode}
                                                </Tooltip>
                                            ))}
                                            {day.events.length > 2 && <li className="text-[--muted]">+ {day.events.length - 2} more</li>}
                                        </ol>
                                    )}
                                </div>
                            ))}
                        </div>
                        <div className="isolate grid w-full grid-cols-7 grid-rows-6 gap-px lg:hidden">
                            {days.map((day) => (
                                <button
                                    key={day.date}
                                    type="button"
                                    className={cn(
                                        day.isCurrentMonth ? "bg-gray-950" : "bg-gray-900",
                                        (day.isSelected || day.isToday) && "font-semibold",
                                        day.isSelected && "text-white",
                                        !day.isSelected && day.isToday && "text-brand",
                                        !day.isSelected && day.isCurrentMonth && !day.isToday && "text-gray-100",
                                        !day.isSelected && !day.isCurrentMonth && !day.isToday && "text-gray-500",
                                        "flex h-14 flex-col py-2 px-3 hover:bg-gray-800 focus:z-10",
                                    )}
                                >
                                    <time
                                        dateTime={day.date}
                                        className={cn(
                                            day.isSelected && "flex h-6 w-6 items-center justify-center rounded-full",
                                            day.isSelected && day.isToday && "bg-brand",
                                            day.isSelected && !day.isToday && "bg-gray-900",
                                            "ml-auto",
                                        )}
                                    >
                                        {day.date.split("-")?.pop()?.replace(/^0/, "")}
                                    </time>
                                    <span className="sr-only">{day.events.length} events</span>
                                    {day.events.length > 0 && (
                                        <span className="-mx-0.5 mt-auto flex flex-wrap-reverse">
                                            {day.events.map((event) => (
                                                <span key={event.id} className="mx-0.5 mb-1 h-1.5 w-1.5 rounded-full bg-gray-400" />
                                            ))}
                                        </span>
                                    )}
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}
