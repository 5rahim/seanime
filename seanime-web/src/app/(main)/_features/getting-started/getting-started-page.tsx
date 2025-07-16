import { Status } from "@/api/generated/types"
import { useGettingStarted } from "@/api/hooks/settings.hooks"
import { useSetServerStatus } from "@/app/(main)/_hooks/use-server-status"
import { GlowingEffect } from "@/components/shared/glowing-effect"
import { LoadingOverlayWithLogo } from "@/components/shared/loading-overlay-with-logo"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Card, CardProps } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { Field, Form } from "@/components/ui/form"
import {
    DEFAULT_TORRENT_PROVIDER,
    getDefaultIinaSocket,
    getDefaultMpvSocket,
    getDefaultSettings,
    gettingStartedSchema,
    TORRENT_PROVIDER,
    useDefaultSettingsPaths,
} from "@/lib/server/settings"
import { AnimatePresence, motion } from "motion/react"
import { useRouter } from "next/navigation"
import React from "react"
import { useFormContext, useWatch } from "react-hook-form"
import { BiChevronLeft, BiChevronRight, BiCloud, BiCog, BiDownload, BiFolder, BiPlay, BiRocket } from "react-icons/bi"
import { FaBook, FaDiscord } from "react-icons/fa"
import { HiOutlineDesktopComputer } from "react-icons/hi"
import { HiEye, HiGlobeAlt } from "react-icons/hi2"
import { ImDownload } from "react-icons/im"
import { IoPlayForwardCircleSharp } from "react-icons/io5"
import { MdOutlineBroadcastOnHome } from "react-icons/md"
import { RiFolderDownloadFill } from "react-icons/ri"
import { SiMpv, SiQbittorrent, SiTransmission, SiVlcmediaplayer } from "react-icons/si"

const containerVariants = {
    hidden: { opacity: 0 },
    visible: {
        opacity: 1,
        transition: {
            staggerChildren: 0.05,
            delayChildren: 0.1,
        },
    },
    exit: {
        opacity: 0,
        transition: {
            staggerChildren: 0.03,
            staggerDirection: -1,
        },
    },
}

const itemVariants = {
    hidden: {
        opacity: 0,
        y: 10,
    },
    visible: {
        opacity: 1,
        y: 0,
    },
    exit: {
        opacity: 0,
        y: -10,
    },
}

const stepVariants = {
    enter: (direction: number) => ({
        x: direction > 0 ? 40 : -40,
        opacity: 0,
    }),
    center: {
        zIndex: 1,
        x: 0,
        opacity: 1,
    },
    exit: (direction: number) => ({
        zIndex: 0,
        x: direction < 0 ? 40 : -40,
        opacity: 0,
    }),
}

const STEPS = [
    {
        id: "library",
        title: "Anime Library",
        description: "Choose your anime collection folder",
        icon: BiFolder,
        gradient: "from-blue-500 to-cyan-500",
    },
    {
        id: "player",
        title: "Media Player",
        description: "Configure your video player",
        icon: BiPlay,
        gradient: "from-green-500 to-emerald-500",
    },
    {
        id: "torrents",
        title: "Torrent Setup",
        description: "Set up downloading and providers",
        icon: BiDownload,
        gradient: "from-orange-500 to-red-500",
    },
    {
        id: "debrid",
        title: "Debrid Service",
        description: "Optional premium streaming",
        icon: BiCloud,
        gradient: "from-indigo-500 to-purple-500",
    },
    {
        id: "features",
        title: "Features",
        description: "Enable additional features",
        icon: BiCog,
        gradient: "from-teal-500 to-blue-500",
    },
]

function StepIndicator({ currentStep, totalSteps, onStepClick }: { currentStep: number; totalSteps: number; onStepClick: (step: number) => void }) {
    return (
        <div className="mb-12">
            <div className="flex items-center justify-center mb-6">
                <div className="relative mx-auto w-24 h-24">
                    <motion.img
                        src="/logo_2.png"
                        alt="Seanime Logo"
                        className="w-full h-full object-contain"
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        transition={{ duration: 0.3 }}
                    />
                </div>
            </div>

            <div className="text-center mb-8">
                <p className="text-[--muted] text-sm ">
                    These settings can be changed later
                </p>
            </div>

            <div className="flex items-start justify-between max-w-4xl mx-auto px-4 border p-4 rounded-lg relative bg-gray-900/50 backdrop-blur-sm">
                <GlowingEffect
                    spread={40}
                    glow={true}
                    disabled={false}
                    proximity={100}
                    inactiveZone={0.01}
                    // movementDuration={4}
                    className="opacity-30"
                />

                {STEPS.map((step, i) => (
                    <div
                        key={step.id}
                        onClick={(e) => {
                            onStepClick(i)
                        }}
                        className={cn("flex flex-col items-center relative group transition-all duration-200 focus:outline-none rounded-lg p-2 w-36",
                            "cursor-pointer")}
                    >
                        <motion.div
                            className={cn(
                                "w-12 h-12 rounded-full flex items-center justify-center mb-3 transition-all duration-200",
                                // i <= currentStep
                                //     ? `bg-gradient-to-r ${step.gradient} text-white`
                                //     : "bg-gray-700 text-gray-500",
                                i <= currentStep
                                    ? "bg-gradient-to-br from-brand-500/20 to-purple-500/20 border border-brand-500/20"
                                    : "bg-[--subtle] text-[--muted]",
                                i <= currentStep && "group-hover:shadow-md",
                            )}
                            initial={{ scale: 0.9 }}
                            animate={{
                                scale: i === currentStep ? 1.05 : 1,
                            }}
                            transition={{ duration: 0.2 }}
                        >
                            <step.icon className="w-6 h-6" />
                        </motion.div>

                        <div className="text-center">
                            <h3
                                className={cn(
                                    "text-sm font-medium transition-colors duration-200 tracking-wide",
                                    i <= currentStep ? "text-white" : "text-[--muted]",
                                    "group-hover:text-[--brand]",
                                )}
                            >
                                {step.title}
                            </h3>
                            {/* <p className="text-xs text-gray-500 mt-1 max-w-20">
                             {step.description}
                             </p> */}
                        </div>

                        {/* {i < STEPS.length - 1 && (
                         <div className="absolute top-8 left-full w-[40%] h-0.5 -translate-y-0 hidden md:block">
                         <div className={cn(
                         "h-full transition-all duration-300",
                         i < currentStep
                         ? "bg-[--subtle]"
                         : "bg-gray-600"
                         )} />
                         </div>
                         )} */}
                    </div>
                ))}
            </div>
        </div>
    )
}

function StepCard({ children, className, ...props }: CardProps) {
    return (
        <motion.div
            variants={itemVariants}
            className={cn(
                "relative rounded-xl bg-gray-900/50 backdrop-blur-sm border",
                className,
            )}
        >
            <GlowingEffect
                spread={40}
                glow={true}
                disabled={false}
                proximity={100}
                inactiveZone={0.01}
                // movementDuration={4}
                className="opacity-30"
            />
            <Card className="bg-transparent border-none shadow-none p-6">
                {children}
            </Card>
        </motion.div>
    )
}


function LibraryStep({ form }: { form: any }) {
    return (
        <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="space-y-8"
        >
            <motion.div variants={itemVariants} className="text-center space-y-4">
                <h2 className="text-3xl font-bold">Anime Library</h2>
                <p className="text-[--muted] text-sm max-w-lg mx-auto">
                    Choose the folder where your anime files are stored. This is where Seanime will scan for your collection.
                </p>
            </motion.div>

            <StepCard className="max-w-2xl mx-auto">
                <motion.div variants={itemVariants}>
                    <Field.DirectorySelector
                        name="libraryPath"
                        label="Anime Library Path"
                        leftIcon={<BiFolder className="text-blue-500" />}
                        shouldExist
                        help="Select the main folder containing your anime collection. You can add more folders later."
                        className="w-full"
                    />
                </motion.div>
            </StepCard>

        </motion.div>
    )
}

function PlayerStep({ form, status }: { form: any, status: Status }) {
    const { watch } = useFormContext()
    const defaultPlayer = useWatch({ name: "defaultPlayer" })

    return (
        <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="space-y-8"
        >
            <motion.div variants={itemVariants} className="text-center space-y-4">
                <h2 className="text-3xl font-bold">Media Player</h2>
                <p className="text-[--muted] text-sm max-w-lg mx-auto">
                    Configure your preferred media player for watching anime and tracking progress automatically.
                </p>
            </motion.div>

            <StepCard className="max-w-2xl mx-auto">
                <motion.div variants={itemVariants} className="space-y-6">
                    <Field.Select
                        name="defaultPlayer"
                        label="Media Player"
                        help={status?.os !== "darwin"
                            ? "MPV is recommended for better subtitle rendering, torrent streaming."
                            : "Both MPV and IINA are recommended for macOS."}
                        required
                        leftIcon={<BiPlay className="text-green-500" />}
                        options={[
                            { label: "MPV (Recommended)", value: "mpv" },
                            { label: "VLC", value: "vlc" },
                            ...(status?.os === "windows" ? [{ label: "MPC-HC", value: "mpc-hc" }] : []),
                            ...(status?.os === "darwin" ? [{ label: "IINA", value: "iina" }] : []),
                        ]}
                    />

                    <AnimatePresence mode="wait">
                        {defaultPlayer === "mpv" && (
                            <>
                                <p>
                                    On Windows, install MPV easily using Scoop or Chocolatey. On macOS, install MPV using Homebrew.
                                </p>
                                <motion.div
                                    key="mpv"
                                    initial={{ opacity: 0, height: 0 }}
                                    animate={{ opacity: 1, height: "auto" }}
                                    exit={{ opacity: 0, height: 0 }}
                                    className="space-y-4 p-4 rounded-lg bg-gray-800/30"
                                >
                                    <div className="flex items-center space-x-3">
                                        <SiMpv className="w-6 h-6 text-purple-400" />
                                        <h4 className="font-semibold">MPV Configuration</h4>
                                    </div>
                                    <Field.Text
                                        name="mpvSocket"
                                        label="Socket / Pipe Path"
                                        help="Path for MPV IPC communication"
                                    />
                                </motion.div>
                            </>
                        )}

                        {defaultPlayer === "iina" && (
                            <motion.div
                                key="iina"
                                initial={{ opacity: 0, height: 0 }}
                                animate={{ opacity: 1, height: "auto" }}
                                exit={{ opacity: 0, height: 0 }}
                                className="space-y-4 p-4 rounded-lg bg-gray-800/30"
                            >
                                <div className="flex items-center space-x-3">
                                    <IoPlayForwardCircleSharp className="w-6 h-6 text-blue-400" />
                                    <h4 className="font-semibold">IINA Configuration</h4>
                                </div>
                                <Field.Text
                                    name="iinaSocket"
                                    label="Socket / Pipe Path"
                                    help="Path for IINA IPC communication"
                                />

                                <Alert
                                    intent="info-basic"
                                    description={<p>For IINA to work correctly with Seanime, make sure <strong>Quit after all windows are
                                                                                                               closed</strong> is <span
                                        className="underline"
                                    >checked</span> and <strong>Keep window open after playback
                                                                finishes</strong> is <span className="underline">unchecked</span> in
                                                    your IINA general settings.</p>}
                                />
                            </motion.div>
                        )}

                        {defaultPlayer === "vlc" && (
                            <motion.div
                                key="vlc"
                                initial={{ opacity: 0, height: 0 }}
                                animate={{ opacity: 1, height: "auto" }}
                                exit={{ opacity: 0, height: 0 }}
                                className="space-y-4 p-4 rounded-lg bg-gray-800/30"
                            >
                                <div className="flex items-center space-x-3">
                                    <SiVlcmediaplayer className="w-6 h-6 text-orange-500" />
                                    <h4 className="font-semibold">VLC Configuration</h4>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <Field.Text name="mediaPlayerHost" label="Host" />
                                    <Field.Number name="vlcPort" label="Port" formatOptions={{ useGrouping: false }} />
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <Field.Text name="vlcUsername" label="Username" />
                                    <Field.Text name="vlcPassword" label="Password" />
                                </div>
                                <Field.Text name="vlcPath" label="VLC Executable Path" />
                            </motion.div>
                        )}

                        {defaultPlayer === "mpc-hc" && (
                            <motion.div
                                key="mpc-hc"
                                initial={{ opacity: 0, height: 0 }}
                                animate={{ opacity: 1, height: "auto" }}
                                exit={{ opacity: 0, height: 0 }}
                                className="space-y-4 p-4 rounded-lg bg-gray-800/30"
                            >
                                <div className="flex items-center space-x-3">
                                    <HiOutlineDesktopComputer className="w-6 h-6 text-blue-500" />
                                    <h4 className="font-semibold">MPC-HC Configuration</h4>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <Field.Text name="mediaPlayerHost" label="Host" />
                                    <Field.Number name="mpcPort" label="Port" formatOptions={{ useGrouping: false }} />
                                </div>
                                <Field.Text name="mpcPath" label="MPC-HC Executable Path" />
                            </motion.div>
                        )}
                    </AnimatePresence>
                </motion.div>
            </StepCard>
        </motion.div>
    )
}

function TorrentStep({ form }: { form: any }) {
    const { watch } = useFormContext()
    const defaultTorrentClient = useWatch({ name: "defaultTorrentClient" })

    return (
        <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="space-y-8"
        >
            <motion.div variants={itemVariants} className="text-center space-y-4">
                <h2 className="text-3xl font-bold">Torrent Setup</h2>
                <p className="text-[--muted] text-sm max-w-lg mx-auto">
                    Configure your default torrent provider and client.
                </p>
            </motion.div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 max-w-6xl mx-auto">
                <StepCard>
                    <motion.div variants={itemVariants} className="space-y-4">
                        <div className="flex items-center space-x-3 mb-4">
                            <RiFolderDownloadFill className="w-6 h-6 text-orange-500" />
                            <h3 className="text-xl font-semibold">Torrent Provider</h3>
                        </div>
                        <p className="text-sm text-[--muted]">
                            Extension for finding anime torrents
                        </p>
                        <Field.Select
                            name="torrentProvider"
                            label="Provider"
                            required
                            options={[
                                { label: "AnimeTosho (Recommended)", value: TORRENT_PROVIDER.ANIMETOSHO },
                                { label: "Nyaa", value: TORRENT_PROVIDER.NYAA },
                                { label: "Nyaa (Non-English)", value: TORRENT_PROVIDER.NYAA_NON_ENG },
                            ]}
                            help="AnimeTosho search results are more precise in most cases."
                        />
                    </motion.div>
                </StepCard>

                <StepCard>
                    <motion.div variants={itemVariants} className="space-y-4">
                        <div className="flex items-center space-x-3 mb-4">
                            <ImDownload className="w-6 h-6 text-blue-500" />
                            <h3 className="text-xl font-semibold">Torrent Client</h3>
                        </div>
                        <p className="text-sm text-[--muted]">
                            Client used to download anime torrents
                        </p>
                        <Field.Select
                            name="defaultTorrentClient"
                            label="Client"
                            options={[
                                { label: "qBittorrent", value: "qbittorrent" },
                                { label: "Transmission", value: "transmission" },
                                { label: "None", value: "none" },
                            ]}
                        />
                    </motion.div>
                </StepCard>
            </div>

            <AnimatePresence mode="wait">
                {(defaultTorrentClient === "qbittorrent" || defaultTorrentClient === "transmission") && (
                    <StepCard className="max-w-4xl mx-auto">
                        <motion.div
                            key={defaultTorrentClient}
                            initial={{ opacity: 0, scale: 0.95 }}
                            animate={{ opacity: 1, scale: 1 }}
                            exit={{ opacity: 0, scale: 0.95 }}
                            className="space-y-6"
                        >
                            {defaultTorrentClient === "qbittorrent" && (
                                <>
                                    <div className="flex items-center space-x-3">
                                        <SiQbittorrent className="w-8 h-8 text-blue-600" />
                                        <h4 className="text-xl font-semibold">qBittorrent Settings</h4>
                                    </div>
                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                        <Field.Text name="qbittorrentHost" label="Host" />
                                        <Field.Text name="qbittorrentUsername" label="Username" />
                                        <Field.Text name="qbittorrentPassword" label="Password" />
                                    </div>
                                    <div className="grid grid-cols-2 gap-4 lg:grid-cols-[200px_1fr]">
                                        <Field.Number name="qbittorrentPort" label="Port" formatOptions={{ useGrouping: false }} />
                                        <Field.Text name="qbittorrentPath" label="Executable Path" />
                                    </div>
                                </>
                            )}

                            {defaultTorrentClient === "transmission" && (
                                <>
                                    <div className="flex items-center space-x-3">
                                        <SiTransmission className="w-8 h-8 text-red-600" />
                                        <h4 className="text-xl font-semibold">Transmission Settings</h4>
                                    </div>
                                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                                        <Field.Text name="transmissionHost" label="Host" />
                                        <Field.Text name="transmissionUsername" label="Username" />
                                        <Field.Text name="transmissionPassword" label="Password" />
                                    </div>
                                    <div className="grid grid-cols-2 gap-4 lg:grid-cols-[200px_1fr]">
                                        <Field.Number name="transmissionPort" label="Port" formatOptions={{ useGrouping: false }} />
                                        <Field.Text name="transmissionPath" label="Executable Path" />
                                    </div>
                                </>
                            )}
                        </motion.div>
                    </StepCard>
                )}
            </AnimatePresence>
        </motion.div>
    )
}

function DebridStep({ form }: { form: any }) {
    const debridProvider = useWatch({ name: "debridProvider" })

    return (
        <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="space-y-8"
        >
            <motion.div variants={itemVariants} className="text-center space-y-4">
                <h2 className="text-3xl font-bold">Debrid Service</h2>
                <p className="text-[--muted] text-sm max-w-lg mx-auto">
                    Debrid services offer faster downloads and instant streaming from the cloud.
                </p>
            </motion.div>

            <StepCard className="max-w-2xl mx-auto">
                <motion.div variants={itemVariants} className="space-y-6">
                    <Field.Select
                        name="debridProvider"
                        label="Debrid Service"
                        leftIcon={<BiCloud className="text-[--purple]" />}
                        options={[
                            { label: "None", value: "none" },
                            { label: "TorBox", value: "torbox" },
                            { label: "Real-Debrid", value: "realdebrid" },
                        ]}
                    />

                    <AnimatePresence>
                        {debridProvider !== "none" && debridProvider !== "" && (
                            <motion.div
                                initial={{ opacity: 0, height: 0 }}
                                animate={{ opacity: 1, height: "auto" }}
                                exit={{ opacity: 0, height: 0 }}
                                className="space-y-4 p-4 rounded-lg bg-gray-800/30"
                            >
                                <Field.Text
                                    name="debridApiKey"
                                    label="API Key"
                                    help="The API key provided by the debrid service."
                                />
                            </motion.div>
                        )}
                    </AnimatePresence>
                </motion.div>
            </StepCard>
        </motion.div>
    )
}

function FeaturesStep({ form }: { form: any }) {
    const features = [
        {
            name: "enableManga",
            icon: FaBook,
            title: "Manga",
            description: "Read and download manga chapters",
            gradient: "from-orange-500 to-yellow-700",
        },
        {
            name: "enableTorrentStreaming",
            icon: BiDownload,
            title: "Torrent Streaming",
            description: "Stream torrents without waiting for download",
            gradient: "from-cyan-500 to-teal-500",
        },
        {
            name: "enableAdultContent",
            icon: HiEye,
            title: "NSFW Content",
            description: "Show adult content in library and search",
            gradient: "from-red-500 to-pink-500",
        },
        {
            name: "enableOnlinestream",
            icon: HiGlobeAlt,
            title: "Online Streaming",
            description: "Watch anime from online sources",
            gradient: "from-purple-500 to-violet-500",
        },
        {
            name: "enableRichPresence",
            icon: FaDiscord,
            title: "Discord Rich Presence",
            description: "Show what you're watching on Discord",
            gradient: "from-indigo-500 to-blue-500",
        },
        {
            name: "enableTranscode",
            icon: MdOutlineBroadcastOnHome,
            title: "Transcoding / Direct Play",
            description: "Stream downloaded files on other devices",
            gradient: "from-cyan-500 to-indigo-500",
        },
    ]

    return (
        <motion.div
            variants={containerVariants}
            initial="hidden"
            animate="visible"
            exit="exit"
            className="space-y-8"
        >
            <motion.div variants={itemVariants} className="text-center space-y-4">
                <h2 className="text-3xl font-bold">Additional Features</h2>
                <p className="text-[--muted] text-sm max-w-lg mx-auto">
                    Choose which additional features you'd like to enable. You can enable or disable these later in settings.
                </p>
            </motion.div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 max-w-6xl mx-auto">
                {features.map((feature, index) => (
                    <motion.div
                        key={feature.name}
                        variants={itemVariants}
                        custom={index}
                    >
                        <Field.Checkbox
                            name={feature.name}
                            label={
                                <div className="flex items-start space-x-4 p-4">
                                    <div
                                        className={cn(
                                            "w-12 h-12 rounded-lg flex items-center justify-center",
                                            `bg-gradient-to-br ${feature.gradient}`,
                                        )}
                                    >
                                        <feature.icon className="w-6 h-6 text-white" />
                                    </div>
                                    <div className="flex-1 min-w-0">
                                        <h3 className="font-semibold text-sm">{feature.title}</h3>
                                        <p className="text-xs text-gray-400 mt-1 leading-relaxed">
                                            {feature.description}
                                        </p>
                                    </div>
                                </div>
                            }
                            size="lg"
                            labelClass={cn(
                                "block cursor-pointer transition-all duration-200 overflow-hidden w-full rounded-xl",
                                "bg-gray-900/50 hover:bg-gray-800/80",
                                "border border-gray-700/50",
                                "hover:border-gray-600",
                                // "hover:shadow-lg hover:scale-[1.02]",
                                "data-[checked=true]:bg-gradient-to-br data-[checked=true]:from-gray-900 data-[checked=true]:to-gray-900",
                                "data-[checked=true]:border-brand-600",
                                // "data-[checked=true]:shadow-lg data-[checked=true]:scale-[1.02]"
                            )}
                            containerClass="flex items-center justify-between h-full"
                            className="absolute top-2 right-2 z-10"
                            fieldClass="relative"
                        />
                    </motion.div>
                ))}
            </div>
        </motion.div>
    )
}

export function GettingStartedPage({ status }: { status: Status }) {
    const router = useRouter()
    const { getDefaultVlcPath, getDefaultQBittorrentPath, getDefaultTransmissionPath } = useDefaultSettingsPaths()
    const setServerStatus = useSetServerStatus()

    const { mutate, data, isPending, isSuccess } = useGettingStarted()

    const [currentStep, setCurrentStep] = React.useState(0)
    const [direction, setDirection] = React.useState(0)

    /**
     * If the settings are returned, redirect to the home page
     */
    React.useEffect(() => {
        if (!isPending && !!data?.settings) {
            setServerStatus(data)
            router.push("/")
        }
    }, [data, isPending])

    const vlcDefaultPath = React.useMemo(() => getDefaultVlcPath(status.os), [status.os])
    const qbittorrentDefaultPath = React.useMemo(() => getDefaultQBittorrentPath(status.os), [status.os])
    const transmissionDefaultPath = React.useMemo(() => getDefaultTransmissionPath(status.os), [status.os])
    const mpvSocketPath = React.useMemo(() => getDefaultMpvSocket(status.os), [status.os])
    const iinaSocketPath = React.useMemo(() => getDefaultIinaSocket(status.os), [status.os])

    const nextStep = () => {
        if (currentStep < STEPS.length - 1) {
            setDirection(1)
            setCurrentStep(currentStep + 1)
        }
    }

    const prevStep = () => {
        if (currentStep > 0) {
            setDirection(-1)
            setCurrentStep(currentStep - 1)
        }
    }

    const goToStep = (step: number) => {
        if (step >= 0 && step < STEPS.length) {
            setDirection(step > currentStep ? 1 : -1)
            setCurrentStep(step)
        }
    }

    if (isPending) return <LoadingOverlayWithLogo />

    if (!data) return (
        <div className="min-h-screen bg-gradient-to-br from-[--background] via-[--background] to-purple-950/10">
            <div className="absolute inset-0 overflow-hidden pointer-events-none">
                {/* <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-blue-500/10 rounded-full blur-3xl" /> */}
                {/* <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-purple-500/10 rounded-full blur-3xl" /> */}
                {/* <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-pink-500/5 rounded-full blur-3xl" /> */}
            </div>

            <div className="container max-w-6xl mx-auto px-4 py-8 relative z-10">
                <Form
                    schema={gettingStartedSchema}
                    onSubmit={data => {
                        if (currentStep === STEPS.length - 1) {
                            mutate(getDefaultSettings(data))
                        } else {
                            nextStep()
                        }
                    }}
                    defaultValues={{
                        mediaPlayerHost: "127.0.0.1",
                        vlcPort: 8080,
                        mpcPort: 13579,
                        defaultPlayer: "mpv",
                        vlcPath: vlcDefaultPath,
                        qbittorrentPath: qbittorrentDefaultPath,
                        qbittorrentHost: "127.0.0.1",
                        qbittorrentPort: 8081,
                        transmissionPath: transmissionDefaultPath,
                        transmissionHost: "127.0.0.1",
                        transmissionPort: 9091,
                        mpcPath: "C:/Program Files/MPC-HC/mpc-hc64.exe",
                        torrentProvider: DEFAULT_TORRENT_PROVIDER,
                        mpvSocket: mpvSocketPath,
                        iinaSocket: iinaSocketPath,
                        enableRichPresence: false,
                        autoScan: false,
                        enableManga: true,
                        enableOnlinestream: false,
                        enableAdultContent: true,
                        enableTorrentStreaming: true,
                        enableTranscode: false,
                        debridProvider: "none",
                        debridApiKey: "",
                        nakamaUsername: "",
                        enableWatchContinuity: true,
                    }}
                >
                    {(f) => (
                        <div className="space-y-8">
                            <StepIndicator currentStep={currentStep} totalSteps={STEPS.length} onStepClick={goToStep} />

                            <AnimatePresence mode="wait" custom={direction}>
                                <motion.div
                                    key={currentStep}
                                    custom={direction}
                                    variants={stepVariants}
                                    initial="enter"
                                    animate="center"
                                    exit="exit"
                                    transition={{
                                        x: { duration: 0.3, ease: "easeInOut" },
                                        opacity: { duration: 0.2 },
                                    }}
                                    className=""
                                >
                                    {currentStep === 0 && <LibraryStep form={f} />}
                                    {currentStep === 1 && <PlayerStep form={f} status={status} />}
                                    {currentStep === 2 && <TorrentStep form={f} />}
                                    {currentStep === 3 && <DebridStep form={f} />}
                                    {currentStep === 4 && <FeaturesStep form={f} />}
                                </motion.div>
                            </AnimatePresence>

                            <motion.div
                                className="flex justify-between items-center max-w-2xl mx-auto pt-8"
                                initial={{ opacity: 0, y: 20 }}
                                animate={{ opacity: 1, y: 0 }}
                                transition={{ delay: 0.5 }}
                            >
                                <Button
                                    type="button"
                                    intent="gray-outline"
                                    onClick={e => {
                                        e.preventDefault()
                                        prevStep()
                                    }}
                                    disabled={currentStep === 0}
                                    className="flex items-center space-x-2"
                                    leftIcon={<BiChevronLeft />}
                                >
                                    Previous
                                </Button>

                                {currentStep === STEPS.length - 1 ? (
                                    <Button
                                        type="submit"
                                        className="flex items-center bg-gradient-to-r from-brand-600 to-indigo-600 hover:ring-2 ring-brand-600"
                                        loading={isPending}
                                        rightIcon={<BiRocket className="size-6" />}
                                    >
                                        <span>Launch Seanime</span>
                                    </Button>
                                ) : (
                                    <Button
                                        type="button"
                                        intent="primary-subtle"
                                        onClick={e => {
                                            e.preventDefault()
                                            nextStep()
                                        }}
                                        className="flex items-center space-x-2"
                                        rightIcon={<BiChevronRight />}
                                    >
                                        Next
                                    </Button>
                                )}
                            </motion.div>
                        </div>
                    )}
                </Form>

                <motion.p
                    className="text-center text-[--muted] mt-12"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ delay: 1 }}
                >
                    Made by 5rahim
                </motion.p>
            </div>
        </div>
    )
}
