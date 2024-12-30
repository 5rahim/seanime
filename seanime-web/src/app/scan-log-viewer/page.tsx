"use client"

import { ScanLogViewer } from "@/app/scan-log-viewer/scan-log-viewer"
import { TextInput } from "@/components/ui/text-input"
import React, { useRef, useState } from "react"

export default function Page() {
    const [content, setContent] = useState<string>("")
    const fileInputRef = useRef<HTMLInputElement>(null)

    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0]
        if (file) {
            const reader = new FileReader()
            reader.onload = (e) => {
                const content = e.target?.result as string
                setContent(content)
            }
            reader.readAsText(file)
        }
    }

    return (
        <div className="container mx-auto bg-gray-900 p-4 min-h-screen relative">
            {/*<h1 className="text-3xl font-bold mb-6 text-brand-300 text-center">Scan Log Viewer</h1>*/}
            <div className="container max-w-2xl">
                <TextInput
                    type="file"
                    ref={fileInputRef}
                    onChange={handleFileChange}
                    className="mb-6 p-1"
                />
            </div>
            <ScanLogViewer content={content} />
        </div>
    )
}
