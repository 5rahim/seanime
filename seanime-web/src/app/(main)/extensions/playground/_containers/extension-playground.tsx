import { Extension_Language, Extension_Type } from "@/api/generated/types"
import { useRunExtensionPlaygroundCode } from "@/api/hooks/extensions.hooks"
import { LuffyError } from "@/components/shared/luffy-error"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/shared/resizable"
import { Alert } from "@/components/ui/alert"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { Separator } from "@/components/ui/separator"
import { Switch } from "@/components/ui/switch"
import { TextInput } from "@/components/ui/text-input"
import { Textarea } from "@/components/ui/textarea"
import { useDebounce } from "@/hooks/use-debounce"
import { autocompletion } from "@codemirror/autocomplete"
import { javascript } from "@codemirror/lang-javascript"
import { StreamLanguage } from "@codemirror/language"
import { go } from "@codemirror/legacy-modes/mode/go"
import { vscodeKeymap } from "@replit/codemirror-vscode-keymap"
import { vscodeDark } from "@uiw/codemirror-theme-vscode"
import CodeMirror, { keymap } from "@uiw/react-codemirror"
import { withImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"
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
        findEpisodeServers: {
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
        findEpisodeServers: {
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
    OnlinestreamFindEpisodeServers = "Onlinestream.findEpisodeServers",
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
        } else if (selectedFunction === Functions.OnlinestreamFindEpisodeServers) {
            func = "findEpisodeServers"
            ret = {
                mediaId: inputs.onlineStreamingProvider.mediaId,
                episode: inputs.onlineStreamingProvider.findEpisodeServers.episode,
                server: inputs.onlineStreamingProvider.findEpisodeServers.server,
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


    return (
        <>
            <div className="">

                <div className="grid grid-cols-1 gap-4 xl:grid-cols-[1fr,minmax(0,1fr)]">
                    <h2 className="w-fit">Playground</h2>
                    <div className="hidden lg:flex flex-wrap gap-2 lg:justify-end">

                        <Button intent="white" loading={isRunning} onClick={() => handleRunCode()}>
                            {isRunning ? "Running..." : "Run"}
                        </Button>

                        <Select
                            value={type as string}
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
                                { value: "go", label: "Go" },
                            ]}
                            onValueChange={v => {
                                onLanguageChange?.(v as Extension_Language)
                            }}
                            disabled={!onLanguageChange}
                            fieldClass="max-w-[140px]"
                        />

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
                                        <div className="overflow-y-auto rounded-tl-md w-full">
                                            <CodeMirror
                                                value={code}
                                                height="100%"
                                                theme={vscodeDark}
                                                extensions={[
                                                    autocompletion({ defaultKeymap: false }),
                                                    keymap.of(vscodeKeymap),
                                                    javascript({ typescript: language === "typescript" }),
                                                    StreamLanguage.define(go),
                                                ]}
                                                onChange={setCode}
                                            />
                                        </div>
                                    </div>
                                </ResizablePanel>
                                <ResizableHandle />
                                <ResizablePanel defaultSize={25} className="!overflow-y-auto">
                                    <div className="flex w-full h-full p-6">
                                        <AppLayoutStack className="w-full">
                                            <p className="font-semibold">Console</p>
                                            <div className="bg-gray-900 rounded-md border max-w-full overflow-x-auto">
                                                <pre className="text-sm max-h-[40rem] p-2 min-h-12 text-white">
                                                    {response?.logs}
                                                </pre>
                                            </div>
                                        </AppLayoutStack>
                                    </div>
                                </ResizablePanel>
                            </ResizablePanelGroup>
                        </ResizablePanel>
                        <ResizableHandle />
                        <ResizablePanel defaultSize={25} className="!overflow-y-auto">
                            <div className="flex w-full h-full max-w-full overflow-y-auto p-6">
                                <div className="w-full">
                                    <div className="space-y-4">

                                        {/*ANIME TORRENT PROVIDER*/}

                                        {type === "anime-torrent-provider" && (
                                            <>
                                                <Select
                                                    label="Method"
                                                    value={selectedFunction}
                                                    options={[
                                                        { value: Functions.AnimeTorrentProviderSearch, label: "search" },
                                                        { value: Functions.AnimeTorrentProviderSmartSearch, label: "smartSearch" },
                                                        { value: Functions.AnimeTorrentProviderGetTorrentInfoHash, label: "getTorrentInfoHash" },
                                                        { value: Functions.AnimeTorrentProviderGetTorrentMagnetLink, label: "getTorrentMagnetLink" },
                                                        { value: Functions.AnimeTorrentProviderGetLatest, label: "getLatest" },
                                                    ]}
                                                    onValueChange={v => {
                                                        setSelectedFunction(v as Functions)
                                                    }}
                                                />

                                                <NumberInput
                                                    label="Media ID"
                                                    min={0}
                                                    formatOptions={{ useGrouping: false }}
                                                    value={inputs.animeTorrentProvider.mediaId}
                                                    onValueChange={v => {
                                                        setInputs(d => {
                                                            d.animeTorrentProvider.mediaId = v
                                                            return
                                                        })
                                                    }}
                                                />

                                                {selectedFunction === Functions.AnimeTorrentProviderSmartSearch && (
                                                    <>
                                                        <TextInput
                                                            label="Query"
                                                            type="text"
                                                            value={inputs.animeTorrentProvider.smartSearch.query}
                                                            onChange={e => {
                                                                setInputs(d => {
                                                                    d.animeTorrentProvider.smartSearch.query = e.target.value
                                                                    return
                                                                })
                                                            }}
                                                        />

                                                        <NumberInput
                                                            label="Episode Number"
                                                            value={inputs.animeTorrentProvider.smartSearch.episodeNumber || 0}
                                                            min={0}
                                                            formatOptions={{ useGrouping: false }}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.animeTorrentProvider.smartSearch.episodeNumber = v
                                                                    return
                                                                })
                                                            }}
                                                        />

                                                        <Select
                                                            label="Resolution"
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
                                                        />

                                                        <Switch
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
                                                            label="Query"
                                                            type="text"
                                                            value={inputs.animeTorrentProvider.smartSearch.query}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.animeTorrentProvider.smartSearch.query = v
                                                                    return
                                                                })
                                                            }}
                                                        />
                                                    </>
                                                )}

                                                {selectedFunction === Functions.AnimeTorrentProviderGetTorrentInfoHash && (
                                                    <>
                                                        <Textarea
                                                            label="Torrent JSON"
                                                            value={inputs.animeTorrentProvider.getTorrentInfoHash.torrent}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.animeTorrentProvider.getTorrentInfoHash.torrent = v
                                                                    return
                                                                })
                                                            }}
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
                                                        />
                                                    </>
                                                )}
                                            </>
                                        )}

                                        {/*MANGA PROVIDER*/}

                                        {type === "manga-provider" && (
                                            <>
                                                <Select
                                                    label="Method"
                                                    value={selectedFunction}
                                                    options={[
                                                        { value: Functions.MangaProviderSearch, label: "search" },
                                                        { value: Functions.MangaProviderFindChapters, label: "findChapters" },
                                                        { value: Functions.MangaProviderFindChapterPages, label: "findChapterPages" },
                                                    ]}
                                                    onValueChange={v => {
                                                        setSelectedFunction(v as Functions)
                                                    }}
                                                />

                                                <NumberInput
                                                    label="Media ID"
                                                    min={0}
                                                    formatOptions={{ useGrouping: false }}
                                                    value={inputs.mangaProvider.mediaId}
                                                    onValueChange={v => {
                                                        setInputs(d => {
                                                            d.mangaProvider.mediaId = v
                                                            return
                                                        })
                                                    }}
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
                                                            label="Manga ID"
                                                            type="text"
                                                            value={inputs.mangaProvider.findChapters.id}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.mangaProvider.findChapters.id = v
                                                                    return
                                                                })
                                                            }}
                                                        />
                                                    </>
                                                )}

                                                {selectedFunction === Functions.MangaProviderFindChapterPages && (
                                                    <>
                                                        <TextInput
                                                            label="Chapter ID"
                                                            type="text"
                                                            value={inputs.mangaProvider.findChapterPages.id}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.mangaProvider.findChapterPages.id = v
                                                                    return
                                                                })
                                                            }}
                                                        />
                                                    </>
                                                )}
                                            </>
                                        )}

                                        {/*ONLINE STREAMING PROVIDER*/}

                                        {type === "onlinestream-provider" && (
                                            <>
                                                <Select
                                                    label="Method"
                                                    value={selectedFunction}
                                                    options={[
                                                        { value: Functions.OnlinestreamSearch, label: "search" },
                                                        { value: Functions.OnlinestreamFindEpisodes, label: "findEpisode" },
                                                        { value: Functions.OnlinestreamFindEpisodeServers, label: "findEpisodeServers" },
                                                    ]}
                                                    onValueChange={v => {
                                                        setSelectedFunction(v as Functions)
                                                    }}
                                                />

                                                <NumberInput
                                                    label="Media ID"
                                                    min={0}
                                                    formatOptions={{ useGrouping: false }}
                                                    value={inputs.onlineStreamingProvider.mediaId}
                                                    onValueChange={v => {
                                                        setInputs(d => {
                                                            d.onlineStreamingProvider.mediaId = v
                                                            return
                                                        })
                                                    }}
                                                />

                                                {selectedFunction === Functions.OnlinestreamSearch && (
                                                    <>
                                                        <Alert intent="info">
                                                            Seanime will automatically select the best match based on the anime titles.
                                                        </Alert>

                                                        <Switch
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
                                                            label="Episode ID"
                                                            type="text"
                                                            value={inputs.onlineStreamingProvider.findEpisodes.id}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.onlineStreamingProvider.findEpisodes.id = v
                                                                    return
                                                                })
                                                            }}
                                                        />
                                                    </>
                                                )}

                                                {selectedFunction === Functions.OnlinestreamFindEpisodeServers && (
                                                    <>
                                                        <Textarea
                                                            label="Episode JSON"
                                                            value={inputs.onlineStreamingProvider.findEpisodeServers.episode}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.onlineStreamingProvider.findEpisodeServers.episode = v
                                                                    return
                                                                })
                                                            }}
                                                        />

                                                        <TextInput
                                                            label="Server"
                                                            type="text"
                                                            value={inputs.onlineStreamingProvider.findEpisodeServers.server}
                                                            onValueChange={v => {
                                                                setInputs(d => {
                                                                    d.onlineStreamingProvider.findEpisodeServers.server = v
                                                                    return
                                                                })
                                                            }}
                                                        />
                                                    </>
                                                )}
                                            </>
                                        )}

                                        <Separator />

                                        <AppLayoutStack>
                                            <p className="font-semibold">Output</p>

                                            <div className="bg-gray-900 border rounded-md max-w-full overflow-x-auto">
                                                <pre className="text-sm text-white min-h-12 max-h-[40rem] p-2">
                                                    {response?.value}
                                                </pre>
                                            </div>
                                        </AppLayoutStack>

                                    </div>
                                </div>
                            </div>
                        </ResizablePanel>
                    </ResizablePanelGroup>
                </div>

            </div>
        </>
    )
}


