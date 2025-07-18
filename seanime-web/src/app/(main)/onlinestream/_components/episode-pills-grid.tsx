import { cn } from "@/components/ui/core/styling"
import { motion } from "motion/react"
import React from "react"

type Episode = {
    number: number
    title?: string | null
    isFiller?: boolean
}

type EpisodePillsGridProps = {
    episodes: Episode[]
    currentEpisodeNumber: number
    onEpisodeSelect: (episodeNumber: number) => void
    progress?: number
    disabled?: boolean
    className?: string
    getEpisodeId: (episode: Episode) => string
}

export function EpisodePillsGrid({
    episodes,
    currentEpisodeNumber,
    onEpisodeSelect,
    progress = 0,
    disabled = false,
    className,
    getEpisodeId,
}: EpisodePillsGridProps) {
    return (
        <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
            className={cn(
                "grid grid-cols-6 sm:grid-cols-8 md:grid-cols-10 lg:grid-cols-10 xl:grid-cols-10 2xl:grid-cols-6 gap-2 pb-8",
                className,
            )}
        >
            {episodes
                ?.filter(Boolean)
                ?.sort((a, b) => a!.number - b!.number)
                ?.map((episode) => {
                    const isSelected = episode.number === currentEpisodeNumber
                    const isWatched = progress > 0 && episode.number <= progress
                    const isFiller = episode.isFiller

                    return (
                        <motion.button
                            key={episode.number}
                            // initial={{ scale: 0.95, opacity: 0 }}
                            // animate={{ scale: 1, opacity: 1 }}
                            // whileHover={{ scale: disabled ? 1 : 1.02 }}
                            // whileTap={{ scale: disabled ? 1 : 0.98 }}
                            // transition={{
                            //     duration: 0.15,
                            //     // delay: episode.number * 0.005
                            // }}
                            onClick={() => !disabled && onEpisodeSelect(episode.number)}
                            disabled={disabled}
                            title={episode.title || `Episode ${episode.number}`}
                            id={getEpisodeId(episode)}
                            className={cn(
                                "relative flex items-center justify-center",
                                "w-full h-10 rounded-md font-medium text-sm",
                                "transition-all duration-150 ease-out",
                                "focus:outline-none",
                                !isSelected && [
                                    "bg-[--subtle]",
                                    "hover:bg-transparent",
                                ],
                                isSelected && [
                                    "bg-brand-500 text-white",
                                ],
                                isFiller && !isSelected && [
                                    "text-orange-300",
                                ],
                                isWatched && !isSelected && [
                                    "text-[--muted]",
                                ],
                                disabled && [
                                    "opacity-50 cursor-not-allowed",
                                    "hover:bg-inherit hover:text-inherit hover:scale-100",
                                ],
                            )}
                        >
                            <span className="relative z-10">{episode.number}</span>

                            {isFiller && (
                                <div
                                    className={cn(
                                        "absolute top-1 right-1 w-1.5 h-1.5 rounded-full",
                                        "bg-orange-400",
                                        isSelected && "bg-orange-200",
                                    )}
                                    title="Filler episode"
                                />
                            )}

                            {/* {isWatched && !isSelected && (
                             <div
                             className="absolute bottom-1 left-1/2 transform -translate-x-1/2 w-1 h-1 rounded-full bg-[--brand]"
                             />
                             )} */}
                        </motion.button>
                    )
                })}
        </motion.div>
    )
}
