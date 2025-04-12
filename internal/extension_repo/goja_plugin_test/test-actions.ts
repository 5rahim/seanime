/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/app.d.ts" />
/// <reference path="../goja_plugin_types/system.d.ts" />

//@ts-ignore
function init() {

    $ui.register((ctx) => {

        const tray = ctx.newTray({
            tooltipText: "Test Plugin",
            iconUrl: "https://raw.githubusercontent.com/5rahim/hibike/main/icons/seadex.png",
            withContent: false,
        });

        const cmd = ctx.newCommandPalette({
            placeholder: "Search for something",
            keyboardShortcut: "t",
        })

        function renderOther() {
            cmd.setItems([
                {
                    value: "Test Item 2",
                    render: () => {
                        return cmd.stack({
                            items: [
                                cmd.text("Test Item 2"),
                                cmd.text("Test Item 2 Description", {style: {color: "gray"}}),
                            ]
                        })
                    },
                    onSelect: () => {
                        ctx.toast.info("Test Item 2 Clicked, Input: " + cmd.getInput());
                        renderFirst();
                    },
                }
            ])
        }

        function renderFirst() {
            cmd.setItems([
                {
                    value: "Test Item 1",
                    render: () => {
                        return cmd.stack({
                            items: [
                                cmd.text("Test Item 1"),
                                cmd.text("Test Item 1 Description", {style: {color: "gray"}}),
                            ]
                        })
                    },
                    filterType: "includes",
                    onSelect: () => {
                        ctx.toast.info("Test Item 1 Clicked, Input: " + cmd.getInput());
                        cmd.setInput("Test Item 1");
                        ctx.setTimeout(() => {
                            cmd.close();
                            renderOther();
                        }, 1000);
                    }
                },
                {
                    value: "Test Item 2",
                    label: "Label 2",
                    filterType: "includes",
                    onSelect: () => {
                        ctx.toast.info("Test Item 2 Clicked, Input: " + cmd.getInput());
                    }
                }
            ])
        }

        renderFirst();

        tray.onClick(() => {
            cmd.open();
        })

        const button = ctx.action.newAnimePageButton({
            label: "Test Button",
        });
        button.mount();

        button.onClick(() => {
            ctx.toast.info("Test Button Clicked");
            button.setLabel("Loading...");
            ctx.setTimeout(() => {
                button.setLabel("Loaded!");
            }, 1000);
        });

        const animeDropdownItem = ctx.action.newAnimePageDropdownItem({
            label: "Test Dropdown Item",
        });
        animeDropdownItem.mount();
        animeDropdownItem.onClick(() => {
            ctx.toast.info("Test Dropdown Item Clicked");
        });

        const animeLibraryDropdownItem = ctx.action.newAnimeLibraryDropdownItem({
            label: "Test Library Dropdown Item",
        });
        animeLibraryDropdownItem.mount();
        animeLibraryDropdownItem.onClick(() => {
            ctx.toast.info("Test Library Dropdown Item Clicked");
        });

        const mangaPageButton = ctx.action.newMangaPageButton({
            label: "Test Manga Page Button",
        });
        mangaPageButton.mount();
        mangaPageButton.onClick(() => {
            ctx.toast.info("Test Manga Page Button Clicked");
        });

        const mediaCardContextMenuItem = ctx.action.newMediaCardContextMenuItem({
            label: "Test Media Card Context Menu Item",
        });
        mediaCardContextMenuItem.mount();
        mediaCardContextMenuItem.onClick((e) => {
            ctx.toast.info("Test Media Card Context Menu Item Clicked");
            console.log("media card context menu item clicked", e);
        });


    });

}
