import { useClearFileCacheMediastreamVideoFiles, useGetFileCacheTotalSize, useRemoveFileCacheBucket } from "@/api/hooks/filecache.hooks"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import React from "react"

type FilecacheSettingsProps = {
    children?: React.ReactNode
}

export function FilecacheSettings(props: FilecacheSettingsProps) {

    const {
        children,
        ...rest
    } = props


    const { data: totalSize, mutate: getTotalSize, isPending: isFetchingSize } = useGetFileCacheTotalSize()

    const { mutate: clearBucket, isPending: _isClearing } = useRemoveFileCacheBucket(() => {
        getTotalSize()
    })

    const { mutate: clearMediastreamCache, isPending: _isClearing2 } = useClearFileCacheMediastreamVideoFiles(() => {
        getTotalSize()
    })

    const isClearing = _isClearing || _isClearing2

    return (
        <div className="space-y-4">
            <div className="flex gap-2 items-center">
                <Button intent="white-subtle" size="sm" onClick={() => getTotalSize()} disabled={isFetchingSize}>
                    Show total size
                </Button>
                {!!totalSize && (
                    <p>
                        {totalSize}
                    </p>
                )}
            </div>

            <Separator />

            <div className="flex gap-2 flex-wrap items-center">
                <Button intent="warning-subtle" onClick={() => clearBucket({ bucket: "manga" })} disabled={isClearing}>
                    Clear manga cache
                </Button>
                <Button intent="warning-subtle" onClick={() => clearMediastreamCache()} disabled={isClearing}>
                    Clear media streaming cache
                </Button>
                <Button intent="warning-subtle" onClick={() => clearBucket({ bucket: "onlinestream" })} disabled={isClearing}>
                    Clear online streaming cache
                </Button>
            </div>

            <Separator />

            <Button intent="alert-subtle" onClick={() => clearBucket({ bucket: "tvdb" })} disabled={isClearing}>
                Clear TVDB metadata
            </Button>

        </div>
    )
}
