import { Extension_Extension } from "@/api/generated/types"
import { useUpdateExtensionCode } from "@/api/hooks/extensions.hooks"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { ScrollArea } from "@/components/ui/scroll-area"
import dynamic from "next/dynamic"
import React from "react"
import "@uiw/react-textarea-code-editor/dist.css"

const CodeEditor = dynamic(() => import("@uiw/react-textarea-code-editor").then((mod) => mod.default), { ssr: false })

type ExtensionCodeModalProps = {
    children?: React.ReactElement
    extension: Extension_Extension
}

export function ExtensionCodeModal(props: ExtensionCodeModalProps) {

    const {
        children,
        extension,
        ...rest
    } = props

    const [code, setCode] = React.useState(extension.payload)

    const { mutate: updateCode, isPending } = useUpdateExtensionCode()

    React.useLayoutEffect(() => {
        setCode(extension.payload)
    }, [extension.payload])

    function handleSave() {
        if (isPending) {
            return
        }
        if (code === extension.payload) {
            return
        }
        if (code.length === 0) {
            return
        }
        updateCode({
            id: extension.id,
            payload: code,
        })
    }

    return (
        <Modal
            contentClass="max-w-5xl"
            trigger={children}
            title="Code"
        >
            <div className="flex">
                <div className="flex flex-1"></div>
                <Button loading={isPending} onClick={handleSave}>
                    Save
                </Button>
            </div>
            <ExtensionCodeEditor code={code} setCode={setCode} />
        </Modal>
    )
}


function ExtensionCodeEditor({
    code,
    setCode,
}: { code: string, setCode: any }) {

    return (
        <ScrollArea className="max-h-[70vh]">
            <CodeEditor
                data-color-mode="dark"
                value={code}
                language="ts"
                placeholder="Please enter the code."
                onChange={(evn) => setCode(evn.target.value)}
                style={{
                    fontSize: 15,
                    backgroundColor: "#0e0e0e",
                    borderRadius: "1rem",
                    fontFamily: "ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace",
                }}
            />
        </ScrollArea>
    )
}
