import { cn } from "@/components/ui/core/styling"

export function getScoreColor(score: number, kind: "audience" | "user"): string {
    if (score < 30) { // 0-29
        return cn(
            kind === "audience" && "bg-red-500 bg-opacity-30",
            kind === "user" && "bg-red-800 bg-opacity-90",
            "text-red-200",
        )
    }
    if (score < 50) { // 30-49
        return cn(
            kind === "audience" && "bg-orange-500 bg-opacity-20",
            kind === "user" && "bg-orange-800 bg-opacity-90",
            "text-orange-200",
        )
    }
    if (score < 60) { // 50-59
        return cn(
            kind === "audience" && "bg-amber-500 bg-opacity-20",
            kind === "user" && "bg-amber-800 bg-opacity-90",
            "text-amber-200",
        )
    }
    if (score < 70) { // 60-69
        return cn(
            kind === "audience" && "bg-lime-500 bg-opacity-20",
            kind === "user" && "bg-lime-800 bg-opacity-90",
            "text-lime-200",
        )
    }
    if (score < 82) { // 70-81
        return cn(
            kind === "audience" && "bg-emerald-400 bg-opacity-20",
            kind === "user" && "bg-emerald-800 bg-opacity-90",
            "text-emerald-100",
        )
    }
    // 82-100
    return cn(
        kind === "audience" && "bg-indigo-500 bg-opacity-30",
        kind === "user" && "bg-indigo-600 bg-opacity-80 text-white",
    )
}
