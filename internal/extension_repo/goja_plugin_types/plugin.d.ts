declare namespace $ui {
    /**
     * Registers the plugin as UI plugin.
     * @param fn - The setup function for the plugin.
     */
    function register(fn: (ctx: Context) => void): void

    interface Context {
        /**
         * Registers an event handler for the plugin.
         * @param eventName - The unique event identifier to register the handler for.
         * @param handler - The handler to register.
         * @returns A function to unregister the handler.
         */
        registerEventHandler(eventName: string, handler: () => void): () => void

        /**
         * Creates a new tray icon.
         * @param options - The options for the tray icon.
         * @returns A tray icon object.
         */
        newTray(options: {
            iconUrl: string
            withContent: boolean
            tooltipText?: string
        }): Tray
    }

    interface Tray {
        /**
         * Invokes the callback when the tray icon is clicked.
         * @param cb - The callback to invoke.
         */
        onClick(cb: () => void): void
    }

}
