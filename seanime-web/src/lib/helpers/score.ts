import { cn } from "@/components/ui/core/styling"

export function getScoreColor(score: number, kind: "audience" | "user"): string {
    if (score < 40) { // 0-39
        return cn(
            // kind === "audience" && "bg-red-800 bg-opacity-70",
            // light: dark token shade (readable on the light-grey pill); dark: original audienceScore hue preserved
            kind === "audience" && "text-[--red] dark:text-audienceScore-300 bg-black bg-opacity-20",
            kind === "user" && "bg-red-800 bg-opacity-90",
            // "text-red-200",
        )
    }
    if (score < 60) { // 30-59
        return cn(
            // kind === "audience" && "bg-amber-800 bg-opacity-70",
            kind === "audience" && "text-[--amber] dark:text-audienceScore-500 bg-black bg-opacity-20",
            kind === "user" && "bg-amber-800 bg-opacity-90",
            // "text-amber-200",
        )
    }
    if (score < 70) { // 60-69
        return cn(
            // kind === "audience" && "bg-lime-800 bg-opacity-70",
            kind === "audience" && "text-[--lime] dark:text-audienceScore-600 bg-black bg-opacity-20",
            kind === "user" && "bg-lime-800 bg-opacity-90",
            // "text-lime-200",
        )
    }
    // if (score < 80) { // 70-79
    //     return cn(
    //         // kind === "audience" && "bg-emerald-800 bg-opacity-70",
    //         // "text-emerald-100",
    //         kind === "audience" && "text-emerald-300 bg-black bg-opacity-20",
    //         kind === "user" && "bg-emerald-800 bg-opacity-90 text-white",
    //     )
    // }
    if (score < 82) {
        return cn(
            // kind === "audience" && "bg-emerald-800 bg-opacity-70",
            // "text-emerald-100",
            kind === "audience" && "text-[--emerald] dark:text-audienceScore-700 bg-black bg-opacity-20",
            kind === "user" && "bg-emerald-800 bg-opacity-90 text-white",
        )
    }
    // 90-100
    return cn(
        // kind === "audience" && "bg-indigo-600 bg-opacity-60 text-gray-100",
        // --indigo is theme-aware (indigo-700 light / indigo-300 dark), so dark is unchanged
        kind === "audience" && "text-[--indigo] bg-black bg-opacity-20",
        kind === "user" && "bg-indigo-600 bg-opacity-80 text-white",
    )
}
