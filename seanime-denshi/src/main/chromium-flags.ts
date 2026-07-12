import { app } from "electron"
import log from "electron-log/main"

export function setupChromiumFlags() {
    app.commandLine.appendSwitch("no-zygote")

    app.commandLine.appendSwitch("autoplay-policy", "no-user-gesture-required")

    const mpvPrismHighPerformanceGpu = process.env.MPV_PRISM_HIGH_PERFORMANCE_GPU ||= "1"
    if (mpvPrismHighPerformanceGpu == "1" || mpvPrismHighPerformanceGpu == "true" || mpvPrismHighPerformanceGpu == "yes" || mpvPrismHighPerformanceGpu == "on") {
        app.commandLine.appendSwitch("force_high_performance_gpu")
    }

    if (mpvPrismHighPerformanceGpu == "0" || mpvPrismHighPerformanceGpu == "false" || mpvPrismHighPerformanceGpu == "no" || mpvPrismHighPerformanceGpu == "off") {
        app.commandLine.appendSwitch("force_low_power_gpu")
    }

    app.commandLine.appendSwitch("disk-cache-size", (400 * 1000 * 1000).toString())
    app.commandLine.appendSwitch("force-effective-connection-type", "4g")

    app.commandLine.appendSwitch("disable-features", [
        "WidgetLayering",
        "ColorProviderRedirection",
        "WebContentsForceDarkMode",
        "HardwareMediaKeyHandling",
        "CalculateNativeWinOcclusion",
    ].join(","))

    app.commandLine.appendSwitch("enable-zero-copy")
    app.commandLine.appendSwitch("enable-hardware-overlays", "single-fullscreen,single-on-top,underlay")
    app.commandLine.appendSwitch("ignore-gpu-blocklist")
    app.commandLine.appendSwitch("enable-accelerated-video-decode")

    app.commandLine.appendSwitch("enable-features", [
        "WebAssemblyLazyCompilation",
        "ThrottleDisplayNoneAndVisibilityHiddenCrossOriginIframes",
        "CanvasOopRasterization",
        "UseSkiaRenderer",
        "PlatformEncryptedDolbyVision",
        "SharedArrayBuffer",
    ].join(","))

    app.commandLine.appendSwitch("enable-unsafe-webgpu")
    app.commandLine.appendSwitch("enable-gpu-rasterization")
    app.commandLine.appendSwitch("enable-oop-rasterization")

    app.commandLine.appendSwitch("disable-background-timer-throttling")
    app.commandLine.appendSwitch("disable-backgrounding-occluded-windows")
    app.commandLine.appendSwitch("disable-renderer-backgrounding")
    app.commandLine.appendSwitch("disable-background-media-suspend")

    app.commandLine.appendSwitch("double-buffer-compositing")
    app.commandLine.appendSwitch("disable-direct-composition-video-overlays")

    if (process.platform === "linux") {
        log.info("Passing --gtk-version=3 to Electron")
        app.commandLine.appendSwitch("gtk-version", "3")
    }
}
