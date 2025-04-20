import { Extension_Language, Extension_Type } from "@/api/generated/types"
import { useRunExtensionPlaygroundCode } from "@/api/hooks/extensions.hooks"
import { LuffyError } from "@/components/shared/luffy-error"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/shared/resizable"
import { Alert } from "@/components/ui/alert"
import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { Textarea } from "@/components/ui/textarea"
import { useDebounce } from "@/hooks/use-debounce"
import { copyToClipboard } from "@/lib/helpers/browser"
import { autocompletion } from "@codemirror/autocomplete"
import { javascript } from "@codemirror/lang-javascript"
import { StreamLanguage } from "@codemirror/language"
import { go } from "@codemirror/legacy-modes/mode/go"
// import { vscodeKeymap } from "@replit/codemirror-vscode-keymap"
import { vscodeDark } from "@uiw/codemirror-theme-vscode"
import CodeMirror, { EditorView } from "@uiw/react-codemirror"
import { withImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import mousetrap from "mousetrap"
import React from "react"
import { BiCopy, BiTerminal } from "react-icons/bi"
import { toast } from "sonner"

type Params = {
    animeTorrentProvider: {
        mediaId: number
        search: {
            query: string
        },
        smartSearch: {
            query: string
            batch: boolean
            episodeNumber: number
            resolution: string
            bestReleases: boolean
        },
        getTorrentInfoHash: {
            torrent: string
        },
        getTorrentMagnetLink: {
            torrent: string
        },
    },
    mangaProvider: {
        mediaId: number
        findChapters: {
            id: string
        },
        findChapterPages: {
            id: string
        },
    },
    onlineStreamingProvider: {
        mediaId: number
        search: {
            dub: boolean
        }
        findEpisodes: {
            id: string
        },
        findEpisodeServer: {
            episode: string
            server: string
        },
    }
}

const DEFAULT_PARAMS: Params = {
    animeTorrentProvider: {
        mediaId: 0,
        search: {
            query: "",
        },
        smartSearch: {
            query: "",
            batch: false,
            episodeNumber: 0,
            resolution: "",
            bestReleases: false,
        },
        getTorrentInfoHash: {
            torrent: "",
        },
        getTorrentMagnetLink: {
            torrent: "",
        },
    },
    mangaProvider: {
        mediaId: 0,
        findChapters: {
            id: "",
        },
        findChapterPages: {
            id: "",
        },
    },
    onlineStreamingProvider: {
        mediaId: 0,
        search: {
            dub: false,
        },
        findEpisodes: {
            id: "",
        },
        findEpisodeServer: {
            episode: "",
            server: "",
        },
    },
}

const enum Functions {
    AnimeTorrentProviderSearch = "AnimeTorrentProvider.search",
    AnimeTorrentProviderSmartSearch = "AnimeTorrentProvider.smartSearch",
    AnimeTorrentProviderGetTorrentInfoHash = "AnimeTorrentProvider.getTorrentInfoHash",
    AnimeTorrentProviderGetTorrentMagnetLink = "AnimeTorrentProvider.getTorrentMagnetLink",
    AnimeTorrentProviderGetLatest = "AnimeTorrentProvider.getLatest",
    MangaProviderSearch = "MangaProvider.search",
    MangaProviderFindChapters = "MangaProvider.findChapters",
    MangaProviderFindChapterPages = "MangaProvider.findChapterPages",
    OnlinestreamSearch = "Onlinestream.search",
    OnlinestreamFindEpisodes = "Onlinestream.findEpisodes",
    OnlinestreamFindEpisodeServer = "Onlinestream.findEpisodeServer",
}

//----------------------------------------------------------------------------------------------------------------------------------------------------


type ExtensionPlaygroundProps = {
    language: Extension_Language
    onLanguageChange?: (lang: Extension_Language) => void
    type: Extension_Type
    onTypeChange?: (type: Extension_Type) => void
    code?: string
    onCodeChange?: (code: string) => void
}

const codeAtom = atomWithStorage<string>("sea-extension-playground-code", "", undefined, { getOnInit: true })
const paramsAtom = atomWithStorage<Params>("sea-extension-playground-params", DEFAULT_PARAMS, undefined, { getOnInit: true })

export function ExtensionPlayground(props: ExtensionPlaygroundProps) {

    const {
        language,
        onLanguageChange,
        type,
        onTypeChange,
        code: EXT_code,
        onCodeChange,
    } = props

    const { data: response, mutate: runCode, isPending: isRunning } = useRunExtensionPlaygroundCode()

    const [selectedFunction, setSelectedFunction] = React.useState(Functions.AnimeTorrentProviderSearch)
    const [inputs, setInputs] = useAtom(withImmer(paramsAtom))

    React.useLayoutEffect(() => {
        if (type === "anime-torrent-provider") {
            setSelectedFunction(Functions.AnimeTorrentProviderSearch)
        } else if (type === "manga-provider") {
            setSelectedFunction(Functions.MangaProviderSearch)
        } else if (type === "onlinestream-provider") {
            setSelectedFunction(Functions.OnlinestreamSearch)
        }
    }, [type])

    //
    // Code
    //

    const [code, setCode] = useAtom(codeAtom)
    const debouncedCode = useDebounce(code, 500)
    const codeRef = React.useRef("")

    React.useEffect(() => {
        if (!!EXT_code && EXT_code !== codeRef.current) {
            setCode(EXT_code)
        }
    }, [EXT_code])

    React.useEffect(() => {
        codeRef.current = code
        if (EXT_code !== code) {
            onCodeChange?.(code)
        }
    }, [debouncedCode])

    function handleRunCode() {

        let ret = {}
        let func = ""

        if (selectedFunction === Functions.AnimeTorrentProviderSearch) {
            func = "search"
            ret = {
                mediaId: inputs.animeTorrentProvider.mediaId,
                query: inputs.animeTorrentProvider.search.query,
            }
        } else if (selectedFunction === Functions.AnimeTorrentProviderSmartSearch) {
            func = "smartSearch"
            ret = {
                mediaId: inputs.animeTorrentProvider.mediaId,
                options: {
                    query: inputs.animeTorrentProvider.smartSearch.query,
                    episodeNumber: inputs.animeTorrentProvider.smartSearch.episodeNumber,
                    resolution: inputs.animeTorrentProvider.smartSearch.resolution,
                    batch: inputs.animeTorrentProvider.smartSearch.batch,
                    bestReleases: inputs.animeTorrentProvider.smartSearch.bestReleases,
                },
            }
        } else if (selectedFunction === Functions.AnimeTorrentProviderGetTorrentInfoHash) {
            func = "getTorrentInfoHash"
            ret = {
                mediaId: inputs.animeTorrentProvider.mediaId,
                torrent: inputs.animeTorrentProvider.getTorrentInfoHash.torrent,
            }
        } else if (selectedFunction === Functions.AnimeTorrentProviderGetTorrentMagnetLink) {
            func = "getTorrentMagnetLink"
            ret = {
                mediaId: inputs.animeTorrentProvider.mediaId,
                torrent: inputs.animeTorrentProvider.getTorrentMagnetLink.torrent,
            }
        } else if (selectedFunction === Functions.AnimeTorrentProviderGetLatest) {
            func = "getLatest"
            ret = {
                mediaId: inputs.animeTorrentProvider.mediaId,
            }
        } else if (selectedFunction === Functions.MangaProviderSearch) {
            func = "search"
            ret = {
                mediaId: inputs.mangaProvider.mediaId,
            }
        } else if (selectedFunction === Functions.MangaProviderFindChapters) {
            func = "findChapters"
            ret = {
                mediaId: inputs.mangaProvider.mediaId,
                id: inputs.mangaProvider.findChapters.id,
            }
        } else if (selectedFunction === Functions.MangaProviderFindChapterPages) {
            func = "findChapterPages"
            ret = {
                mediaId: inputs.mangaProvider.mediaId,
                id: inputs.mangaProvider.findChapterPages.id,
            }
        } else if (selectedFunction === Functions.OnlinestreamSearch) {
            func = "search"
            ret = {
                mediaId: inputs.onlineStreamingProvider.mediaId,
                dub: inputs.onlineStreamingProvider.search.dub,
            }
        } else if (selectedFunction === Functions.OnlinestreamFindEpisodes) {
            func = "findEpisodes"
            ret = {
                mediaId: inputs.onlineStreamingProvider.mediaId,
                id: inputs.onlineStreamingProvider.findEpisodes.id,
            }
        } else if (selectedFunction === Functions.OnlinestreamFindEpisodeServer) {
            func = "findEpisodeServer"
            ret = {
                mediaId: inputs.onlineStreamingProvider.mediaId,
                episode: inputs.onlineStreamingProvider.findEpisodeServer.episode,
                server: inputs.onlineStreamingProvider.findEpisodeServer.server,
            }
        } else {
            toast.error("Invalid function selected.")
            return
        }

        runCode({
            params: {
                type: type,
                language: language,
                code: code,
                function: func,
                inputs: ret,
            },
        })
    }

    React.useEffect(() => {
        mousetrap.bind(["cmd+s", "ctrl+s"], () => {
            handleRunCode()
        })

        return () => {
            mousetrap.unbind(["cmd+s", "ctrl+s"])
        }
    }, [])


    return (
        <>
            <div className="w-full">

                <div className="flex items-center w-full">
                    <div className="w-full flex items-center gap-4">
                        <h2 className="w-fit">Playground</h2>

                        <Select
                            value={type as string}
                            intent="filled"
                            options={[
                                { value: "anime-torrent-provider", label: "Anime Torrent Provider" },
                                { value: "manga-provider", label: "Manga Provider" },
                                { value: "onlinestream-provider", label: "Online Streaming Provider" },
                            ]}
                            onValueChange={v => {
                                onTypeChange?.(v as Extension_Type)
                            }}
                            disabled={!onTypeChange}
                            fieldClass="max-w-[250px]"
                        />

                        <Select
                            value={language as string}
                            options={[
                                { value: "typescript", label: "Typescript" },
                                { value: "javascript", label: "Javascript" },
                                // { value: "go", label: "Go" },
                            ]}
                            onValueChange={v => {
                                onLanguageChange?.(v as Extension_Language)
                            }}
                            disabled={!onLanguageChange}
                            fieldClass="max-w-[140px]"
                        />
                    </div>
                    <div className="flex items-center gap-2 lg:flex-none w-fit">

                        <Button intent="primary" loading={isRunning} onClick={() => handleRunCode()} leftIcon={<BiTerminal className="size-6" />}>
                            {isRunning ? "Running..." : "Run"}
                        </Button>

                    </div>
                </div>

                <div className="block lg:hidden">
                    <LuffyError title="Oops!">
                        Your screen size is too small.
                    </LuffyError>
                </div>

                <div className="hidden lg:block">
                    <ResizablePanelGroup
                        autoSaveId="sea-extension-playground-1"
                        direction="horizontal"
                        className="w-full border rounded-md !h-[calc(100vh-16rem)] xl:!h-[calc(100vh-14rem)] mt-8"
                    >
                        <ResizablePanel defaultSize={75}>
                            <ResizablePanelGroup direction="vertical" autoSaveId="sea-extension-playground-2">
                                <ResizablePanel defaultSize={75}>
                                    <div className="flex w-full h-full">
                                        <div className="overflow-y-auto rounded-tl-sm w-full">
                                            <CodeMirror
                                                value={code}
                                                height="100%"
                                                theme={vscodeDark}
                                                extensions={[
                                                    autocompletion({ defaultKeymap: false }),
                                                    // keymap.of(vscodeKeymap),
                                                    javascript({ typescript: language === "typescript" }),
                                                    StreamLanguage.define(go),
                                                    EditorView.theme({
                                                        "&": {
                                                            fontSize: "14px",
                                                            font: "'JetBrains Mono', monospace",
                                                        },
                                                    }),
                                                ]}
                                                onChange={setCode}
                                            />
                                        </div>
                                    </div>
                                </ResizablePanel>
                                <ResizableHandle />
                                <ResizablePanel defaultSize={25} className="!overflow-y-auto">
                                    <div className="flex w-full h-full p-2">
                                        <div className="w-full">
                                            <div className="bg-gray-950 rounded-md border max-w-full overflow-x-auto h-full">
                                                {/* <p className="font-semibold mb-2 p-2 border-b text-sm">Console</p> */}
                                                <pre className="h-full whitespace-pre-wrap break-all">
                                                    {response?.logs?.split("\n")?.filter(l => l.trim() !== "").map((l, i) => (
                                                        <p
                                                            key={i}
                                                            className={cn(
                                                                "w-full hover:bg-gray-800 hover:text-white text-sm py-1 px-2 tracking-wide leading-6",
                                                                i % 2 === 0 ? "bg-gray-950" : "bg-gray-900",
                                                                l.includes("|ERR|") && "text-white bg-red-800/10",
                                                                l.includes("|WRN|") && "text-orange-500",
                                                                l.includes("|INF|") && "text-blue-200",
                                                                l.includes("|TRC|") && "text-[--muted]",
                                                                l.includes("extension > (console.warn):") && "text-orange-200/80",
                                                            )}
                                                        >
                                                            {l.includes(" |") ? (
                                                                <>
                                                                    <span className="opacity-40 tracking-normal">{l.split(" |")?.[0]} </span>
                                                                    {l.includes("|DBG|") &&
                                                                        <span className="text-yellow-200/40 font-medium">|DBG|</span>}
                                                                    {l.includes("|ERR|") && <span className="text-red-400 font-medium">|ERR|</span>}
                                                                    {l.includes("|WRN|") &&
                                                                        <span className="text-orange-400 font-medium">|WRN|</span>}
                                                                    {l.includes("|INF|") && <span className="text-blue-400 font-medium">|INF|</span>}
                                                                    {l.includes("|TRC|") &&
                                                                        <span className="text-purple-400 font-medium">|TRC|</span>}
                                                                    <span>{l.split("|")?.[2]
                                                                        .replace("extension > (console.log):", "log >")
                                                                        .replace("extension > (console.error):", "error >")
                                                                        .replace("extension > (console.warn):", "warn >")
                                                                        .replace("extension > (console.info):", "info >")
                                                                        .replace("extension > (console.debug):", "debug >")
                                                                    }</span>
                                                                </>
                                                            ) : (
                                                                l
                                                            )}
                                                        </p>
                                                    ))}
                                                </pre>
                                            </div>
                                        </div>
                                    </div>
                                </ResizablePanel>
                            </ResizablePanelGroup>
                        </ResizablePanel>
                        <ResizableHandle />
                        <ResizablePanel defaultSize={25} className="!overflow-y-auto">
                            <div className="flex w-full h-full max-w-full overflow-y-auto">
                                <div className="w-full">
                                    <ResizablePanelGroup direction="vertical" autoSaveId="sea-extension-playground-3">

                                        <ResizablePanel defaultSize={30} className="!overflow-y-auto">
                                            {/* <div className="p-3 sticky z-[2] top-0 right-0 w-full border-b bg-[--background]">
                                             <Button intent="primary" size="sm" className="w-full" loading={isRunning} onClick={() => handleRunCode()} leftIcon={<BiTerminal className="size-6" />}>
                                             {isRunning ? "Running..." : "Run"}
                                             </Button>
                                             </div> */}

                                            <div className="space-y-4 p-3">
                                                {/*ANIME TORRENT PROVIDER*/}

                                                {type === "anime-torrent-provider" && (
                                                    <>
                                                        <Select
                                                            leftAddon="Method"
                                                            value={selectedFunction}
                                                            options={[
                                                                { value: Functions.AnimeTorrentProviderSearch, label: "search" },
                                                                { value: Functions.AnimeTorrentProviderSmartSearch, label: "smartSearch" },
                                                                {
                                                                    value: Functions.AnimeTorrentProviderGetTorrentInfoHash,
                                                                    label: "getTorrentInfoHash",
                                                                },
                                                                {
                                                                    value: Functions.AnimeTorrentProviderGetTorrentMagnetLink,
                                                                    label: "getTorrentMagnetLink",
                                                                },
                                                                { value: Functions.AnimeTorrentProviderGetLatest, label: "getLatest" },
                                                            ]}
                                                            onValueChange={v => {
                                                                setSelectedFunction(v as Functions)
                                                            }}
                                                            addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                        />

                                                        <NumberInput
                                                            leftAddon="Media ID"
                                                            min={0}
                                                            formatOptions={{ useGrouping: false }}
                                                            value={inputs.animeTorrentProvider.mediaId}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.animeTorrentProvider.mediaId = v
                                                                    return
                                                                })
                                                            }}
                                                            addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                        />

                                                        {selectedFunction === Functions.AnimeTorrentProviderSmartSearch && (
                                                            <>
                                                                <TextInput
                                                                    leftAddon="Query"
                                                                    type="text"
                                                                    value={inputs.animeTorrentProvider.smartSearch.query}
                                                                    onChange={e => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.smartSearch.query = e.target.value
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />

                                                                <NumberInput
                                                                    leftAddon="Episode Number"
                                                                    value={inputs.animeTorrentProvider.smartSearch.episodeNumber || 0}
                                                                    min={0}
                                                                    formatOptions={{ useGrouping: false }}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.smartSearch.episodeNumber = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />

                                                                <Select
                                                                    leftAddon="Resolution"
                                                                    options={[
                                                                        { value: "-", label: "Any" },
                                                                        { value: "1080p", label: "1080" },
                                                                        { value: "720p", label: "720" },
                                                                        { value: "540p", label: "540" },
                                                                        { value: "480p", label: "480" },
                                                                    ]}
                                                                    value={inputs.animeTorrentProvider.smartSearch.resolution || "-"}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.smartSearch.resolution = v === "-" ? "" : v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />

                                                                <Switch
                                                                    side="right"
                                                                    label="Batch"
                                                                    value={inputs.animeTorrentProvider.smartSearch.batch}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.smartSearch.batch = v
                                                                            return
                                                                        })
                                                                    }}
                                                                />

                                                                <Switch
                                                                    side="right"
                                                                    label="Best Releases"
                                                                    value={inputs.animeTorrentProvider.smartSearch.bestReleases}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.smartSearch.bestReleases = v
                                                                            return
                                                                        })
                                                                    }}
                                                                />
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.AnimeTorrentProviderSearch && (
                                                            <>
                                                                <TextInput
                                                                    leftAddon="Query"
                                                                    type="text"
                                                                    value={inputs.animeTorrentProvider.search.query}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.search.query = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.AnimeTorrentProviderGetTorrentInfoHash && (
                                                            <>
                                                                <Textarea
                                                                    leftAddon="Torrent JSON"
                                                                    value={inputs.animeTorrentProvider.getTorrentInfoHash.torrent}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.getTorrentInfoHash.torrent = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.AnimeTorrentProviderGetTorrentMagnetLink && (
                                                            <>
                                                                <Textarea
                                                                    label="Torrent JSON"
                                                                    value={inputs.animeTorrentProvider.getTorrentMagnetLink.torrent}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.animeTorrentProvider.getTorrentMagnetLink.torrent = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}
                                                    </>
                                                )}

                                                {/*MANGA PROVIDER*/}

                                                {type === "manga-provider" && (
                                                    <>
                                                        <Select
                                                            leftAddon="Method"
                                                            value={selectedFunction}
                                                            options={[
                                                                { value: Functions.MangaProviderSearch, label: "search" },
                                                                { value: Functions.MangaProviderFindChapters, label: "findChapters" },
                                                                { value: Functions.MangaProviderFindChapterPages, label: "findChapterPages" },
                                                            ]}
                                                            onValueChange={v => {
                                                                setSelectedFunction(v as Functions)
                                                            }}
                                                            addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                        />

                                                        <NumberInput
                                                            leftAddon="Media ID"
                                                            min={0}
                                                            formatOptions={{ useGrouping: false }}
                                                            value={inputs.mangaProvider.mediaId}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.mangaProvider.mediaId = v
                                                                    return
                                                                })
                                                            }}
                                                            addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                        />

                                                        {selectedFunction === Functions.MangaProviderSearch && (
                                                            <>
                                                                <Alert intent="info">
                                                                    Seanime will automatically select the best match based on the manga titles.
                                                                </Alert>
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.MangaProviderFindChapters && (
                                                            <>
                                                                <TextInput
                                                                    leftAddon="Manga ID"
                                                                    type="text"
                                                                    value={inputs.mangaProvider.findChapters.id}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.mangaProvider.findChapters.id = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.MangaProviderFindChapterPages && (
                                                            <>
                                                                <TextInput
                                                                    leftAddon="Chapter ID"
                                                                    type="text"
                                                                    value={inputs.mangaProvider.findChapterPages.id}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.mangaProvider.findChapterPages.id = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}
                                                    </>
                                                )}

                                                {/*ONLINE STREAMING PROVIDER*/}

                                                {type === "onlinestream-provider" && (
                                                    <>
                                                        <Select
                                                            leftAddon="Method"
                                                            value={selectedFunction}
                                                            options={[
                                                                { value: Functions.OnlinestreamSearch, label: "search" },
                                                                { value: Functions.OnlinestreamFindEpisodes, label: "findEpisodes" },
                                                                { value: Functions.OnlinestreamFindEpisodeServer, label: "findEpisodeServer" },
                                                            ]}
                                                            onValueChange={v => {
                                                                setSelectedFunction(v as Functions)
                                                            }}
                                                            addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                        />

                                                        <NumberInput
                                                            leftAddon="Media ID"
                                                            min={0}
                                                            formatOptions={{ useGrouping: false }}
                                                            value={inputs.onlineStreamingProvider.mediaId}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.onlineStreamingProvider.mediaId = v
                                                                    return
                                                                })
                                                            }}
                                                            addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                        />

                                                        {selectedFunction === Functions.OnlinestreamSearch && (
                                                            <>
                                                                <Alert intent="info" className="text-sm">
                                                                    Seanime will automatically select the best match based on the anime titles.
                                                                </Alert>

                                                                <Switch
                                                                    side="right"
                                                                    label="Dubbed"
                                                                    value={inputs.onlineStreamingProvider.search.dub}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.onlineStreamingProvider.search.dub = v
                                                                            return
                                                                        })
                                                                    }}
                                                                />
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.OnlinestreamFindEpisodes && (
                                                            <>
                                                                <TextInput
                                                                    leftAddon="Episode ID"
                                                                    type="text"
                                                                    value={inputs.onlineStreamingProvider.findEpisodes.id}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.onlineStreamingProvider.findEpisodes.id = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}

                                                        {selectedFunction === Functions.OnlinestreamFindEpisodeServer && (
                                                            <>
                                                                <Textarea
                                                                    leftAddon="Episode JSON"
                                                                    value={inputs.onlineStreamingProvider.findEpisodeServer.episode}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.onlineStreamingProvider.findEpisodeServer.episode = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                    className="text-sm"
                                                                />

                                                                <TextInput
                                                                    leftAddon="Server"
                                                                    type="text"
                                                                    value={inputs.onlineStreamingProvider.findEpisodeServer.server}
                                                                    onValueChange={v => {
                                                                        setInputs(d => {
                                                                            d.onlineStreamingProvider.findEpisodeServer.server = v
                                                                            return
                                                                        })
                                                                    }}
                                                                    addonClass="w-[100px] border-r font-semibold text-sm justify-center text-center"
                                                                />
                                                            </>
                                                        )}
                                                    </>
                                                )}
                                            </div>
                                        </ResizablePanel>


                                        <ResizableHandle />

                                        <ResizablePanel defaultSize={70}>
                                            <div className="h-full w-full p-2">
                                                <div className="flex items-center gap-2 justify-between mb-2">
                                                    <p className="font-semibold">Output</p>
                                                    <IconButton
                                                        intent="gray-subtle" size="sm" onClick={() => {
                                                        if (response?.value) {
                                                            copyToClipboard(response?.value || "")
                                                            toast.success("Copied to clipboard")
                                                        } else {
                                                            toast.warning("No output to copy")
                                                        }
                                                    }} icon={<BiCopy className="size-4" />}
                                                    />
                                                </div>

                                                <div className="bg-gray-950 border rounded-md max-w-full overflow-x-auto h-[calc(100%-2.5rem)]">
                                                    <pre className="text-sm text-white h-full break-all max-w-full">
                                                        {response?.value?.split("\n").map((l, i) => (
                                                            <p
                                                                key={i}
                                                                className={cn(
                                                                    "w-full px-2 py-[.15rem] text-[.8rem] tracking-wider break-all",
                                                                    i % 2 === 0 ? "bg-gray-950" : "bg-gray-900",
                                                                    "hover:bg-gray-800 hover:text-white",
                                                                )}
                                                            >{l}</p>
                                                        ))}
                                                    </pre>
                                                </div>
                                            </div>
                                        </ResizablePanel>

                                    </ResizablePanelGroup>
                                </div>
                            </div>
                        </ResizablePanel>
                    </ResizablePanelGroup>
                </div>

            </div>
        </>
    )
}


