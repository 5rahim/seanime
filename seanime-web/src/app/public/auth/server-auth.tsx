import { serverAuthTokenAtom } from "@/app/(main)/_atoms/server-status.atoms"
import { defineSchema, Field, Form } from "@/components/ui/form"
import { Modal } from "@/components/ui/modal"
import { useAtom } from "jotai"
import { sha256 } from "js-sha256"
import React, { useState } from "react"

// async function hashSHA256Hex(str: string): Promise<string> {
//     const encoder = new TextEncoder()
//     const data = encoder.encode(str)
//     const hashBuffer = await window.crypto.subtle.digest("SHA-256", data)
//     return Array.from(new Uint8Array(hashBuffer)).map(b => b.toString(16).padStart(2, "0")).join("")
// }

export function ServerAuth() {

    const [, setAuthToken] = useAtom(serverAuthTokenAtom)
    const [loading, setLoading] = useState(false)

    return (<>
        <Modal
            title="Password required"
            description="This Seanime server requires authentication."
            open={true}
            onOpenChange={(v) => {}}
            overlayClass="bg-opacity-100 bg-gray-900"
            contentClass="border focus:outline-none focus-visible:outline-none outline-none"
            hideCloseButton
        >
            <Form
                schema={defineSchema(({ z }) => z.object({
                    password: z.string().min(1, "Password is required"),
                }))}
                onSubmit={async data => {
                    setLoading(true)
                    // const hash = await hashSHA256Hex(data.password)
                    const hash = sha256(data.password)
                    setAuthToken(hash)
                    React.startTransition(() => {
                        window.location.href = "/"
                        setLoading(false)
                    })
                }}
            >
                <Field.Text
                    type="password"
                    name="password"
                    label="Enter the password"
                    fieldClass=""
                />
                <Field.Submit showLoadingOverlayOnSuccess loading={loading}>Continue</Field.Submit>
            </Form>
        </Modal>
    </>)
}
