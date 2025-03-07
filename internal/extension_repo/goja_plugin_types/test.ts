/// <reference path="plugin.d.ts" />

function init() {
    $ui.register((ctx) => {
        const tray = ctx.newTray({
            iconUrl: "https://example.com/icon.png",
            tooltipText: "Example Tray",
        })

        tray.onClick(() => {
            console.log("Tray clicked")
        })
    })
}