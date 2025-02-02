import { AL_BaseAnime, AL_BaseManga } from "@/api/generated/types"
import { imageShimmer } from "@/components/shared/image-helpers"
import { CommandGroup, CommandItem, CommandShortcut } from "@/components/ui/command"
import Image from "next/image"
import { useSeaCommandContext } from "../sea-command"


export function CommandItemMedia({ media }: { media: AL_BaseAnime | AL_BaseManga }) {
    return (
        <div className="flex gap-3 items-center">
            <div className="size-12 flex-none rounded-[--radius-md] relative overflow-hidden">
                <Image
                    src={media.coverImage?.medium || ""}
                    alt="episode image"
                    fill
                    className="object-center object-cover"
                    placeholder={imageShimmer(700, 475)}
                />
            </div>
            <div className="flex gap-1 items-center w-full">
                <p className="w-full line-clamp-1">{media?.title?.userPreferred || ""}</p>
            </div>
        </div>
    )
}

export function CommandHelperText({ command, description, show }: { command: string, description: string, show: boolean }) {
    if (!show) return null
    return (
        <p className="py-1 px-6 text-center text-sm sm:px-14 tracking-widest text-[--muted]">
            <span className="text-[--foreground]">{command}</span> <span className="tracking-wide">{description}</span>
        </p>
    )
}

export function SeaCommandAutocompleteSuggestions({
    commands,
}: {
    commands: { command: string, description: string, show?: boolean }[]
}) {

    const { input, setInput, select, command: { isCommand, command, args }, scrollToTop } = useSeaCommandContext()

    if (input !== "/") return null

    return (
        <>

            <CommandGroup heading="Suggestions">
                {commands.filter(command => command.show === true).map(command => (
                    <CommandItem
                        key={command.command}
                        onSelect={() => {
                            setInput(`/${command.command}`)
                        }}
                    >
                        <span className="tracking-widest text-sm">/{command.command}</span>
                        <CommandShortcut className="text-[--muted]">{command.description}</CommandShortcut>
                    </CommandItem>
                ))}
            </CommandGroup>
        </>
    )
}
