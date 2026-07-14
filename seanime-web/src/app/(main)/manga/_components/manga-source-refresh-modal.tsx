import { Manga_MangaSourceRefreshJob, Manga_MangaSourceRefreshMode } from "@/api/generated/types"
import { useListMangaProviderExtensions } from "@/api/hooks/extensions.hooks"
import { useStartMangaSourceRefresh, useStopMangaSourceRefresh } from "@/api/hooks/manga.hooks"
import { __manga_preferencesHydratedAtom } from "@/app/(main)/manga/_lib/handle-manga-selected-provider"
import { SeaLink } from "@/components/shared/sea-link"
import { Alert } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { Disclosure, DisclosureContent, DisclosureItem, DisclosureTrigger } from "@/components/ui/disclosure"
import { Modal } from "@/components/ui/modal"
import { ProgressBar } from "@/components/ui/progress-bar"
import { RadioGroup } from "@/components/ui/radio-group"
import { useAtomValue } from "jotai/react"
import React from "react"
import { LuChevronDown, LuRefreshCcw } from "react-icons/lu"

type MangaSourceRefreshModalProps = {
    open: boolean
    onOpenChange: (open: boolean) => void
    job: Manga_MangaSourceRefreshJob | null | undefined
    returnFocusRef: React.RefObject<HTMLButtonElement | null>
}

const discoveryModes: Manga_MangaSourceRefreshMode[] = ["find_missing", "refresh_and_find", "reevaluate_all"]

export function MangaSourceRefreshModal({ open, onOpenChange, job, returnFocusRef }: MangaSourceRefreshModalProps) {
    const hydrated = useAtomValue(__manga_preferencesHydratedAtom)
    const { data: providers } = useListMangaProviderExtensions()
    const { mutate: startRefresh, isPending: isStarting } = useStartMangaSourceRefresh()
    const { mutate: stopRefresh, isPending: isStopping } = useStopMangaSourceRefresh()
    const [mode, setMode] = React.useState<Manga_MangaSourceRefreshMode>("refresh_selected")
    const statusHeadingRef = React.useRef<HTMLParagraphElement>(null)

    const terminal = job?.status === "completed" || job?.status === "cancelled" || job?.status === "failed"
    const running = job?.status === "running" || job?.status === "stopping"

    React.useEffect(() => {
        if (open && job) {
            statusHeadingRef.current?.focus()
        }
    }, [job?.id, job?.status, open])

    const providerCount = providers?.length ?? 0
    const discoveryDisabled = providerCount === 0
    const startDisabled = !hydrated || isStarting || discoveryDisabled && discoveryModes.includes(mode)
    const progress = job?.total ? Math.min(100, Math.round(job.current / job.total * 100)) : 0
    const failedMediaIds = [...new Set(job?.result.issues
        ?.filter(issue => issue.kind === "provider_error")
        .map(issue => issue.mediaId) ?? [])]
    const affectedMediaIds = [...new Set(job?.result.issues?.map(issue => issue.mediaId) ?? [])]
    const canFindAlternatives = providerCount > 0 &&
        (job?.mode === "refresh_selected" || job?.mode === "refresh_and_find") &&
        affectedMediaIds.length > 0

    const dismissJob = React.useCallback((runAgain: boolean) => {
        stopRefresh(undefined, {
            onSuccess: () => {
                if (runAgain) {
                    setMode(job?.mode ?? "refresh_selected")
                } else {
                    onOpenChange(false)
                }
            },
        })
    }, [job?.mode, onOpenChange, stopRefresh])

    const retryFailed = React.useCallback(() => {
        if (!job || !failedMediaIds.length) return
        stopRefresh(undefined, {
            onSuccess: () => startRefresh({ mode: job.mode, mediaIds: failedMediaIds }),
        })
    }, [failedMediaIds, job, startRefresh, stopRefresh])

    const findAlternatives = React.useCallback(() => {
        if (!affectedMediaIds.length) return
        stopRefresh(undefined, {
            onSuccess: () => startRefresh({ mode: "reevaluate_all", mediaIds: affectedMediaIds }),
        })
    }, [affectedMediaIds, startRefresh, stopRefresh])

    return (
        <Modal
            open={open}
            onOpenChange={onOpenChange}
            title="Refresh manga sources"
            contentClass="max-w-xl"
            onCloseAutoFocus={event => {
                event.preventDefault()
                returnFocusRef.current?.focus()
            }}
            footer={running ? (
                <>
                    <Button intent="gray-outline" onClick={() => onOpenChange(false)}>Close</Button>
                    <Button
                        intent="warning"
                        loading={job?.status === "stopping" || isStopping}
                        disabled={job?.status === "stopping" || isStopping}
                        onClick={() => stopRefresh()}
                    >
                        {job?.status === "stopping" ? "Stopping..." : "Stop refresh"}
                    </Button>
                </>
            ) : terminal ? (
                <>
                    {canFindAlternatives && (
                        <Button intent="gray-outline" loading={isStarting || isStopping} onClick={findAlternatives}>
                            Find alternatives
                        </Button>
                    )}
                    {!!failedMediaIds.length && (
                        <Button intent="gray-outline" loading={isStarting || isStopping} onClick={retryFailed}>
                            Retry failed
                        </Button>
                    )}
                    <Button intent="gray-outline" loading={isStopping} onClick={() => dismissJob(true)}>Run again</Button>
                    <Button intent="primary" loading={isStopping} onClick={() => dismissJob(false)}>Done</Button>
                </>
            ) : (
                <>
                    <Button intent="gray-outline" onClick={() => onOpenChange(false)}>Cancel</Button>
                    <Button
                        intent="primary"
                        leftIcon={<LuRefreshCcw />}
                        loading={isStarting}
                        disabled={startDisabled}
                        onClick={() => startRefresh({ mode })}
                    >Start refresh</Button>
                </>
            )}
        >
            {running && job ? (
                <div className="space-y-5" aria-live="polite">
                    <div className="space-y-2">
                        <div className="flex items-center justify-between gap-4 text-sm">
                            <p ref={statusHeadingRef} tabIndex={-1} className="font-medium outline-none">
                                {job.status === "stopping"
                                    ? "Stopping after the current request"
                                    : job.stage === "refreshing" ? "Refreshing selected sources" : "Searching installed sources"}
                            </p>
                            <p className="shrink-0 text-[--muted]">{job.current} of {job.total}</p>
                        </div>
                        <ProgressBar value={progress} size="sm" aria-label="Source refresh progress" />
                    </div>
                    <p className="text-sm text-[--muted]">
                        You can close this modal. The refresh will continue in the background.
                    </p>
                </div>
            ) : terminal && job ? (
                <div className="space-y-5">
                    <div aria-live={job.status === "failed" ? "assertive" : "polite"}>
                        <p ref={statusHeadingRef} tabIndex={-1} className="font-medium outline-none">
                            {job.status === "completed" ? "Source refresh complete" : job.status === "cancelled"
                                ? "Source refresh stopped"
                                : "Source refresh failed"}
                        </p>
                        <p className="mt-1 text-sm text-[--muted]">
                            {formatRefreshSummary(job)}
                        </p>
                    </div>

                    {!!job.error && <Alert intent="alert-basic" description={job.error} />}

                    {!!job.result.issues?.length && (
                        <Disclosure type="single" collapsible>
                            <DisclosureItem value="issues">
                                <DisclosureTrigger>
                                    <Button intent="gray-outline" className="w-full justify-between" rightIcon={<LuChevronDown />}>
                                        Review issues ({job.result.issues.length})
                                    </Button>
                                </DisclosureTrigger>
                                <DisclosureContent className="pt-3 max-h-56 overflow-y-auto">
                                    <div className="space-y-2">
                                        {job.result.issues.map(issue => (
                                            <div key={`${issue.mediaId}-${issue.kind}`} className="min-w-0 text-sm">
                                                <SeaLink href={`/manga/entry?id=${issue.mediaId}`} className="font-medium break-words">
                                                    {issue.title}
                                                </SeaLink>
                                                <p className="text-[--muted] break-words">
                                                    {issue.kind === "not_found"
                                                        ? job.mode === "refresh_selected"
                                                            ? "The saved source returned no chapters."
                                                            : "No matching source was found."
                                                        : "One or more providers failed."}
                                                    {!!issue.providers?.length && ` ${issue.providers.join(", ")}`}
                                                </p>
                                            </div>
                                        ))}
                                    </div>
                                </DisclosureContent>
                            </DisclosureItem>
                        </Disclosure>
                    )}
                </div>
            ) : (
                <div className="space-y-5">
                    <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm text-[--muted]">
                        <span>{providerCount} installed {providerCount === 1 ? "provider" : "providers"}</span>
                        <span>Current and re-reading manga only</span>
                    </div>

                    <RadioGroup
                        value={mode}
                        onValueChange={value => setMode(value as Manga_MangaSourceRefreshMode)}
                        stackClass="space-y-2"
                        itemContainerClass="items-start gap-3 rounded-xl border border-[--border] px-3 py-3 data-[state=checked]:border-brand/60 data-[state=checked]:bg-brand/5"
                        itemClass="mt-0.5 shrink-0"
                        itemLabelClass="min-w-0 flex-1"
                        options={[
                            {
                                value: "refresh_selected",
                                label: <ModeLabel title="Refresh selected sources" description="Update manga that already have a saved source." />,
                            },
                            {
                                value: "find_missing",
                                disabled: discoveryDisabled,
                                label: <ModeLabel
                                    title="Find missing sources"
                                    description="Search every installed provider for manga without a source."
                                />,
                            },
                            {
                                value: "refresh_and_find",
                                disabled: discoveryDisabled,
                                label: <ModeLabel
                                    title="Refresh and find missing"
                                    description="Update saved sources, then search for missing ones."
                                />,
                            },
                            {
                                value: "reevaluate_all",
                                disabled: discoveryDisabled,
                                label: <ModeLabel
                                    title="Re-evaluate all sources"
                                    description="Compare every installed provider and allow saved sources to change."
                                />,
                            },
                        ]}
                    />

                    {mode === "reevaluate_all" && (
                        <Alert
                            intent="warning-basic"
                            description="Existing source selections may be replaced when another provider has more distinct chapters."
                        />
                    )}
                    {!hydrated && (
                        <Alert intent="info-basic" description="Waiting for server-backed manga preferences to finish syncing." />
                    )}
                    {discoveryDisabled && (
                        <Alert intent="warning-basic" description="Install a manga provider to search for missing or alternative sources." />
                    )}
                </div>
            )}
        </Modal>
    )
}

function ModeLabel({ title, description }: { title: string, description: string }) {
    return (
        <span className="block min-w-0">
            <span className="block font-medium text-[--foreground] break-words">{title}</span>
            <span className="mt-0.5 block text-sm font-normal text-[--muted] break-words">{description}</span>
        </span>
    )
}

function formatRefreshSummary(job: Manga_MangaSourceRefreshJob) {
    const parts = [
        `${job.result.refreshed} refreshed`,
        `${job.result.found} found`,
        `${job.result.replaced} changed`,
    ]
    if (job.result.notFound > 0) parts.push(`${job.result.notFound} not found`)
    if (job.result.failed > 0) parts.push(`${job.result.failed} failed`)
    return `${parts.join(", ")}.`
}
