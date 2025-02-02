import { Extension_Extension } from "@/api/generated/types"
import { useUpdateExtensionCode } from "@/api/hooks/extensions.hooks"
import { Button } from "@/components/ui/button"
import { Modal } from "@/components/ui/modal"
import { javascript } from "@codemirror/lang-javascript"
import { StreamLanguage } from "@codemirror/language"
import { go } from "@codemirror/legacy-modes/mode/go"
import { vscodeDark } from "@uiw/codemirror-theme-vscode"
import CodeMirror from "@uiw/react-codemirror"
import React from "react"


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
            onInteractOutside={e => e.preventDefault()}
            // size="xl"
            // contentClass="space-y-4"
        >
            <div>
                <p>
                    {extension.name}
                </p>
                <div className="text-sm text-[--muted]">
                    You can edit the code of the extension here.
                </div>
            </div>
            <div className="flex">
                <Button intent="white" loading={isPending} onClick={handleSave}>
                    Save
                </Button>
                <div className="flex flex-1"></div>
            </div>
            <ExtensionCodeEditor
                code={code}
                setCode={setCode}
                language={extension.language}
            />
        </Modal>
    )
}


function ExtensionCodeEditor({
    code,
    setCode,
    language,
}: { code: string, language: string, setCode: any }) {

    return (
        <div className="overflow-hidden rounded-[--radius-md]">
            <CodeMirror
                value={code}
                height="75vh"
                theme={vscodeDark}
                extensions={[javascript({ typescript: language === "typescript" }), StreamLanguage.define(go)]}
                onChange={setCode}
            />
        </div>
    )
}
