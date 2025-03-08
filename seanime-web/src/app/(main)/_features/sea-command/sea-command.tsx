import { SeaCommandActions } from "@/app/(main)/_features/sea-command/sea-command-actions"
import { SeaCommandSearch } from "@/app/(main)/_features/sea-command/sea-command-search"
import { SeaCommand_ParsedCommandProps, useSeaCommand_ParseCommand } from "@/app/(main)/_features/sea-command/utils"
import { CommandDialog, CommandInput, CommandList } from "@/components/ui/command"
import { atom } from "jotai"
import { useAtom, useSetAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import mousetrap from "mousetrap"
import { usePathname, useRouter } from "next/navigation"
import React from "react"
import { SeaCommandHandler } from "./config"
import { SeaCommandInjectables } from "./sea-command-injectables"
import { SeaCommandList } from "./sea-command-list"
import { SeaCommandNavigation, SeaCommandUserMediaNavigation } from "./sea-command-navigation"

export const __seaCommand_shortcuts = atomWithStorage<string[]>("sea-command-shortcuts", ["meta+j", "q"], undefined, { getOnInit: true })

export type SeaCommandContextProps = {
    input: string
    setInput: (input: string) => void
    resetInput: () => void
    close: () => void
    select: (func?: () => void) => void
    command: SeaCommand_ParsedCommandProps
    scrollToTop: () => () => void
    commandListRef: React.RefObject<HTMLDivElement>
    router: {
        pathname: string
    }
}

export const SeaCommandContext = React.createContext<SeaCommandContextProps>({
    input: "",
    setInput: () => { },
    resetInput: () => { },
    close: () => { },
    select: () => { },
    command: { command: "", isCommand: false, args: [] },
    scrollToTop: () => () => { },
    commandListRef: React.createRef<HTMLDivElement>(),
    router: {
        pathname: "",
    },
})

export function useSeaCommandContext() {
    return React.useContext(SeaCommandContext) as SeaCommandContextProps
}

const __seaCommand_isOpen = atom(false)

export function useOpenSeaCommand() {
    const setOpen = useSetAtom(__seaCommand_isOpen)
    return {
        setSeaCommandOpen: setOpen,
    }
}

export function SeaCommand() {

    const router = useRouter()
    const pathname = usePathname()

    const [open, setOpen] = useAtom(__seaCommand_isOpen)
    const [input, setInput] = React.useState("")
    const [activeItemId, setActiveItemId] = React.useState("")

    const [shortcuts, setShortcuts] = useAtom(__seaCommand_shortcuts)

    const parsedCommandProps = useSeaCommand_ParseCommand(input)

    React.useEffect(() => {
        mousetrap.bind(shortcuts, () => {
            setInput("")
            React.startTransition(() => {
                setOpen(true)
            })
        })

        return () => {
            mousetrap.unbind(shortcuts)
        }
    }, [shortcuts])

    React.useEffect(() => {
        if (!open) setInput("")
    }, [open])

    React.useEffect(() => {
        mousetrap.bind(["s"], () => {
            setOpen(true)
            React.startTransition(() => {
                setTimeout(() => {
                    setInput("/search ")
                }, 100)
            })
        })

        return () => {
            mousetrap.unbind(["s"])
        }
    }, [])

    const commandListRef = React.useRef<HTMLDivElement>(null)

    function scrollToTop() {
        const list = commandListRef.current
        if (!list) return () => {}

        const t = setTimeout(() => {
            list.scrollTop = 0
            // Find and focus the first command item
            const firstItem = list.querySelector("[cmdk-item]") as HTMLElement
            if (firstItem) {
                const value = firstItem.getAttribute("data-value")
                if (value) {
                    setActiveItemId(value)
                }
            }
        }, 100)

        return () => clearTimeout(t)
    }

    React.useEffect(() => {
        const cl = scrollToTop()
        return () => cl()
    }, [input, pathname])

    return (
        <SeaCommandContext.Provider
            value={{
                input: input,
                setInput: setInput,
                resetInput: () => setInput(""),
                close: () => {
                    React.startTransition(() => {
                        setOpen(false)
                    })
                },
                select: (func?: () => void) => {
                    func?.()
                    setInput("")
                },
                scrollToTop,
                command: parsedCommandProps,
                commandListRef: commandListRef,
                router: {
                    pathname: pathname,
                },
            }}
        >
            <CommandDialog
                open={open}
                onOpenChange={setOpen}
                commandProps={{
                    value: activeItemId,
                    onValueChange: setActiveItemId,
                }}
                overlayClass="bg-black/30"
                contentClass="max-w-2xl"
                commandClass="h-[300px]"
            >

                <CommandInput
                    placeholder="Type a command or input..."
                    value={input}
                    onValueChange={setInput}
                />
                <CommandList className="mb-2" ref={commandListRef}>

                    {/*Active commands*/}
                    <SeaCommandHandler
                        shouldShow={ctx => ctx.command.command === "search"}
                        render={() => <SeaCommandSearch />}
                    />
                    <SeaCommandHandler
                        shouldShow={ctx => (
                            ctx.command.command === "anime"
                            || ctx.command.command === "manga"
                            // || ctx.command.command === "library"
                        )}
                        render={() => <SeaCommandUserMediaNavigation />}
                    />
                    <SeaCommandHandler
                        shouldShow={ctx => (
                            ctx.command.command === "logs"
                            || ctx.command.command === "issue"
                        )}
                        render={() => <SeaCommandActions />}
                    />

                    {/*Injected items*/}
                    <SeaCommandHandler
                        shouldShow={() => true}
                        render={() => <SeaCommandInjectables />}
                    />

                    {/*Page items*/}
                    {/* <SeaCommandHandler
                     type="anime-entry"
                     shouldShow={ctx => !ctx.command.isCommand && ctx.params.page === "anime-entry"}
                     render={() => <SeaCommandAnimeEntry />}
                     /> */}
                    {/* <SeaCommandHandler
                     type="anime-library"
                     shouldShow={ctx => !ctx.command.isCommand && ctx.params.page === "anime-library"}
                     render={() => <SeaCommandAnimeLibrary />}
                     /> */}

                    {/*Suggestions*/}
                    <SeaCommandHandler
                        shouldShow={ctx => ctx.input.startsWith("/")}
                        render={() => <SeaCommandList />}
                    />

                    {/*Navigation*/}
                    <SeaCommandHandler
                        shouldShow={() => true}
                        render={() => <SeaCommandNavigation />}
                    />

                </CommandList>
            </CommandDialog>
        </SeaCommandContext.Provider>
    )
}
