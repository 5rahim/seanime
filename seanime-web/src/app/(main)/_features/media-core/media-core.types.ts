import React from "react"

export interface MediaCoreChapter {
    startTime: number
    endTime: number
    text: string
}

export interface MediaCoreSelectOption {
    label: string
    value: any
    moreInfo?: string
    description?: string
}
