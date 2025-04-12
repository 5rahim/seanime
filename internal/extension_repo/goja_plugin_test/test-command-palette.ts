/// <reference path="../goja_plugin_types/plugin.d.ts" />
/// <reference path="../goja_plugin_types/app.d.ts" />

//@ts-ignore
function init() {

    $ui.register((ctx) => {

        const tray = ctx.newTray({
            tooltipText: "Test Plugin",
            iconUrl: "https://raw.githubusercontent.com/5rahim/hibike/main/icons/seadex.png",
            withContent: false,
        })

        const cmd = ctx.newCommandPalette({
            placeholder: "Search for something",
            keyboardShortcut: "t",
        })

        tray.onClick(() => {
            cmd.open()
        })

        const renderTodos = async () => {
            const res = await ctx.fetch("https://jsonplaceholder.typicode.com/todos")
            const todos = await res.json()

            cmd.setItems(todos.map((todo: any) => ({
                label: todo.title,
                value: todo.title,
                filterType: "includes",
                onSelect: () => {
                    ctx.toast.info(`Todo ${todo.title} selected`)
                },
            })))
        }

        cmd.setItems([
            {
                label: "Fetch Todos",
                value: "fetch-todos",
                onSelect: async () => {
                    await renderTodos()
                },
            },
        ])

    })

}
