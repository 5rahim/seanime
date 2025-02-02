import { AL_BaseAnime_Title, AL_BaseManga_Title } from "@/api/generated/types"
import { Nullish } from "@/types/common"

export function useSeaCommand_ParseCommand(input: string) {
    const [command, ...args] = input.split(/\s+/)

    const ret = {
        isCommand: input.startsWith("/"),
        command: command.slice(1),
        args: args,
    }

    return ret
}

export type SeaCommand_ParsedCommandProps = ReturnType<typeof useSeaCommand_ParseCommand>

export function seaCommand_compareMediaTitles(titles: Nullish<AL_BaseAnime_Title | AL_BaseManga_Title>, query: string) {
    if (!titles) return false
    return (!!titles.english && cleanMediaTitle(titles.english).includes(cleanMediaTitle(query)))
        || (!!titles.romaji && cleanMediaTitle(titles.romaji).includes(cleanMediaTitle(query)))
}

function cleanMediaTitle(str: string) {
    // remove all non-alphanumeric characters
    return str.replace(/[^a-zA-Z0-9 ]/g, "").toLowerCase()
}
