export const VOLUME_BOOST_MIN = 1
export const VOLUME_BOOST_MAX = 3
export const VOLUME_BOOST_STEP = 0.25

export const VOLUME_BOOST_OPTIONS = [
    { label: "Off", value: 1 },
    { label: "125%", value: 1.25 },
    { label: "150%", value: 1.5 },
    { label: "200%", value: 2 },
    { label: "250%", value: 2.5 },
    { label: "300%", value: 3 },
]

export function clampVolumeBoost(value: number) {
    if (!Number.isFinite(value)) return VOLUME_BOOST_MIN
    return Math.min(VOLUME_BOOST_MAX, Math.max(VOLUME_BOOST_MIN, Math.round(value * 100) / 100))
}

export function formatVolumeBoost(value: number) {
    return value <= VOLUME_BOOST_MIN ? "Off" : `${Math.round(value * 100)}%`
}
