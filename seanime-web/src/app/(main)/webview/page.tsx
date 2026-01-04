"use client"
import { PluginWebviewSlot } from "@/app/(main)/_features/plugin/webview/plugin-webviews"

export default function Page() {
    return <>
        <PluginWebviewSlot slot="screen" />
    </>
}
