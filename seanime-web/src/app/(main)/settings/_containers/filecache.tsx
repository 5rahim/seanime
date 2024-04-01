import { Button } from "@/components/ui/button"
import { SeaEndpoints } from "@/lib/server/endpoints"
import { useSeaMutation } from "@/lib/server/query"
import { useQueryClient } from "@tanstack/react-query"
import React from "react"
import { toast } from "sonner"

type FilecacheSettingsProps = {
    children?: React.ReactNode
}

export function FilecacheSettings(props: FilecacheSettingsProps) {

    const {
        children,
        ...rest
    } = props

    const qc = useQueryClient()

    const [totalSize, setTotalSize] = React.useState<string>("")

    const { mutate: getTotalSize, isPending: isFetchingSize } = useSeaMutation<string>({
        endpoint: SeaEndpoints.FILECACHE_TOTAL_SIZE,
        mutationKey: ["get-filecache-total-size"],
        method: "get",
        onSuccess: data => {
            if (data) {
                setTotalSize(data)
            }
        },
    })

    const { mutate: clearBucket, isPending: isClearing } = useSeaMutation<void, { bucket: string }>({
        endpoint: SeaEndpoints.FILECACHE_BUCKET,
        mutationKey: ["clear-filecache-bucket"],
        method: "delete",
        onSuccess: async () => {
            toast.success("Cache cleared")
            getTotalSize()
        },
    })

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
            <div className="flex gap-2 flex-wrap items-center">
                <Button intent="alert-subtle" size="sm" onClick={() => clearBucket({ bucket: "manga" })} disabled={isClearing}>
                    Clear manga cache
                </Button>
                <Button intent="alert-subtle" size="sm" onClick={() => clearBucket({ bucket: "onlinestream" })} disabled={isClearing}>
                    Clear streaming cache
                </Button>
                <Button intent="alert-subtle" size="sm" onClick={() => clearBucket({ bucket: "tvdb" })} disabled={isClearing}>
                    Clear TVDB metadata
                </Button>
            </div>
        </div>
    )
}
