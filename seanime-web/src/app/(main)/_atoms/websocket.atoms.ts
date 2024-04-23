import { atom } from "jotai"
import { createContext } from "react"

export const WebSocketContext = createContext<WebSocket | null>(null)

export const websocketAtom = atom<WebSocket | null>(null)

