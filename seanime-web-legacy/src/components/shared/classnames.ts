import { cn } from "@/components/ui/core/styling"

export const tabsTriggerClass = cn(
    "text-base px-6 rounded-[--radius-md] w-fit md:w-full border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white")

export const tabsListClass = cn("w-full flex flex-wrap md:flex-nowrap h-fit md:h-12")

export const monochromeCheckboxClasses = {
    className: "hidden",
    labelClass: cn(
        "items-start cursor-pointer transition border-transparent rounded-xl py-1.5 px-3 w-full",
        "hover:bg-[--subtle] dark:bg-gray-900",
        "data-[checked=true]:bg-white dark:data-[checked=true]:bg-gray-950 dark:hover:bg-[--subtle] dark:hover:data-[checked=true]:bg-[--subtle]",
        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "border border-transparent data-[checked=true]:border-gray-500 data-[checked=true]:ring-offset-2",
        "data-[checked=true]:text-white active:translate-y-0.5 transition-all",
        "w-fit",
    ),
}
export const primaryPillCheckboxClasses = {
    className: "hidden",
    labelClass: cn(
        "text-gray-300 data-[checked=true]:text-white hover:!bg-[--highlight]",
        "items-start cursor-pointer transition border-transparent rounded-xl py-1.5 px-3 w-full",
        "hover:bg-[--subtle] dark:bg-gray-900",
        "data-[checked=true]:bg-white dark:data-[checked=true]:bg-gray-950",
        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "border border-transparent data-[checked=true]:border-[--brand] data-[checked=true]:ring-offset-0",
        "w-fit",
    ),
}

export const episodeCardCarouselItemClass = (smaller: boolean) => {
    return cn(
        !smaller && "md:basis-1/2 lg:basis-1/2 2xl:basis-1/3 min-[2000px]:basis-1/4",
        smaller && "md:basis-1/2 lg:basis-1/3 2xl:basis-1/4 min-[2000px]:basis-1/5",
    )
}
