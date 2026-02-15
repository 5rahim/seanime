import { ScanLogViewer } from "@/app/scan-log-viewer/scan-log-viewer"
import React, { useCallback, useEffect, useRef, useState } from "react"
import { BiTrash, BiUpload } from "react-icons/bi"
import { toast } from "sonner"

const DB_NAME = "seanime-scan-logs-db"
const STORE_NAME = "logs"
const KEY = "latest_log"

const initDB = (): Promise<IDBDatabase> => {
    return new Promise((resolve, reject) => {
        const request = indexedDB.open(DB_NAME, 1)
        request.onerror = () => reject(request.error)
        request.onsuccess = () => resolve(request.result)
        request.onupgradeneeded = (event) => {
            const db = (event.target as IDBOpenDBRequest).result
            if (!db.objectStoreNames.contains(STORE_NAME)) {
                db.createObjectStore(STORE_NAME)
            }
        }
    })
}

const saveLogToDB = async (content: string) => {
    try {
        const db = await initDB()
        return new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, "readwrite")
            const store = tx.objectStore(STORE_NAME)
            const request = store.put(content, KEY)
            request.onsuccess = () => resolve()
            request.onerror = () => reject(request.error)
        })
    }
    catch (error) {
        console.error("Failed to save log to DB:", error)
        toast.error("Failed to save log to browser storage")
    }
}

const getLogFromDB = async (): Promise<string | null> => {
    try {
        const db = await initDB()
        return new Promise((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, "readonly")
            const store = tx.objectStore(STORE_NAME)
            const request = store.get(KEY)
            request.onsuccess = () => resolve(request.result || null)
            request.onerror = () => reject(request.error)
        })
    }
    catch (error) {
        console.error("Failed to get log from DB:", error)
        return null
    }
}

const clearLogFromDB = async () => {
    try {
        const db = await initDB()
        return new Promise<void>((resolve, reject) => {
            const tx = db.transaction(STORE_NAME, "readwrite")
            const store = tx.objectStore(STORE_NAME)
            const request = store.delete(KEY)
            request.onsuccess = () => resolve()
            request.onerror = () => reject(request.error)
        })
    }
    catch (error) {
        console.error("Failed to clear log from DB:", error)
    }
}

export default function Page() {
    const [content, setContent] = useState<string>("")
    const [isDragging, setIsDragging] = useState(false)
    const [isLoading, setIsLoading] = useState(true)
    const fileInputRef = useRef<HTMLInputElement>(null)
    const dragCounter = useRef(0)

    // Load saved log on mount
    useEffect(() => {
        getLogFromDB().then((savedContent) => {
            if (savedContent) {
                setContent(savedContent)
                toast.success("Restored previous scan log")
            }
            setIsLoading(false)
        })
    }, [])

    const readFile = useCallback((file: File) => {
        const reader = new FileReader()
        reader.onload = async (e) => {
            const result = e.target?.result as string
            setContent(result)
            toast.promise(saveLogToDB(result), {
                loading: "Saving log locally...",
                success: "Log saved for future sessions",
                error: "Failed to save log",
            })
        }
        reader.readAsText(file)
    }, [])

    const handleClear = useCallback(async () => {
        await clearLogFromDB()
        setContent("")
        toast.success("Cleared saved log")
    }, [])

    const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        const file = event.target.files?.[0]
        if (file) readFile(file)
    }

    const handleDragEnter = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
        dragCounter.current++
        setIsDragging(true)
    }, [])

    const handleDragLeave = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
        dragCounter.current--
        if (dragCounter.current === 0) {
            setIsDragging(false)
        }
    }, [])

    const handleDragOver = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
    }, [])

    const handleDrop = useCallback((e: React.DragEvent) => {
        e.preventDefault()
        e.stopPropagation()
        setIsDragging(false)
        dragCounter.current = 0

        const file = e.dataTransfer.files?.[0]
        if (file) readFile(file)
    }, [readFile])

    if (isLoading) {
        return (
            <div className="flex items-center justify-center min-h-screen text-[--muted]">
                <p>Loading saved logs...</p>
            </div>
        )
    }

    return (
        <div
            className="container mx-auto p-4 min-h-screen relative"
            onDragEnter={handleDragEnter}
            onDragLeave={handleDragLeave}
            onDragOver={handleDragOver}
            onDrop={handleDrop}
        >
            {isDragging && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-gray-950/80 backdrop-blur-sm">
                    <div className="flex flex-col items-center gap-3 p-8 border-2 border-dashed border-indigo-500 rounded-xl bg-gray-900/50">
                        <BiUpload className="text-4xl text-indigo-400" />
                        <p className="text-lg font-medium text-indigo-300">Drop log file</p>
                    </div>
                </div>
            )}

            <div className="mb-4">
                <div className="flex items-center gap-4 justify-between">
                    <div className="flex items-center gap-4">
                        <h1 className="text-xl font-bold text-gray-200 tracking-tight">Scan Log Analyzer</h1>
                        <label
                            className="flex items-center gap-2 px-3 py-1.5 bg-gray-800 border border-[--border] rounded-md cursor-pointer hover:bg-gray-700 transition-colors text-sm text-gray-300"
                        >
                            <BiUpload />
                            <span>{content ? "Load another file" : "Load scan log file"}</span>
                            <input
                                type="file"
                                ref={fileInputRef}
                                onChange={handleFileChange}
                                accept=".log,.txt"
                                className="hidden"
                            />
                        </label>
                    </div>

                    {content && (
                        <button
                            onClick={handleClear}
                            className="flex items-center gap-2 px-3 py-1.5 text-sm text-red-400 hover:text-red-300 hover:bg-red-950/30 rounded-md transition-colors"
                        >
                            <BiTrash />
                            Clear log
                        </button>
                    )}
                </div>
            </div>
            <ScanLogViewer content={content} />
        </div>
    )
}
