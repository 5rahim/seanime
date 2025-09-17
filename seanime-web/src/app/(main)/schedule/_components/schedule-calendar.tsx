import { AL_MediaListStatus, Anime_Episode } from "@/api/generated/types"
import { useGetAnimeCollectionSchedule } from "@/api/hooks/anime_collection.hooks"
import { SeaLink } from "@/components/shared/sea-link"
import { IconButton } from "@/components/ui/button"
import { CheckboxGroup } from "@/components/ui/checkbox"
import { cn } from "@/components/ui/core/styling"
import { Popover } from "@/components/ui/popover"
import { RadioGroup } from "@/components/ui/radio-group"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { addMonths, Day, endOfMonth, endOfWeek, format, isSameMonth, isToday, startOfMonth, startOfWeek, subMonths } from "date-fns"
import { addDays } from "date-fns/addDays"
import { useImmerAtom } from "jotai-immer"
import { useAtom, useAtomValue } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import { sortBy } from "lodash"
import Image from "next/image"
import React, { Fragment } from "react"
import { AiOutlineArrowLeft, AiOutlineArrowRight } from "react-icons/ai"
import { BiCog } from "react-icons/bi"
import { FaCheck, FaFlag } from "react-icons/fa6"
import { __anilist_userAnimeListDataAtom } from "../../_atoms/anilist.atoms"

type CalendarParams = {
    indicateWatchedEpisodes: boolean
    listStatuses: AL_MediaListStatus[]
}

const MAX_EVENT_COUNT = 4

export const weekStartsOnAtom = atomWithStorage("sea-calendar-week-starts-on", 1)
export const calendarDisableAnimations = atomWithStorage("sea-calendar-disable-animations", false)
export const calendarParamsAtom = atomWithStorage("sea-release-calendar-params", {
    indicateWatchedEpisodes: true,
    listStatuses: ["PLANNING", "CURRENT", "COMPLETED", "PAUSED"] as AL_MediaListStatus[],
})

type ScheduleCalendarProps = {
    children?: React.ReactNode
    missingEpisodes: Anime_Episode[]
}

export function ScheduleCalendar(props: ScheduleCalendarProps) {

    const {
        children,
        missingEpisodes,
        ...rest
    } = props

    const anilistListData = useAtomValue(__anilist_userAnimeListDataAtom)

    const { data: schedule } = useGetAnimeCollectionSchedule()

    // State for the current displayed month
    const [currentDate, setCurrentDate] = React.useState(new Date())

    const [calendarParams, setCalendarParams] = useImmerAtom(calendarParamsAtom)
    const [animationsDisabled, setAnimationDisabled] = useAtom(calendarDisableAnimations)

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


    function isStatusIncluded(mediaId: number) {
        const entry = anilistListData[String(mediaId)]
        if (!entry || !entry.status) return false
        return calendarParams.listStatuses.includes(entry.status)
    }

    function isEpisodeWatched(mediaId: number, episodeNumber: number) {
        const entry = anilistListData[String(mediaId)]
        if (!entry || !entry.progress) return false
        return entry.progress >= episodeNumber
    }

    const days = React.useMemo(() => {
        const startOfCurrentMonth = startOfMonth(currentDate)
        const endOfCurrentMonth = endOfMonth(currentDate)

        const startOfCalendar = startOfWeek(startOfCurrentMonth, { weekStartsOn: weekStartsOn as Day })
        const endOfCalendar = endOfWeek(endOfCurrentMonth, { weekStartsOn: weekStartsOn as Day })

        const daysArray = []
        let day = startOfCalendar

        while (day <= endOfCalendar) {
            let events = schedule?.filter(item => isSameDayUtc(new Date(item.dateTime!), day) && isStatusIncluded(item.mediaId))?.map(item => {
                return {
                    id: String(item.mediaId) + "-" + String(item.episodeNumber) + "-" + String(item.dateTime),
                    name: item.title,
                    time: item.time,
                    datetime: item.dateTime!,
                    href: `/entry?id=${item.mediaId}`,
                    image: item.image,
                    episode: item.episodeNumber || 1,
                    isSeasonFinale: item.isSeasonFinale && !item.isMovie,
                    isMovie: item.isMovie,
                    isWatched: isEpisodeWatched(item.mediaId, item.episodeNumber),
                }
            }) ?? []
            events = sortBy(events, (e) => e.episode)
            events = sortBy(events, (e) => e.datetime)

            daysArray.push({
                date: format(day, "yyyy-MM-dd"),
                isCurrentMonth: isSameMonth(day, currentDate),
                isToday: isToday(day),
                isSelected: false,
                events: events,
            })
            day = addDays(day, 1)
        }
        return daysArray
    }, [currentDate, missingEpisodes, weekStartsOn, schedule, calendarParams, anilistListData])


    return (
        <>
            <div className="flex h-full flex-col rounded-[--radius-md] border">
                <header className="relative flex items-center justify-center py-3 px-4 lg:py-4 lg:px-6 gap-3 lg:gap-4 flex-none rounded-tr-[--radius-md] rounded-tl-[--radius-md] border-b bg-[--background]">
                    <IconButton icon={<AiOutlineArrowLeft />} onClick={goToPreviousMonth} rounded intent="gray-outline" size="sm" />
                    <h1
                        className={cn(
                            "text-base lg:text-lg font-semibold text-[--muted] text-center flex-1 min-w-0",
                            isSameMonth(currentDate, new Date()) && "text-gray-100",
                        )}
                    >
                        <time dateTime={format(currentDate, "yyyy-MM")}>
                            <span className="hidden lg:inline">{format(currentDate, "MMMM yyyy")}</span>
                            <span className="lg:hidden">{format(currentDate, "MMM yyyy")}</span>
                        </time>
                    </h1>
                    <IconButton icon={<AiOutlineArrowRight />} onClick={goToNextMonth} rounded intent="gray-outline" size="sm" />

                    <Popover
                        trigger={<IconButton icon={<BiCog />} intent="gray-basic" size="sm" />}
                        className="w-[300px] lg:w-[400px] space-y-2"
                    >
                        <RadioGroup
                            label="Week starts on" options={[
                            { label: "Monday", value: "1" },
                            { label: "Sunday", value: "0" },
                        ]} value={String(weekStartsOn)} onValueChange={v => setWeekStartsOn(Number(v))}
                        />
                        <Separator />
                        <CheckboxGroup
                            label="Status" options={[
                            { label: "Watching", value: "CURRENT" },
                            { label: "Planning", value: "PLANNING" },
                            { label: "Completed", value: "COMPLETED" },
                            { label: "Paused", value: "PAUSED" },
                        ]} value={calendarParams.listStatuses} onValueChange={v => setCalendarParams(draft => {
                            draft.listStatuses = v as AL_MediaListStatus[]
                            return
                        })}
                            stackClass="grid grid-cols-2 gap-0 items-center !space-y-0"
                        />
                        <Separator />
                        <Switch
                            label="Indicate watched episodes"
                            side="right"
                            value={calendarParams.indicateWatchedEpisodes}
                            onValueChange={v => setCalendarParams(draft => {
                                draft.indicateWatchedEpisodes = v
                                return
                            })}
                        />
                        <Separator />
                        <Switch
                            label="Disable image transitions"
                            side="right"
                            value={animationsDisabled}
                            onValueChange={v => setAnimationDisabled(v)}
                        />
                    </Popover>
                </header>
                <div className="flex flex-auto flex-col rounded-br-[--radius-md] rounded-bl-[--radius-md] overflow-hidden">
                    <div className="hidden lg:grid grid-cols-7 gap-px border-b bg-[--background] text-center text-base font-semibold leading-6 text-gray-200 flex-none">
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

                    <div className="lg:hidden flex-auto bg-[--background] overflow-y-auto">
                        <MobileCalendarList days={days} />
                    </div>

                    <div className="hidden lg:flex bg-[--background] text-xs leading-6 text-gray-200 flex-auto">
                        <div className="w-full grid grid-cols-7 grid-rows-6 gap-2 p-2">
                            {days.map((day, index) => (
                                <CalendarDay
                                    key={index}
                                    day={day}
                                    index={index}
                                />
                            ))}
                        </div>
                    </div>
                </div>
            </div>
        </>
    )
}

type CalendarEvent = {
    id: string
    name: string
    time: string
    datetime: string
    href: string
    image: string
    episode: number
    isSeasonFinale: boolean
    isMovie: boolean
    isWatched: boolean
}

interface MobileCalendarListProps {
    days: any[]
}

function MobileCalendarList({ days }: MobileCalendarListProps) {
    const calendarParams = useAtomValue(calendarParamsAtom)

    // Filter days to only show those with events or today
    const relevantDays = days.filter(day =>
        day.events.length > 0 || day.isToday,
    )

    if (relevantDays.length === 0) {
        return (
            <div className="p-6 text-center text-[--muted]">
                <p>No scheduled episodes for this month</p>
            </div>
        )
    }

    return (
        <div className="divide-y divide-gray-800">
            {relevantDays.map((day, index) => (
                <MobileDayItem
                    key={day.date}
                    day={day}
                    calendarParams={calendarParams}
                />
            ))}
        </div>
    )
}

interface MobileDayItemProps {
    day: any
    calendarParams: CalendarParams
}

function MobileDayItem({ day, calendarParams }: MobileDayItemProps) {
    const dayName = format(new Date(day.date), "EEEE")
    const dayNumber = day.date.split("-")?.pop()?.replace(/^0/, "")
    const monthDay = format(new Date(day.date), "MMM d")

    return (
        <div className="p-4">
            <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-3">
                    <div
                        className={cn(
                            "flex h-8 w-8 lg:h-10 lg:w-10 items-center justify-center rounded-full font-bold text-sm lg:text-base",
                            day.isToday
                                ? "bg-brand text-white"
                                : "bg-gray-800 text-gray-300",
                        )}
                    >
                        {dayNumber}
                    </div>
                    <div>
                        <h3
                            className={cn(
                                "font-semibold",
                                day.isToday ? "text-brand" : "text-gray-200",
                            )}
                        >
                            {dayName}
                        </h3>
                        <p className="text-sm text-[--muted]">{monthDay}</p>
                    </div>
                </div>
                {day.events.length > 0 && (
                    <div className="text-xs text-[--muted] bg-gray-800 px-2 py-1 rounded-full">
                        {day.events.length} episode{day.events.length !== 1 ? "s" : ""}
                    </div>
                )}
            </div>

            {day.events.length > 0 && (
                <div className="space-y-3 ml-0 lg:ml-13">
                    {day.events.map((event: CalendarEvent) => (
                        <MobileEventItem
                            key={event.id}
                            event={event}
                            calendarParams={calendarParams}
                        />
                    ))}
                </div>
            )}

            {day.isToday && day.events.length === 0 && (
                <div className="ml-0 lg:ml-13 text-sm text-[--muted] italic">
                    No episodes scheduled for today
                </div>
            )}
        </div>
    )
}

interface MobileEventItemProps {
    event: CalendarEvent
    calendarParams: any
}

function MobileEventItem({ event, calendarParams }: MobileEventItemProps) {
    return (
        <SeaLink href={event.href} className="block">
            <div className="flex items-start gap-2 lg:gap-3 p-2 lg:p-3 rounded-lg bg-gray-900/50 hover:bg-gray-800/50 transition-colors">
                <div className="relative w-10 h-14 lg:w-12 lg:h-16 rounded overflow-hidden flex-shrink-0">
                    <Image
                        src={event.image || ""}
                        alt={event.name}
                        fill
                        className="object-cover"
                    />
                </div>

                <div className="flex-1 min-w-0">
                    <div className="flex items-start justify-between gap-2">
                        <p
                            className={cn(
                                "font-medium text-md text-gray-100 line-clamp-2",
                                event.isWatched && calendarParams.indicateWatchedEpisodes && "text-[--muted]",
                            )}
                        >
                            {event.name}
                        </p>
                        <div className="flex items-center gap-1 flex-shrink-0">
                            {event.isSeasonFinale && !event.isWatched && (
                                <FaFlag className="size-3 text-[--blue]" />
                            )}
                            {event.isWatched && calendarParams.indicateWatchedEpisodes && (
                                <FaCheck className="size-3 text-[--muted]" />
                            )}
                        </div>
                    </div>

                    <div className="flex items-center gap-4 mt-2 text-sm text-[--muted]">
                        <span className="font-medium">Episode {event.episode}</span>
                        {event.time && <span>• {event.time}</span>}
                        {event.isSeasonFinale && (
                            <span className="text-[--blue] font-medium">• Finale</span>
                        )}
                    </div>
                </div>
            </div>
        </SeaLink>
    )
}

interface CalendarDayBackgroundProps {
    events: CalendarEvent[]
    isToday: boolean
    hoveredEventId: string | null
}

function CalendarDayBackground({ events, isToday, hoveredEventId }: CalendarDayBackgroundProps) {

    const [focusedEventIndex, setFocusedEventIndex] = React.useState<number | null>(null)
    const transitionDisabled = useAtomValue(calendarDisableAnimations)
    React.useEffect(() => {
        if (transitionDisabled) {
            setFocusedEventIndex(0)
            return
        }
        // carousel
        const interval = setInterval(() => {
            setFocusedEventIndex(prev => {
                if (prev === null) return 0
                if (prev === events.length - 1) return 0
                return prev + 1
            })
        }, 5000)
        return () => clearInterval(interval)
    }, [events, transitionDisabled])

    const displayedEvent = React.useMemo(() => {
        if (hoveredEventId) {
            const hoveredEvent = events.find(e => e.id === hoveredEventId)
            if (hoveredEvent) return hoveredEvent
        } else if (focusedEventIndex !== null && focusedEventIndex < events.length) {
            return events[focusedEventIndex]
        }
        return events[0]
    }, [hoveredEventId, events, focusedEventIndex])

    return (
        <>
            <div
                className={cn(
                    "absolute top-0 left-0 z-[0] w-full h-full overflow-hidden rounded-md transition-all duration-500 ease-out",
                    isToday ? "opacity-80" : "opacity-20 group-hover:opacity-30",
                )}
            >
                <Image
                    src={displayedEvent?.image || ""}
                    alt="banner"
                    fill
                    className="object-cover transition-all duration-500 ease-out transform"
                    key={displayedEvent?.id}
                />
            </div>
            <div
                className={cn(
                    "absolute left-0 bottom-0 z-[1] w-full h-full bg-gradient-to-t from-gray-950/100 via-gray-950/80 via-40% to-transparent transition-all duration-300",
                    isToday && "from-gray-950/90 via-gray-950/80 via-40%",
                )}
            />
        </>
    )
}

interface CalendarEventListProps {
    events: CalendarEvent[]
    onEventHover: (eventId: string | null) => void
}

function CalendarEventList({ events, onEventHover }: CalendarEventListProps) {
    const handleEventMouseEnter = (eventId: string) => {
        onEventHover(eventId)
    }

    const handleEventMouseLeave = () => {
        onEventHover(null)
    }

    const calendarParams = useAtomValue(calendarParamsAtom)

    return (
        <ol className="mt-1 sm:mt-2 relative z-[1] space-y-0.5 sm:space-y-1">
            {events.slice(0, MAX_EVENT_COUNT).map((event) => (
                <li
                    key={event.id}
                    onMouseEnter={() => handleEventMouseEnter(event.id)}
                    onMouseLeave={handleEventMouseLeave}
                >
                    <SeaLink className="group flex" href={event.href}>
                        <div className="flex-auto truncate">
                            <p
                                className={cn(
                                    "truncate font-medium text-gray-100 flex items-center gap-1",
                                    "text-xs lg:text-sm",
                                    event.isWatched && calendarParams.indicateWatchedEpisodes ? "text-[--muted]" : "group-hover:text-gray-200",
                                )}
                            >
                                {event.isSeasonFinale && !event.isWatched &&
                                    <FaFlag className="size-2 lg:size-3 text-[--blue] flex-none group-hover:scale-[1.15] transition-transform duration-300" />}
                                {event.isWatched && calendarParams.indicateWatchedEpisodes &&
                                    <FaCheck className="size-2 lg:size-3 text-[--muted] flex-none group-hover:scale-[1.15] transition-transform duration-300" />}
                                <span className="truncate">
                                    {event.name.length > 20 ? event.name.slice(0, 17) + "..." : event.name}
                                </span>
                            </p>
                            <p className="text-xs text-[--muted] lg:hidden">
                                Ep. {event.episode}
                                {event.time && <span className="ml-1">• {event.time}</span>}
                            </p>
                        </div>
                        <time
                            dateTime={event.datetime}
                            className="ml-3 hidden flex-none text-[--muted] group-hover:text-gray-200 lg:flex items-center"
                        >
                            <span className="mr-1 text-sm group-hover:text-[--foreground] font-semibold ">
                                Ep. {event.episode}
                            </span>
                        </time>
                    </SeaLink>
                </li>
            ))}
            {events.length > MAX_EVENT_COUNT && (
                <Popover
                    className="w-[280px] lg:w-full max-w-sm lg:max-w-sm"
                    trigger={
                        <li className="text-[--muted] cursor-pointer text-xs lg:text-sm py-1">+ {events.length - MAX_EVENT_COUNT} more</li>
                    }
                >
                    <ol className="text-sm max-w-full block space-y-2">
                        {events.slice(MAX_EVENT_COUNT).map((event) => (
                            <li key={event.id}>
                                <SeaLink className="group flex gap-2" href={event.href}>
                                    <p
                                        className={cn("flex-auto truncate font-medium text-gray-100 flex items-center gap-2",
                                            event.isWatched && calendarParams.indicateWatchedEpisodes
                                                ? "text-[--muted]"
                                                : "group-hover:text-gray-200")}
                                    >
                                        {event.isSeasonFinale && !event.isWatched &&
                                            <FaFlag className="size-3 text-[--blue] flex-none group-hover:scale-[1.15] transition-transform duration-300" />}
                                        {event.isWatched && calendarParams.indicateWatchedEpisodes &&
                                            <FaCheck className="size-3 text-[--muted] flex-none group-hover:scale-[1.15] transition-transform duration-300" />}
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
                </Popover>
            )}
        </ol>
    )
}

function CalendarDay({ day, index }: { day: any, index: number }) {
    const [hoveredEventId, setHoveredEventId] = React.useState<string | null>(null)

    const hoveredEvent = React.useMemo(() => {
        if (hoveredEventId) {
            return day.events.find((e: CalendarEvent) => e.id === hoveredEventId)
        }
        return null
    }, [hoveredEventId, day.events])

    return (
        <div
            key={index}
            className={cn(
                day.isCurrentMonth ? "bg-[--background]" : "opacity-20",
                "relative py-1 px-1 sm:py-2 sm:px-3 h-24 sm:h-32 lg:h-40 rounded-md",
                "flex flex-col justify-between group",
            )}
        >
            {day.events[0] && (
                <CalendarDayBackground
                    events={day.events}
                    isToday={day.isToday}
                    hoveredEventId={hoveredEventId}
                />
            )}

            <div className="absolute -top-2 left-10 right-1 z-[5] pointer-events-none hidden lg:block">
                <div
                    className={cn(
                        "transition-all duration-300 ease-out",
                        hoveredEvent ? "opacity-100 transform translate-y-0" : "opacity-0 transform -translate-y-2",
                    )}
                >
                    {hoveredEvent && (
                        <div className="bg-gray-900/70 backdrop-blur-sm rounded-md px-2 py-1.5 border">
                            <p className="text-xs font-medium text-gray-100 line-clamp-2 leading-tight">
                                <span className="text-[--muted] font-normal">{hoveredEvent.name.slice(0, 28) + (hoveredEvent.name.length > 28
                                    ? "..."
                                    : "")}</span>
                                {hoveredEvent.isSeasonFinale && <span className="text-[--blue] ml-1">Finale</span>}
                                <span className="ml-1"> Ep. {hoveredEvent.episode}</span>
                                {hoveredEvent.time && <span className="ml-1">- {hoveredEvent.time}</span>}
                            </p>
                        </div>
                    )}
                </div>
            </div>

            <time
                dateTime={day.date}
                className={
                    day.isToday
                        ? "z-[1] relative flex h-5 w-5 sm:h-6 sm:w-6 lg:h-7 lg:w-7 text-sm sm:text-base lg:text-lg items-center justify-center rounded-full bg-brand font-bold group-hover:rotate-12 transition-transform duration-300 ease-out text-white"
                        : "text-xs sm:text-sm lg:text-base group-hover:text-white group-hover:font-bold transition-transform duration-300 ease-out"
                }
            >
                {day.date.split("-")?.pop()?.replace(/^0/, "")}
            </time>
            {day.events.length > 0 && (
                <CalendarEventList
                    events={day.events}
                    onEventHover={setHoveredEventId}
                />
            )}
        </div>
    )
}
