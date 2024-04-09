import { cn } from "@/components/ui/core/styling"

export const tabsTriggerClass = cn(
    "text-base px-6 rounded-md w-fit md:w-full border-none data-[state=active]:bg-[--subtle] data-[state=active]:text-white dark:hover:text-white")

export const tabsListClass = cn("w-full flex flex-wrap md:flex-nowrap h-fit md:h-12")

export const monochromeCheckboxClass = {
    className: "hidden",
    labelClass: cn(
        "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-3 w-full",
        "hover:bg-[--subtle] dark:bg-gray-900",
        "data-[checked=true]:bg-white dark:data-[checked=true]:bg-gray-950",
        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "border border-transparent data-[checked=true]:border-[--gray] data-[checked=true]:ring-offset-0",
        "w-fit",
    ),
}
export const primaryPillCheckboxClass = {
    className: "hidden",
    labelClass: cn(
        "text-gray-300 data-[checked=true]:text-white hover:!bg-[--highlight]",
        "items-start cursor-pointer transition border-transparent rounded-[--radius] py-1.5 px-3 w-full",
        "hover:bg-[--subtle] dark:bg-gray-900",
        "data-[checked=true]:bg-white dark:data-[checked=true]:bg-gray-950",
        "focus:ring-2 ring-transparent dark:ring-transparent outline-none ring-offset-1 ring-offset-[--background] focus-within:ring-2 transition",
        "border border-transparent data-[checked=true]:border-[--brand] data-[checked=true]:ring-offset-0",
        "w-fit",
    ),
}
