import { Anime_AutoDownloaderProfile } from "@/api/generated/types"
import { useDeleteAutoDownloaderProfile, useGetAutoDownloaderProfiles, useUpdateAutoDownloaderProfile } from "@/api/hooks/auto_downloader.hooks"
import { AutoDownloaderProfileForm } from "@/app/(main)/auto-downloader/_containers/autodownloader-profile-form"
import { ConfirmationDialog, useConfirmationDialog } from "@/components/shared/confirmation-dialog"
import { Button, IconButton } from "@/components/ui/button"
import { Card } from "@/components/ui/card"
import { cn } from "@/components/ui/core/styling"
import { LoadingSpinner } from "@/components/ui/loading-spinner"
import { Modal } from "@/components/ui/modal"
import { Switch } from "@/components/ui/switch"
import React from "react"
import { BiPencil, BiPlus, BiTrash } from "react-icons/bi"
import { PiTargetBold } from "react-icons/pi"

export function AutoDownloaderProfiles() {
    const { data: profiles, isLoading } = useGetAutoDownloaderProfiles()

    const [selectedProfile, setSelectedProfile] = React.useState<Anime_AutoDownloaderProfile | null>(null)
    const [isCreateOpen, setIsCreateOpen] = React.useState(false)

    if (isLoading) return <LoadingSpinner />

    return (
        <div className="pt-4">
            <Card className="p-4 space-y-3">
                <p className="text-[--muted]">
                    Profiles allow you to define a set of filters that can be applied to your rules.
                </p>
                <div className="flex items-center">
                    <Button
                        intent="white-subtle"
                        className="rounded-full"
                        leftIcon={<BiPlus className="text-xl" />}
                        onClick={() => setIsCreateOpen(true)}
                    >
                        New Profile
                    </Button>
                </div>

                <div className="space-y-3">
                    {profiles?.toSorted((a, b) => a.name?.toLowerCase()?.localeCompare(b.name?.toLowerCase()))?.map(profile => (
                        <ProfileItem
                            key={profile.dbId}
                            profile={profile}
                            onEdit={() => setSelectedProfile(profile)}
                        />
                    ))}
                    {!profiles?.length && (
                        <div className="p-4 text-center text-[--muted]">
                            No profiles created yet.
                        </div>
                    )}
                </div>
            </Card>

            <Modal
                title="Create Profile"
                open={isCreateOpen}
                onOpenChange={setIsCreateOpen}
                contentClass="max-w-3xl"
            >
                <AutoDownloaderProfileForm
                    onSuccess={() => setIsCreateOpen(false)}
                />
            </Modal>

            <Modal
                title={`Edit Profile: ${selectedProfile?.name}`}
                open={!!selectedProfile}
                onOpenChange={(v) => !v && setSelectedProfile(null)}
                contentClass="max-w-3xl"
            >
                {selectedProfile && (
                    <AutoDownloaderProfileForm
                        profile={selectedProfile}
                        onSuccess={() => setSelectedProfile(null)}
                    />
                )}
            </Modal>
        </div>
    )
}

function ProfileItem({ profile, onEdit }: { profile: Anime_AutoDownloaderProfile, onEdit: () => void }) {
    const { mutate: deleteProfile, isPending: deleting } = useDeleteAutoDownloaderProfile(profile.dbId)
    const { mutate: updateProfile, isPending: updating } = useUpdateAutoDownloaderProfile()

    const confirmDialog = useConfirmationDialog({
        title: "Delete profile",
        description: `Are you sure you want to delete the profile "${profile.name}"? This action cannot be undone.`,
        actionText: "Delete",
        actionIntent: "alert",
        onConfirm: async () => {
            deleteProfile()
        },
    })

    return (
        <>
            <Card className="p-3 flex items-center justify-between gap-4">
                <div className="flex items-center gap-3 w-full">
                    <div className="size-10 rounded-full bg-[--subtle] flex items-center justify-center">
                        <PiTargetBold
                            className={cn(
                                "text-xl opacity-50",
                                profile.global && "text-blue-300",
                            )}
                        />
                    </div>
                    <div className="w-full">
                        <h5 className="font-semibold flex items-center gap-2 line-clamp-1">
                            {profile.name}
                            {profile.global && <span className="text-xs bg-blue-500/20 text-blue-300 px-1.5 py-0.5 rounded-md">Global</span>}
                        </h5>
                        <div className="text-sm text-[--muted] line-clamp-1 gap-2 space-x-3">
                            {!!profile.resolutions?.length && <span className="!pl-0">{profile.resolutions.join(", ")}</span>}
                            {!!profile.conditions?.length &&
                                <span>{profile.conditions?.length} condition{(profile.conditions?.length !== 1) ? "s" : ""}</span>}
                            {!!profile.minimumScore && <span>{`>=`} {profile.minimumScore}</span>}
                            {!!profile.delayMinutes && <span>{profile.delayMinutes} min. delay</span>}

                        </div>
                    </div>
                </div>
                <div className="flex items-center gap-2">
                    <div className="items-center gap-2 mr-4 hidden lg:flex">
                        <span className="text-sm text-[--muted]">Global</span>
                        <Switch
                            value={profile.global}
                            onValueChange={(v) => updateProfile({ ...profile, global: v })}
                            disabled={updating}
                        />
                    </div>

                    <Button
                        intent="gray-subtle"
                        size="sm"
                        leftIcon={<BiPencil />}
                        onClick={onEdit}
                    >
                        Edit
                    </Button>
                    <IconButton
                        intent="alert-basic"
                        size="sm"
                        icon={<BiTrash />}
                        onClick={() => confirmDialog.open()}
                        loading={deleting}
                    />
                </div>
            </Card>
            <ConfirmationDialog {...confirmDialog} />
        </>
    )
}
