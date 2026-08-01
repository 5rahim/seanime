export type ElectronGpuPreference = "high-performance" | "low-power" | null

export function toElectronGpuPreference(value: string | undefined): ElectronGpuPreference {
    const normalized = value?.trim().toLowerCase()
    if (normalized === "1" || normalized === "true" || normalized === "yes" || normalized === "on") {
        return "high-performance"
    }
    if (normalized === "0" || normalized === "false" || normalized === "no" || normalized === "off") {
        return "low-power"
    }
    return null
}
