import { Extension_Language, Extension_Type } from "@/api/generated/types"
import { useRunExtensionPlaygroundCode } from "@/api/hooks/extensions.hooks"
import { LuffyError } from "@/components/shared/luffy-error"
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@/components/shared/resizable"
import { AppLayoutStack } from "@/components/ui/app-layout"
import { Button } from "@/components/ui/button"
import { NumberInput } from "@/components/ui/number-input"
import { Select } from "@/components/ui/select"
import { TextInput } from "@/components/ui/text-input"
import { useDebounce } from "@/hooks/use-debounce"
import { javascript } from "@codemirror/lang-javascript"
import { StreamLanguage } from "@codemirror/language"
import { go } from "@codemirror/legacy-modes/mode/go"
import { vscodeDark } from "@uiw/codemirror-theme-vscode"
import CodeMirror from "@uiw/react-codemirror"
import { withImmer } from "jotai-immer"
import { useAtom } from "jotai/react"
import { atomWithStorage } from "jotai/utils"
import React from "react"

type Params = {
    animeTorrentProvider: {
        mediaId: number
        inputs: {
            search: {
                query: string
            },
            smartsearch: {
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
        }
    }
}

const DEFAULT_PARAMS: Params = {
    animeTorrentProvider: {
        mediaId: 0,
        inputs: {
            search: {
                query: "",
            },
            smartsearch: {
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
    },
}

const enum Functions {
    AnimeTorrentProviderSearch = "AnimeTorrentProvider.search",
    AnimeTorrentProviderSmartSearch = "AnimeTorrentProvider.smartSearch",
    AnimeTorrentProviderGetTorrentInfoHash = "AnimeTorrentProvider.getTorrentInfoHash",
    AnimeTorrentProviderGetTorrentMagnetLink = "AnimeTorrentProvider.getTorrentMagnetLink",
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
                query: inputs.animeTorrentProvider.inputs.search.query,
            }
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

                        <Select
                            value={type as string}
                            options={[
                                { value: "anime-torrent-provider", label: "Anime Torrent Provider" },
                                { value: "manga-provider", label: "Manga Provider" },
                                { value: "online-streaming-provider", label: "Online Streaming Provider" },
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
                            fieldClass="max-w-[250px]"
                        />

                        <Button intent="white" loading={isRunning} onClick={() => handleRunCode()}>
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
                        className="w-full border rounded-md !h-[65dvh] xl:!h-[75dvh] mt-8"
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
                                                extensions={[javascript({ typescript: language === "typescript" }), StreamLanguage.define(go)]}
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
                                            <div className="bg-gray-950 rounded-md border p-4">
                                    <pre className="text-sm max-h-[40rem]">
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
                                                    ]}
                                                    onValueChange={v => {
                                                        setSelectedFunction(v as Functions)
                                                    }}
                                                />

                                                <NumberInput
                                                    label="Media ID"
                                                    value={inputs.animeTorrentProvider.mediaId}
                                                    onValueChange={v => {
                                                        setInputs(d => {
                                                            d.animeTorrentProvider.mediaId = v
                                                            return
                                                        })
                                                    }}
                                                />

                                                {selectedFunction === Functions.AnimeTorrentProviderSearch && (
                                                    <>
                                                        <TextInput
                                                            label="Query"
                                                            type="text"
                                                            value={inputs.animeTorrentProvider.inputs.search.query}
                                                            onChange={e => {
                                                                setInputs(d => {
                                                                    d.animeTorrentProvider.inputs.search.query = e.target.value
                                                                    return
                                                                })
                                                            }}
                                                        />
                                                    </>
                                                )}
                                            </>
                                        )}

                                        <AppLayoutStack>
                                            <p className="font-semibold">Output</p>

                                            <div className="bg-gray-900 border rounded-md p-4">
                                        <pre className="text-sm text-white max-h-[40rem]">
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


