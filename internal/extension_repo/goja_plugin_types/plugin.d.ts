declare namespace $ui {
    function register(fn: (ctx: Context) => void): void

    type Context = {
        registerEventHandler(event: string, handler: () => void): () => void

        // Tray
        newTray(props: {
            iconUrl: string
            tooltipText?: string
        }): Tray
    }

    type Tray = {
        onClick(cb: () => void): void
    }

}
