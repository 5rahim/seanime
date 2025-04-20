import { AL_BaseAnime, Anime_Episode } from "@/api/generated/types"
import { SeaLink } from "@/components/shared/sea-link"
import { IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { Modal } from "@/components/ui/modal"
import { Popover } from "@/components/ui/popover"
import { RadioGroup } from "@/components/ui/radio-group"
import { Tooltip } from "@/components/ui/tooltip"
import { addMonths, Day, endOfMonth, endOfWeek, format, isSameMonth, isToday, startOfMonth, startOfWeek, subMonths } from "date-fns"
import { addDays } from "date-fns/addDays"
import { isSameDay } from "date-fns/isSameDay"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import Image from "next/image"
import React, { Fragment } from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight } from "react-icons/ai"
import { BiCog } from "react-icons/bi"

type WeekCalendarProps = {
    children?: React.ReactNode
    media: AL_BaseAnime[]
    missingEpisodes: Anime_Episode[]
}

const MAX_EVENT_COUNT = 5

export const weekStartsOnAtom = atomWithStorage("sea-calendar-week-starts-on", 1)

export function MonthCalendar(props: WeekCalendarProps) {

    const {
        children,
        media,
        missingEpisodes,
        ...rest
    } = props

    // State for the current displayed month
    const [currentDate, setCurrentDate] = React.useState(new Date())

    const [weekStartsOn, setWeekStartsOn] = useAtom(weekStartsOnAtom)

    // Function to go to the previous month
    const goToPreviousMonth = () => {
        setCurrentDate(prevDate => subMonths(prevDate, 1))
    }

    // Function to go to the next month
    const goToNextMonth = () => {
        setCurrentDate(prevDate => addMonths(prevDate, 1))
    }

    const isSameDayUtc = (dateLeft: Date, dateRight: Date) => {
        return (
            dateLeft.getFullYear() === dateRight.getFullYear() &&
            dateLeft.getMonth() === dateRight.getMonth() &&
            dateLeft.getDate() === dateRight.getDate()
        )
    }


    const days = React.useMemo(() => {
        const startOfCurrentMonth = startOfMonth(currentDate)
        const endOfCurrentMonth = endOfMonth(currentDate)

        const startOfCalendar = startOfWeek(startOfCurrentMonth, { weekStartsOn: weekStartsOn as Day })
        const endOfCalendar = endOfWeek(endOfCurrentMonth, { weekStartsOn: weekStartsOn as Day })

        const daysArray = []
        let day = startOfCalendar

        while (day <= endOfCalendar) {
            const upcomingMedia = media.filter((item) => !!item.nextAiringEpisode?.airingAt && isSameDay(new Date(item.nextAiringEpisode?.airingAt * 1000),
                day)).map((item) => {
                return {
                    id: String(item.id) + String(item.nextAiringEpisode?.episode!),
                    name: item.title?.userPreferred,
                    time: format(new Date(item.nextAiringEpisode?.airingAt! * 1000), "h:mm a"),
                    datetime: format(new Date(item.nextAiringEpisode?.airingAt! * 1000), "yyyy-MM-dd'T'HH:mm"),
                    href: `/entry?id=${item.id}`,
                    image: item.bannerImage ?? item.coverImage?.extraLarge ?? item.coverImage?.large ?? item.coverImage?.medium,
                    episode: item.nextAiringEpisode?.episode || 1,
                }
            })

            const pastMedia = missingEpisodes.filter((item) => !!item.episodeMetadata?.airDate && isSameDayUtc(new Date(item.episodeMetadata?.airDate),
                day)).map((item) => {
                return {
                    id: String(item.baseAnime?.id!) + String(item.fileMetadata?.episode!),
                    name: item.baseAnime?.title?.userPreferred,
                    time: "Aired",
                    datetime: item.episodeMetadata?.airDate,
                    href: `/entry?id=${item.baseAnime?.id}`,
                    image: item.baseAnime?.bannerImage ?? item.baseAnime?.coverImage?.extraLarge ?? item.baseAnime?.coverImage?.large ?? item.baseAnime?.coverImage?.medium,
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
    }, [currentDate, media, missingEpisodes, weekStartsOn])

    if (media?.length === 0 && missingEpisodes?.length === 0) return null

    return (
        <>
            <div className="hidden lg:flex lg:h-full lg:flex-col rounded-[--radius-md] border">
                <header className="relative flex items-center justify-center py-4 px-6 gap-4 lg:flex-none rounded-tr-[--radius-md] rounded-tl-[--radius-md] border-b bg-[--background]">
                    <IconButton icon={<AiOutlineArrowLeft />} onClick={goToPreviousMonth} rounded intent="gray-outline" size="sm" />
                    <h1
                        className={cn(
                            "text-lg font-semibold text-[--muted] text-center w-[200px]",
                            isSameMonth(currentDate, new Date()) && "text-gray-100",
                        )}
                    >
                        <time dateTime={format(currentDate, "yyyy-MM")}>
                            {format(currentDate, "MMMM yyyy")}
                        </time>
                    </h1>
                    <IconButton icon={<AiOutlineArrowRight />} onClick={goToNextMonth} rounded intent="gray-outline" size="sm" />

                    <Modal
                        title="Calendar Settings"
                        trigger={<IconButton icon={<BiCog />} intent="gray-basic" className="absolute right-3 top-4" size="sm" />}
                    >
                        <RadioGroup
                            label="Week starts on" options={[
                            { label: "Monday", value: "1" },
                            { label: "Sunday", value: "0" },
                        ]} value={String(weekStartsOn)} onValueChange={v => setWeekStartsOn(Number(v))}
                        />
                    </Modal>
                </header>
                <div className="lg:flex lg:flex-auto lg:flex-col rounded-br-[--radius-md] rounded-bl-[--radius-md] overflow-hidden">
                    <div className="grid grid-cols-7 gap-px border-b bg-[--background] text-center text-base font-semibold leading-6 text-gray-200 lg:flex-none">
                        {weekStartsOn === 0 && <div className="py-2">
                            S<span className="sr-only sm:not-sr-only">un</span>
                        </div>}
                        <div className="py-2">
                            M<span className="sr-only sm:not-sr-only">on</span>
                        </div>
                        <div className="py-2">
                            T<span className="sr-only sm:not-sr-only">ue</span>
                        </div>
                        <div className="py-2">
                            W<span className="sr-only sm:not-sr-only">ed</span>
                        </div>
                        <div className="py-2">
                            T<span className="sr-only sm:not-sr-only">hu</span>
                        </div>
                        <div className="py-2">
                            F<span className="sr-only sm:not-sr-only">ri</span>
                        </div>
                        <div className="py-2">
                            S<span className="sr-only sm:not-sr-only">at</span>
                        </div>
                        {weekStartsOn === 1 && <div className="py-2">
                            S<span className="sr-only sm:not-sr-only">un</span>
                        </div>}
                    </div>
                    <div className="flex bg-gray-950 text-xs leading-6 text-gray-200 lg:flex-auto">
                        <div className="hidden w-full lg:grid lg:grid-cols-7 lg:grid-rows-6 lg:gap-px">
                            {days.map((day, index) => (
                                <div
                                    key={index}
                                    className={cn(
                                        day.isCurrentMonth ? "bg-[--background]" : "opacity-30",
                                        "relative py-2 px-3 lg:min-h-24 overflow-hidden",
                                        // "hover:bg-gray-900",
                                        "flex flex-col justify-between",
                                    )}
                                >
                                    {day.events[0] && (
                                        <>
                                            <div
                                                className={`absolute top-0 left-0 z-[0] w-full h-full overflow-hidden ${
                                                    day.isToday ? "opacity-80" : "opacity-20"
                                                }`}
                                            >
                                                <Image src={day.events[0]?.image || ""} alt="banner" fill className="object-cover" />
                                            </div>
                                            <div
                                                className={cn(
                                                    "absolute left-0 bottom-0 z-[1] w-full h-full bg-gradient-to-t from-gray-950/100 via-gray-950/80 via-40% to-transparent",
                                                    day.isToday && "from-gray-950/90 via-gray-950/80 via-40%",
                                                )}
                                            />
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
                                            {day.events.slice(0, MAX_EVENT_COUNT).map((event) => (
                                                <Tooltip
                                                    key={event.id}
                                                    trigger={
                                                        <li key={event.id}>
                                                            <SeaLink className="group flex" href={event.href}>
                                                                <p className="flex-auto truncate font-medium text-gray-100 group-hover:text-gray-200">
                                                                    {event.name}
                                                                </p>
                                                                <time
                                                                    dateTime={event.datetime}
                                                                    className="ml-3 hidden flex-none text-[--muted] group-hover:text-gray-200 xl:block"
                                                                >
                                                                    {event.time}
                                                                </time>
                                                            </SeaLink>
                                                        </li>
                                                    }
                                                >
                                                    Episode {event.episode}
                                                </Tooltip>
                                            ))}
                                            {day.events.length > MAX_EVENT_COUNT && <Popover
                                                className="w-full max-w-sm lg:max-w-sm"
                                                trigger={
                                                    <li className="text-[--muted]">+ {day.events.length - MAX_EVENT_COUNT} more</li>}
                                            >
                                                <ol className="text-sm max-w-full block">
                                                    {day.events.map((event) => (
                                                        <li key={event.id}>
                                                            <SeaLink className="group flex gap-2" href={event.href}>
                                                                <p className="flex-1 truncate font-medium">
                                                                    {event.name}
                                                                </p>
                                                                <p className="flex-none">
                                                                    Ep. {event.episode}
                                                                </p>
                                                                <time
                                                                    dateTime={event.datetime}
                                                                    className="ml-3 hidden flex-none text-[--muted] group-hover:text-gray-200 xl:block"
                                                                >
                                                                    {event.time}
                                                                </time>
                                                            </SeaLink>
                                                        </li>
                                                    ))}
                                                </ol>
                                            </Popover>}
                                        </ol>
                                    )}
                                </div>
                            ))}
                        </div>
                        <div className="isolate grid w-full grid-cols-7 grid-rows-6 gap-px lg:hidden">
                            {days.map((day, index) => (
                                <button
                                    key={index}
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
