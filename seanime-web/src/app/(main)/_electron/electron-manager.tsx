"use client"

import React from "react"

type ElectronManagerProps = {
    children?: React.ReactNode
}

// This is only rendered on the Electron Desktop client
export function ElectronManager(props: ElectronManagerProps) {
    const {
        children,
        ...rest
    } = props

    // No-op

    return null
}
