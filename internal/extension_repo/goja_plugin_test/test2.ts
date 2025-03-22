/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/hooks.d.ts" />
/// <reference path="../goja_plugin_types/system.d.ts" />

function init() {
    $ui.register((ctx) => {
        // Create the tray icon
        const tray = ctx.newTray({
            tooltipText: "Blur spoilers",
            iconUrl: "https://seanime.rahim.app/logo_2.png",
            withContent: false,
        })

        ctx.dom.observe("[data-episode-card-image]", (elements) => {
            elements.forEach((element) => {
                console.log(element.getStyle())
                element.setStyle("filter", "blur(24px)")
            })
        })
    })

}
