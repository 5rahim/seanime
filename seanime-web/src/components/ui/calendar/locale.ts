import enUS from "date-fns/locale/en-US"
import fr from "date-fns/locale/fr"

export const getDateLocaleLibrary = (locale: string) => locale.startsWith("fr") ? fr : enUS

/**
 *
 * @param weekDays - from useCalendarGrid()
 * @param locale
 */
export const getShortenedWeekDays = (weekDays: string[], locale: string) => {
    const [first, ...r] = weekDays
    if (locale.startsWith("fr")) {
        return ["L", "M", "M", "J", "V", "S", "D"]
    }
    return [...r!, first!]
}
