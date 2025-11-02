"use client"

import { Button, IconButton } from "@/components/ui/button"
import { cn } from "@/components/ui/core/styling"
import { __isDesktop__ } from "@/types/constants"
import React, { useRef, useState } from "react"
import { BiX } from "react-icons/bi"
import { RiPlayList2Fill } from "react-icons/ri"

export type ContainerPosition = "bottom-right" | "bottom-left";
export type ContainerSize = "sm" | "md" | "lg" | "xl" | "full";

const containerConfig = {
    dimensions: {
        sm: "sm:max-w-sm sm:max-h-[500px]",
        md: "sm:max-w-md sm:max-h-[600px]",
        lg: "sm:max-w-lg sm:max-h-[700px]",
        xl: "sm:max-w-xl sm:max-h-[800px]",
        full: "sm:w-full sm:h-full",
    },
    positions: {
        "bottom-right": "bottom-5 right-5",
        "bottom-left": "bottom-5 left-5",
    },
    containerPositions: {
        "bottom-right": "sm:bottom-[calc(100%+10px)] sm:right-0",
        "bottom-left": "sm:bottom-[calc(100%+10px)] sm:left-0",
    },
    states: {
        open: "pointer-events-auto opacity-100 visible scale-100 translate-y-0",
        closed:
            "pointer-events-none opacity-0 invisible scale-100 sm:translate-y-5",
    },
}

interface PlaylistManagerPopupProps extends React.HTMLAttributes<HTMLDivElement> {
    position?: ContainerPosition;
    size?: ContainerSize;
    icon?: React.ReactNode;
}

const PlaylistManagerPopup: React.FC<PlaylistManagerPopupProps> = ({
    className,
    position = "bottom-right",
    size = "md",
    icon,
    children,
    ...props
}) => {
    const [isOpen, setIsOpen] = useState(false)
    const containerRef = useRef<HTMLDivElement>(null)

    const toggleContainer = () => setIsOpen(!isOpen)

    return (
        <div
            className={cn(`fixed ${containerConfig.positions[position]} z-[25]`, className)}
            {...props}
        >
            <div
                ref={containerRef}
                className={cn(
                    "flex flex-col bg-[--paper] overflow-hidden border sm:rounded-2xl shadow-md transition-all duration-250 ease-out sm:absolute sm:w-[90vw] sm:h-[80vh] fixed inset-0 w-full h-full sm:inset-auto",
                    containerConfig.containerPositions[position],
                    containerConfig.dimensions[size],
                    isOpen ? containerConfig.states.open : containerConfig.states.closed,
                    __isDesktop__ && "pt-8 sm:pt-0",
                    className,
                )}
            >
                {children}
                <IconButton
                    intent="white"
                    size="sm"
                    className={cn(
                        "absolute top-2 right-2 sm:hidden rounded-full",
                        __isDesktop__ && "top-8",
                    )}
                    onClick={toggleContainer}
                    icon={<BiX className="h-6 w-6" />}
                />
            </div>
            <PlaylistManagerPopupToggle
                icon={icon}
                isOpen={isOpen}
                toggleContainer={toggleContainer}
            />
        </div>
    )
}

PlaylistManagerPopup.displayName = "PlaylistManagerPopup"

const PlaylistManagerPopupHeader: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
    className,
    ...props
}) => (
    <div
        className={cn("flex items-center justify-between p-4 border-b", className)}
        {...props}
    />
)

PlaylistManagerPopupHeader.displayName = "PlaylistManagerPopupHeader"

const PlaylistManagerPopupBody: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
    className,
    ...props
}) => <div className={cn("flex-grow overflow-y-auto", className)} {...props} />

PlaylistManagerPopupBody.displayName = "PlaylistManagerPopupBody"

const PlaylistManagerPopupFooter: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
    className,
    ...props
}) => <div className={cn("border-t p-4", className)} {...props} />

PlaylistManagerPopupFooter.displayName = "PlaylistManagerPopupFooter"

interface PlaylistManagerPopupToggleProps
    extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    icon?: React.ReactNode;
    isOpen: boolean;
    toggleContainer: () => void;
}

const PlaylistManagerPopupToggle: React.FC<PlaylistManagerPopupToggleProps> = ({
    className,
    icon,
    isOpen,
    toggleContainer,
    ...props
}) => (
    <Button
        intent="white"
        onClick={toggleContainer}
        className={cn(
            "shadow-xl w-14 h-14 rounded-full flex items-center justify-center hover:shadow-lg hover:shadow-black/30 transition-all duration-300",
            className,
        )}
        {...props}
    >
        {isOpen ? (
            <BiX
                className={cn(
                    "h-6 w-6",
                )}
            />
        ) : (
            icon || <RiPlayList2Fill className="h-6 w-6" />
        )}
    </Button>
)

PlaylistManagerPopupToggle.displayName = "PlaylistManagerPopupToggle"

export {
    PlaylistManagerPopup,
    PlaylistManagerPopupHeader,
    PlaylistManagerPopupBody,
    PlaylistManagerPopupFooter,
}
