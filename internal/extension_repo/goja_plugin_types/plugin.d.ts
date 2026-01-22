/// <reference path="app.d.ts" />
/// <reference path="core.d.ts" />

declare namespace $ui {
    /**
     * Registers the plugin as UI plugin.
     * @param fn - The setup function for the plugin.
     */
    function register(fn: (ctx: Context) => void): void

    interface Context {
        /**
         * Screen navigation and management
         */
        screen: Screen
        /**
         * Toast notifications
         */
        toast: Toast
        /**
         * Actions
         */
        action: Action

        /**
         * DOM
         */
        dom: DOM

        /**
         * Playback
         */
        playback: Playback

        /**
         * MPV
         */
        mpv: MPV

        /**
         * Notifications
         */
        notification: Notification

        /**
         * Manga
         */
        manga: Manga

        /**
         * Anime
         */
        anime: Anime

        /**
         * Discord
         */
        discord: Discord

        /**
         * Continuity
         */
        continuity: Continuity

        /**
         * Auto Scanner
         */
        autoScanner: AutoScanner

        /**
         * External Player Link
         */
        externalPlayerLink: ExternalPlayerLink

        /**
         * Auto Downloader
         */
        autoDownloader: AutoDownloader

        /**
         * Filler Manager
         */
        fillerManager: FillerManager

        /**
         * Torrent Client
         */
        torrentClient: TorrentClient

        /**
         * Creates a new state object with an initial value.
         * @param initialValue - The initial value for the state
         * @returns A state object that can be used to get and set values
         */
        state<T>(initialValue?: T): State<T>

        /**
         * Sets a timeout to execute a function after a delay.
         * @param fn - The function to execute
         * @param delay - The delay in milliseconds
         * @returns A function to cancel the timeout
         */
        setTimeout(fn: () => void, delay: number): () => void

        /**
         * Sets an interval to execute a function repeatedly.
         * @param fn - The function to execute
         * @param delay - The delay in milliseconds between executions
         * @returns A function to cancel the interval
         */
        setInterval(fn: () => void, delay: number): () => void

        /**
         * Creates an effect that runs when dependencies change.
         * @param fn - The effect function to run
         * @param deps - Array of dependencies that trigger the effect
         * @returns A function to clean up the effect
         */
        effect(fn: () => void, deps: State<any>[]): () => void

        /**
         * Makes a fetch request.
         * @param url - The URL to fetch
         * @param options - Fetch options
         * @returns A promise that resolves to the fetch response
         */
        fetch(url: string, options?: FetchOptions): Promise<FetchResponse>

        /**
         * Registers an event handler for the plugin.
         * @param eventName - The unique event identifier to register the handler for.
         * @param handler - The handler to register.
         * @returns A function to unregister the handler.
         */
        registerEventHandler(eventName: string, handler: (event: any) => void): () => void

        /**
         * Registers an event handler for the plugin.
         * @param uniqueKey - A unique key to identify the handler. This is to avoid memory leaks caused by re-rendering the same component.
         * @param handler - The handler to register.
         * @returns The event handler id.
         */
        eventHandler(uniqueKey: string, handler: (event: any) => void): string

        /**
         * Registers a field reference for field components.
         * @returns A field reference object
         * @param defaultValue
         */
        fieldRef<T extends any = string>(defaultValue?: T): FieldRef<T>

        /**
         * Creates a new tray icon.
         * @param options - The options for the tray icon.
         * @returns A tray icon object.
         */
        newTray(options: TrayOptions): Tray

        /**
         * Creates a new webview.
         * @param options - The options for the webview.
         * @returns A webview object.
         */
        newWebview(options: WebviewOptions): Webview

        /**
         * Creates a new command palette.
         * @param options - The options for the command palette
         * @returns A command palette object
         */
        newCommandPalette(options: CommandPaletteOptions): CommandPalette

        /**
         * Use a headless browser.
         */
        chromeDP: ChromeDP

        /**
         * Video Core for controlling the built-in player
         */
        videoCore: VideoCore
    }

    interface State<T> {
        /** The current value */
        value: T
        /** Length of the value if it's a string */
        length?: number

        /** Gets the current value */
        get(): T

        /** Sets a new value */
        set(value: T | ((prev: T) => T)): void
    }

    type ReadOnlyState<T> = Omit<State<T>, "set">

    interface FetchOptions {
        /** HTTP method, defaults to GET */
        method?: string
        /** Request headers */
        headers?: Record<string, string>
        /** Request body */
        body?: any
        /** Whether to bypass cloudflare */
        noCloudflareBypass?: boolean
        /** Timeout in seconds, defaults to 35 */
        timeout?: number
        /** AbortSignal to cancel the request */
        signal?: AbortSignal
    }

    /**
     * AbortContext provides a way to abort certain tasks
     */
    class AbortContext {
        /** The signal object associated with this context */
        readonly signal: AbortSignal

        /**
         * Aborts the associated task
         * @param reason - Optional reason for aborting
         */
        abort(reason?: string): void
    }

    /**
     * AbortSignal represents a signal that can be used to abort requests
     */
    interface AbortSignal {
        /** Whether the signal has been aborted */
        readonly aborted: boolean
        /** The reason for aborting, if any */
        readonly reason: string
    }

    interface FetchResponse {
        /** Response status code */
        status: number
        /** Response status text */
        statusText: string
        /** Request method used */
        method: string
        /** Raw response headers */
        rawHeaders: Record<string, string[]>
        /** Whether the response was successful (status in range 200-299) */
        ok: boolean
        /** Request URL */
        url: string
        /** Response headers */
        headers: Record<string, string>
        /** Response cookies */
        cookies: Record<string, string>
        /** Whether the response was redirected */
        redirected: boolean
        /** Response content type */
        contentType: string
        /** Response content length */
        contentLength: number

        /** Get response text */
        text(): string

        /** Parse response as JSON */
        json<T = any>(): T
    }

    interface FieldRef<T> {
        /** The current value of the field */
        current: T

        /** Sets the value of the field */
        setValue(value: T): void

        /** Sets the callback to be called when the value changes */
        onValueChange(callback: (value: T) => void): void
    }

    interface TrayOptions {
        /** URL of the tray icon */
        iconUrl: string
        /** Whether the tray has content */
        withContent: boolean
        /** Width of the tray */
        width?: string
        /** Minimum height of the tray */
        minHeight?: string
    }

    interface Tray {
        /** UI components for building tray content */
        div: DivComponentFunction
        flex: FlexComponentFunction
        stack: StackComponentFunction
        text: TextComponentFunction
        button: ButtonComponentFunction
        anchor: AnchorComponentFunction
        input: InputComponentFunction
        select: SelectComponentFunction
        checkbox: CheckboxComponentFunction
        radioGroup: RadioGroupComponentFunction
        switch: SwitchComponentFunction
        css: CSSComponentFunction
        tooltip: TooltipComponentFunction
        modal: ModalComponentFunction
        dropdownMenu: DropdownMenuComponentFunction
        dropdownMenuItem: DropdownMenuItemComponentFunction
        dropdownMenuSeparator: DropdownMenuSeparatorComponentFunction
        dropdownMenuLabel: DropdownMenuLabelComponentFunction
        popover: PopoverComponentFunction
        a: AComponentFunction
        p: PComponentFunction
        alert: AlertComponentFunction
        tabs: TabsComponentFunction
        tabsList: TabsListComponentFunction
        tabsTrigger: TabsTriggerComponentFunction
        tabsContent: TabsContentComponentFunction
        badge: BadgeComponentFunction
        span: SpanComponentFunction
        img: ImgComponentFunction

        /** Invoked when the tray icon is clicked */
        onClick(cb: () => void): void

        /** Invoked when the tray icon is opened */
        onOpen(cb: () => void): void

        /** Invoked when the tray icon is closed */
        onClose(cb: () => void): void

        /** Registers the render function for the tray content */
        render(fn: () => void): void

        /** Registers the render function for the tray content */
        htm(fn: () => string): void

        /** Schedules a re-render of the tray content */
        update(): void

        /** Opens the tray */
        open(): void

        /** Closes the tray */
        close(): void

        /** Updates the badge number of the tray icon. 0 = no badge. Default intent is "info". */
        updateBadge(options: { number: number, intent?: "success" | "error" | "warning" | "info" }): void
    }

    interface WebviewOptions {
        slot: "screen" |
            "fixed" |
            "after-home-screen-toolbar" |
            "home-screen-bottom" |
            "schedule-screen-top" |
            "schedule-screen-bottom" |
            "anime-screen-bottom" |
            "after-anime-entry-episode-list" |
            "after-anime-episode-list" |
            "before-anime-entry-episode-list" |
            "manga-screen-bottom" |
            "manga-entry-screen-bottom" |
            "after-manga-entry-chapter-list" |
            "after-discover-screen-header" |
            "after-media-entry-details" |
            "after-media-entry-form"

        // Styling options
        className?: string
        style?: string
        width?: string
        height?: string
        maxWidth?: string
        maxHeight?: string
        zIndex?: number

        /**
         * Only applies if slot is "fixed"
         */
        window?: {
            draggable?: boolean
            defaultX?: number
            defaultY?: number
            defaultPosition?: "top-left" | "top-right" | "bottom-left" | "bottom-right"
            frameless?: boolean
        }
        /**
         * Whether the height of the webview should be automatically adjusted to fit its content.
         */
        autoHeight?: boolean
        /**
         * Whether the width of the webview should be automatically adjusted to fit its container.
         */
        fullWidth?: boolean

        sidebar?: {
            label: string,
            icon: string,
        }

        hidden?: boolean
    }

    interface WebviewChannel {
        /**
         * Automatically syncs a state with the webview.
         * @example
         * // Plugin context:
         * const count = ctx.state(0)
         * myWebview.channel.sync("count", count)
         * //...
         * count.set(count.get() + 1)
         *
         * // Webview code:
         * webview.channel.on("count", (count) => {
         *     console.log("Received from plugin context: " + count)
         * })
         * @param eventName
         * @param state
         */
        sync(eventName: string, state: State<any> | ReadOnlyState<any>): void

        /**
         * Registers an event listener for messages from the webview.
         * @example
         * myWebview.channel.on("eventName", (payload) => {
         *     // Handle message here
         * })
         * @param eventName
         * @param cb
         */
        on(eventName: string, cb: (payload: any) => void): void

        /**
         * Sends a message to the webview.
         * @example
         * // Plugin context:
         * myWebview.channel.send("eventName", payload)
         *
         * // Webview code:
         * webview.channel.on("eventName", (payload) => {
         *     // Handle message here
         * })
         * @param eventName The name of the event to send.
         * @param payload The payload to send.
         */
        send(eventName: string, payload: any): void
    }

    interface Webview {
        /** UI components for building webviews */
        div: DivComponentFunction
        flex: FlexComponentFunction
        stack: StackComponentFunction
        text: TextComponentFunction
        button: ButtonComponentFunction
        anchor: AnchorComponentFunction
        input: InputComponentFunction
        select: SelectComponentFunction
        checkbox: CheckboxComponentFunction
        radioGroup: RadioGroupComponentFunction
        switch: SwitchComponentFunction
        css: CSSComponentFunction
        tooltip: TooltipComponentFunction
        modal: ModalComponentFunction
        dropdownMenu: DropdownMenuComponentFunction
        dropdownMenuItem: DropdownMenuItemComponentFunction
        dropdownMenuSeparator: DropdownMenuSeparatorComponentFunction
        dropdownMenuLabel: DropdownMenuLabelComponentFunction
        popover: PopoverComponentFunction
        a: AComponentFunction
        p: PComponentFunction
        alert: AlertComponentFunction
        tabs: TabsComponentFunction
        tabsList: TabsListComponentFunction
        tabsTrigger: TabsTriggerComponentFunction
        tabsContent: TabsContentComponentFunction
        badge: BadgeComponentFunction
        span: SpanComponentFunction
        img: ImgComponentFunction
        /**
         * Communication channel between the webview and the Plugin context.
         */
        channel: WebviewChannel

        /** Invoked when the webview is mounted, before it's loaded */
        onMount(cb: () => void): void

        /** Invoked when the webview is loaded, after it's mounted */
        onLoad(cb: () => void): void

        /** Invoked after the webview is unmounted */
        onUnmount(cb: () => void): void

        /**
         * Updates the webview's content.
         *
         * This is useful if the webview's content depends on the state of the Plugin context.
         */
        update(): void

        /** Set webview's iframe content */
        setContent(fn: () => string): void

        /** Update webview options dynamically */
        setOptions(options: Partial<Omit<WebviewOptions, "sidebar">>): void

        /** Removes the webview from the DOM (not reversible) */
        close(): void

        /** Show the webview (reverses hide) */
        show(): void

        /** Hide the webview */
        hide(): void

        isHidden(): boolean

        /**
         * Returns the path of the webview's screen
         * @example /webview?id=my-plugin
         */
        getScreenPath(): string
    }

    interface Playback {
        /**
         * Plays a file using the media player
         * @param filePath - The path to the file to play
         * @returns A promise that resolves when the file has started playing
         */
        playUsingMediaPlayer(filePath: string): Promise<void>

        /**
         * Streams a file using the media player
         * @param windowTitle - The title of the window
         * @param streamUrl - The URL of the stream to play
         * @param anime - The anime object
         * @param aniDbEpisode - The AniDB episode number
         * @throws Error if an error occurs
         */
        streamUsingMediaPlayer(windowTitle: string, streamUrl: string, anime: $app.AL_BaseAnime, aniDbEpisode: string): Promise<void>

        /**
         * Registers an event listener for the playback instance
         * @param callback - The callback to call when the event occurs
         * @returns A function to remove the event listener
         */
        registerEventListener(callback: (event: PlaybackEvent) => void): () => void

        /**
         * Cancels the tracking of the current media being played.
         * @throws Error if an error occurs, or if the playback is not running
         */
        cancel(): void

        /**
         * Pauses the playback
         * @throws Error if an error occurs, or if the playback is not running
         */
        pause(): void

        /**
         * Resumes the playback
         * @throws Error if an error occurs, or if the playback is not paused
         */
        resume(): void

        /**
         * Seeks to a specific position in the playback
         * @param seconds - The position to seek to
         * @throws Error if an error occurs, or if the playback is not running
         */
        seek(seconds: number): void

        /**
         * Gets the next episode to play for the current media being played
         * @returns The next episode to play
         */
        getNextEpisode(): Promise<$app.Anime_LocalFile | undefined>

        /**
         * Plays the next episode for the current media being played
         * @throws Error if an error occurs, or if the playback is not running
         */
        playNextEpisode(): Promise<void>

    }

    interface PlaybackEvent {
        /** Whether the video has started */
        isVideoStarted: boolean
        /** Whether the video has stopped */
        isVideoStopped: boolean
        /** Whether the video has completed */
        isVideoCompleted: boolean
        /** Whether the stream has started */
        isStreamStarted: boolean
        /** Whether the stream has stopped */
        isStreamStopped: boolean
        /** Whether the stream has completed */
        isStreamCompleted: boolean
        /** The event that occurred when the video started */
        startedEvent: { filename: string }
        /** The event that occurred when the video stopped */
        stoppedEvent: { reason: string }
        /** The event that occurred when the video completed */
        completedEvent: { filename: string }
        /** The state of the playback */
        state: PlaybackState
        /** The status of the playback */
        status: PlaybackStatus
    }

    interface PlaybackState {
        /** The episode number */
        episodeNumber: number
        /** The title of the media */
        mediaTitle: string
        /** The cover image of the media */
        mediaCoverImage: string
        /** The total number of episodes */
        mediaTotalEpisodes: number
        /** The filename */
        filename: string
        /** The completion percentage */
        completionPercentage: number
        /** Whether the next episode can be played */
        canPlayNext: boolean
        /** Whether the progress has been updated */
        progressUpdated: boolean
        /** The media ID */
        mediaId: number
    }

    interface PlaybackStatus {
        /** The completion percentage of the playback */
        completionPercentage: number
        /** Whether the playback is playing */
        playing: boolean
        /** The filename */
        filename: string
        /** The path */
        path: string
        /** The duration of the playback, in milliseconds */
        duration: number
        /** The filepath */
        filepath: string
        /** The current time in seconds */
        currentTimeInSeconds: number
        /** The duration in seconds */
        durationInSeconds: number
    }

    interface MPV {
        /**
         * Opens and plays a file
         * @throws Error if an error occurs
         * @returns A promise that resolves when the file has started playing
         */
        openAndPlay(filePath: string): Promise<void>

        /**
         * Stops the playback
         * @returns A promise that resolves when the playback has stopped
         */
        stop(): Promise<void>

        /**
         * Returns the connection object
         * @returns The connection object or undefined if the connection is not open
         */
        getConnection(): MpvConnection | undefined

        /**
         * Registers an event listener for the MPV instance
         *
         * Some properties are already observed by default with the following IDs:
         * - 42 = time-pos
         * - 43 = pause
         * - 44 = duration
         * - 45 = filename
         * - 46 = path
         *
         * You can observe other properties by getting the connection object and calling conn.call("observe_property", id, property)
         *
         * @param event - The event to listen for
         * @param callback - The callback to call when the event occurs
         */
        onEvent(callback: (event: MpvEvent) => void): void
    }

    interface MpvConnection {
        /**
         * Calls a command on the MPV instance
         * @param command - The command to call
         * @param args - The arguments to pass to the command
         */
        call(...args: any[]): void

        /**
         * Sets a property on the MPV instance
         * @param property - The property to set
         * @param value - The value to set the property to
         */
        set(property: string, value: any): void

        /**
         * Gets a property from the MPV instance
         * @param property - The property to get
         * @returns The value of the property
         */
        get(property: string): any

        /**
         * Closes the connection to the MPV instance
         */
        close(): void

        /**
         * Whether the connection is closed
         */
        isClosed(): boolean
    }

    interface MpvEvent {
        /** The name of the event */
        event: string
        /** The data associated with the event */
        data: any
        /** The reason for the event */
        reason: string
        /** The prefix for the event */
        prefix: string
        /** The level of the event */
        level: string
        /** The text of the event */
        text: string
        /** The ID of the event */
        id: number
    }

    type AnimePageButtonAction = ActionObject<{ media: $app.AL_BaseAnime }> & { setTooltipText(text: string): void }

    type MangaPageButtonAction = ActionObject<{ media: $app.AL_BaseManga }> & { setTooltipText(text: string): void }

    interface Action {
        /**
         * Creates a new button for the anime page
         * @param props - Button properties
         */
        newAnimePageButton(props: { label: string, intent?: Intent, style?: Record<string, string>, tooltipText?: string }): AnimePageButtonAction

        /**
         * Creates a new dropdown menu item for the anime page
         * @param props - Dropdown item properties
         */
        newAnimePageDropdownItem(props: { label: string, style?: Record<string, string> }): ActionObject<{ media: $app.AL_BaseAnime }>

        /**
         * Creates a new dropdown menu item for the anime library
         * @param props - Dropdown item properties
         */
        newAnimeLibraryDropdownItem(props: { label: string, style?: Record<string, string> }): ActionObject

        /**
         * Creates a new context menu item for media cards
         * @param props - Context menu item properties
         */
        newMediaCardContextMenuItem<F extends "anime" | "manga" | "both">(props: {
            label: string,
            for?: F,
            style?: Record<string, string>
        }): ActionObject<{
            media: F extends "anime" ? $app.AL_BaseAnime : F extends "manga" ? $app.AL_BaseManga : $app.AL_BaseAnime | $app.AL_BaseManga
        }> & {
            /** Sets the 'for' property of the action */
            setFor(forMedia: "anime" | "manga" | "both"): void
        }

        /**
         * Creates a new button for the manga page
         * @param props - Button properties
         */
        newMangaPageButton(props: { label: string, intent?: Intent, style?: Record<string, string>, tooltipText?: string }): MangaPageButtonAction

        /**
         * Creates a new context menu item for the episode card
         * @param props - Context menu item properties
         */
        newEpisodeCardContextMenuItem(props: { label: string, style?: Record<string, string> }): ActionObject<{ episode: $app.Anime_Episode }>

        /**
         * Creates a new menu item for the episode grid item
         * @param props - Menu item properties
         */
        newEpisodeGridItemMenuItem(props: {
            label: string,
            style?: Record<string, string>,
            type: "library" | "torrentstream" | "debridstream" | "onlinestream" | "undownloaded" | "medialinks" | "mediastream"
        }): ActionObject<{
            episode: $app.Anime_Episode | $app.Onlinestream_Episode,
            type: "library" | "torrentstream" | "debridstream" | "onlinestream" | "undownloaded" | "medialinks" | "mediastream"
        }>
    }

    interface ActionObject<E extends any = {}> {
        /** Mounts the action to make it visible */
        mount(): void

        /** Unmounts the action to hide it */
        unmount(): void

        /** Sets the label of the action */
        setLabel(label: string): void

        /** Sets the loading state of the action */
        setLoading(loading: boolean): void

        /** Sets the disabled state of the action */
        setDisabled(disabled: boolean): void

        /** Sets the style of the action */
        setStyle(style: Record<string, string>): void

        /** Sets the click handler for the action */
        onClick(handler: (event: E) => void): void

        /** Sets the intent of the action */
        setIntent(intent: Intent): void
    }

    interface CommandPaletteOptions {
        /** Placeholder text for the command palette input */
        placeholder?: string
        /** Keyboard shortcut to open the command palette */
        keyboardShortcut?: string
    }

    interface CommandPalette {
        /** UI components for building command palette items */
        div: DivComponentFunction
        flex: FlexComponentFunction
        stack: StackComponentFunction
        text: TextComponentFunction
        button: ButtonComponentFunction
        anchor: AnchorComponentFunction
        input: InputComponentFunction
        select: SelectComponentFunction
        checkbox: CheckboxComponentFunction
        radioGroup: RadioGroupComponentFunction
        switch: SwitchComponentFunction
        css: CSSComponentFunction
        tooltip: TooltipComponentFunction
        modal: ModalComponentFunction
        dropdownMenu: DropdownMenuComponentFunction
        dropdownMenuItem: DropdownMenuItemComponentFunction
        dropdownMenuSeparator: DropdownMenuSeparatorComponentFunction
        dropdownMenuLabel: DropdownMenuLabelComponentFunction
        popover: PopoverComponentFunction
        a: AComponentFunction
        p: PComponentFunction
        alert: AlertComponentFunction
        tabs: TabsComponentFunction
        tabsList: TabsListComponentFunction
        tabsTrigger: TabsTriggerComponentFunction
        tabsContent: TabsContentComponentFunction
        badge: BadgeComponentFunction
        span: SpanComponentFunction
        img: ImgComponentFunction

        /** Sets the items in the command palette */
        setItems(items: CommandPaletteItem[]): void

        /** Refreshes the command palette items */
        refresh(): void

        /** Sets the placeholder text */
        setPlaceholder(placeholder: string): void

        /** Opens the command palette */
        open(): void

        /** Closes the command palette */
        close(): void

        /** Sets the input value */
        setInput(input: string): void

        /** Gets the current input value */
        getInput(): string

        /** Called when the command palette is opened */
        onOpen(cb: () => void): void

        /** Called when the command palette is closed */
        onClose(cb: () => void): void
    }

    interface CommandPaletteItem {
        /** Label for the item */
        label?: string
        /** Value associated with the item */
        value: string
        /**
         * Type of filtering to apply when the input changes.
         * If not provided, the item will not be filtered.
         */
        filterType?: "includes" | "startsWith"
        /** Heading for the item group */
        heading?: string
        /** Custom render function for the item */
        render?: () => void
        /** Called when the item is selected */
        onSelect: () => void
    }

    interface Screen {
        /** Navigates to a specific path */
        navigateTo(path: string, searchParams?: Record<string, string>): void

        /** Reloads the current screen */
        reload(): void

        /** Calls onNavigate with the current screen data */
        loadCurrent(): void

        /** Called when navigation occurs */
        onNavigate(cb: (event: { pathname: string, searchParams: Record<string, string> }) => void): void

        /** Returns a state object containing the current screen data */
        state(): ReadOnlyState<{ pathname: string, searchParams: Record<string, string> }>
    }

    interface Toast {
        /** Shows a success toast */
        success(message: string): void

        /** Shows an error toast */
        error(message: string): void

        /** Shows an info toast */
        info(message: string): void

        /** Shows a warning toast */
        warning(message: string): void
    }

    type ComponentFunction = (props: any) => void
    type ComponentProps = {
        style?: Record<string, string>,
        className?: string,
    }
    type FieldComponentProps<V = string> = {
        fieldRef?: FieldRef<V>,
        value?: V,
        onChange?: string,
        disabled?: boolean,
        size?: "sm" | "md" | "lg",

    } & ComponentProps

    type DivComponentFunction = {
        (props: { items: any[] } & ComponentProps): void
        (items: any[], props?: ComponentProps): void
    }
    type CSSComponentFunction = {
        (props: { css: string }): void
        (css: string): void
    }
    type FlexComponentFunction = {
        (props: { items: any[], gap?: number, direction?: "row" | "column" } & ComponentProps): void
        (items: any[], props?: { gap?: number, direction?: "row" | "column" } & ComponentProps): void
    }
    type StackComponentFunction = {
        (props: { items: any[], gap?: number } & ComponentProps): void
        (items: any[], props?: { gap?: number } & ComponentProps): void
    }
    type TextComponentFunction = {
        (props: { text: string } & ComponentProps): void
        (text: string, props?: ComponentProps): void
    }
    type TooltipComponentFunction = {
        (props: { text: string, item: any, side?: "top" | "right" | "bottom" | "left", sideOffset?: number }): void
        (item: any, props: { text: string, side?: "top" | "right" | "bottom" | "left", sideOffset?: number }): void
    }

    /**
     * @default size="sm"
     */
    type ButtonComponentFunction = {
        (props: {
            label?: string,
            onClick?: string,
            intent?: Intent,
            disabled?: boolean,
            loading?: boolean,
            size?: "xs" | "sm" | "md" | "lg"
        } & ComponentProps): void
        (label: string,
            props?: { onClick?: string, intent?: Intent, disabled?: boolean, loading?: boolean, size?: "xs" | "sm" | "md" | "lg" } & ComponentProps,
        ): void
    }
    /**
     * @default target="_blank"
     */
    type AnchorComponentFunction = {
        (props: {
            text: string,
            href: string,
            target?: string,
            onClick?: string
        } & ComponentProps): void
        (text: string,
            props: { href: string, target?: string, onClick?: string } & ComponentProps,
        ): void
    }
    /**
     * @default size="md"
     */
    type InputComponentFunction = {
        (props: { label?: string, placeholder?: string, textarea?: boolean, onSelect?: string } & FieldComponentProps): void
        (label: string, props?: { placeholder?: string, textarea?: boolean, onSelect?: string } & FieldComponentProps): void
    }
    /**
     * @default size="md"
     */
    type SelectComponentFunction = {
        (props: { label: string, placeholder?: string, options: { label: string, value: string }[] } & FieldComponentProps): void
        (label: string, options: { placeholder?: string, value?: string, options: { label: string, value: string }[] } & FieldComponentProps): void
    }
    /**
     * @default size="md"
     */
    type CheckboxComponentFunction = {
        (props: { label: string } & FieldComponentProps<boolean>): void
        (label: string, props?: FieldComponentProps<boolean>): void
    }
    /**
     * @default size="md"
     */
    type RadioGroupComponentFunction = {
        (props: { label: string, options: { label: string, value: string }[] } & FieldComponentProps): void
        (label: string, options: { value?: string, options: { label: string, value: string }[] } & FieldComponentProps): void
    }
    /**
     * @default side="right"
     * @default size="sm"
     */
    type SwitchComponentFunction = {
        (props: { label: string, side?: "left" | "right" } & FieldComponentProps<boolean>): void
        (label: string, props?: { side?: "left" | "right" } & FieldComponentProps<boolean>): void
    }

    type ModalComponentFunction = {
        (props: {
            trigger: any,
            title?: string,
            description?: string,
            items?: any[],
            footer?: any[],
            open?: boolean,
            onOpenChange?: string
        } & ComponentProps): void
    }

    type DropdownMenuComponentFunction = {
        (props: {
            trigger: any,
            items: any[]
        } & ComponentProps): void
    }

    type DropdownMenuItemComponentFunction = {
        (props: {
            item: any,
            onClick?: string,
            disabled?: boolean
        } & ComponentProps): void
        (item: any, props?: { onClick?: string, disabled?: boolean } & ComponentProps): void
    }

    type DropdownMenuSeparatorComponentFunction = {
        (props?: ComponentProps): void
    }

    type DropdownMenuLabelComponentFunction = {
        (props: { label: string } & ComponentProps): void
        (label: string, props?: ComponentProps): void
    }

    type PopoverComponentFunction = {
        (props: {
            trigger: any,
            items: any[]
        } & ComponentProps): void
    }

    /**
     * @default target="_blank"
     */
    type AComponentFunction = {
        (props: {
            href: string,
            items: any[],
            target?: string,
            onClick?: string
        } & ComponentProps): void
        (items: any[], props: { href: string, target?: string, onClick?: string } & ComponentProps): void
    }

    type PComponentFunction = {
        (props: { items: any[] } & ComponentProps): void
        (items: any[], props?: ComponentProps): void
    }

    /**
     * @default intent="info"
     */
    type AlertComponentFunction = {
        (props: {
            title?: string,
            description?: string,
            intent?: "info" | "success" | "warning" | "alert"
        } & ComponentProps): void
    }

    type TabsComponentFunction = {
        (props: {
            defaultValue?: string,
            items: any[]
        } & ComponentProps): void
        (items: any[], props?: { defaultValue?: string } & ComponentProps): void
    }

    type TabsListComponentFunction = {
        (props: { items: any[] } & ComponentProps): void
        (items: any[], props?: ComponentProps): void
    }

    type TabsTriggerComponentFunction = {
        (item: any, props: { value: string }): void
        (props: {
            item: any,
            value: string
        }): void
    }

    type TabsContentComponentFunction = {
        (props: {
            value: string,
            items: any[]
        } & ComponentProps): void
        (items: any[], props: { value: string } & ComponentProps): void
    }

    /**
     * @default intent="gray"
     * @default size="md"
     */
    type BadgeComponentFunction = {
        (props: {
            text: string,
            intent?: "gray" | "primary" | "success" | "warning" | "alert" | "info" | "blue",
            size?: "sm" | "md" | "lg" | "xl"
        } & ComponentProps): void
        (text: string, props?: {
            intent?: "gray" | "primary" | "success" | "warning" | "alert" | "info" | "blue",
            size?: "sm" | "md" | "lg" | "xl"
        } & ComponentProps): void
    }

    type SpanComponentFunction = {
        (props: { text: string, items?: any[] } & ComponentProps): void
        (text: string, props?: { items?: any[] } & ComponentProps): void
    }

    type ImgComponentFunction = {
        (props: { src: string, alt?: string, width?: string, height?: string } & ComponentProps): void
        (src: string, props?: { alt?: string, width?: string, height?: string } & ComponentProps): void
    }

    // DOM Element interface
    interface DOMElement {
        id: string
        tagName: string
        attributes: Record<string, string>
        textContent?: string
        // Only available if withInnerHTML is true
        innerHTML?: string
        // Only available if withOuterHTML is true
        outerHTML?: string

        /**
         * Gets the text content of the element
         * @returns A promise that resolves to the text content of the element
         */
        getText(): Promise<string>

        /**
         * Sets the text content of the element
         * @param text - The text content to set
         */
        setText(text: string): void

        /**
         * Gets the value of an attribute
         * @param name - The name of the attribute
         * @returns A promise that resolves to the value of the attribute
         */
        getAttribute(name: string): Promise<string | null>

        /**
         * Gets all attributes of the element
         * @returns A promise that resolves to a record of all attributes
         */
        getAttributes(): Promise<Record<string, string>>

        /**
         * Sets the value of an attribute
         * @param name - The name of the attribute
         * @param value - The value to set
         */
        setAttribute(name: string, value: string): void

        /**
         * Removes an attribute
         * @param name - The name of the attribute
         */
        removeAttribute(name: string): void

        /**
         * Checks if the element has an attribute
         * @param name - The name of the attribute
         * @returns A promise that resolves to true if the attribute exists
         */
        hasAttribute(name: string): Promise<boolean>

        /**
         * Gets a property of the element
         * @param name - The name of the property
         * @returns A promise that resolves to the value of the property
         */
        getProperty(name: string): Promise<string | null>

        /**
         * Sets a property of the element
         * @param name - The name of the property
         * @param value - The value to set
         */
        setProperty(name: string, value: any): void

        /**
         * Adds a class to the element
         * @param classNames - The classes to add
         */
        addClass(...classNames: string[]): void

        /**
         * Removes a class from the element
         * @param classNames - The classes to add
         */
        removeClass(...classNames: string[]): void

        /**
         * Checks if the element has a class
         * @param className - The class to check
         * @returns A promise that resolves to true if the class exists
         */
        hasClass(className: string): Promise<boolean>

        /**
         * Sets the CSS text of the element
         * @param cssText - The CSS text to set
         */
        setCssText(cssText: string): void

        /**
         * Sets the style of the element
         * @param property - The property to set
         * @param value - The value to set
         */
        setStyle(property: string, value: string): void

        /**
         * Gets the style of the element
         * @param property - Optional property to get. If omitted, returns all styles.
         * @returns A promise that resolves to the value of the property or record of all styles
         */
        getStyle(property?: string): Promise<string | Record<string, string>>

        /**
         * Checks if the element has a style property set
         * @param property - The property to check
         * @returns A promise that resolves to true if the property is set
         */
        hasStyle(property: string): Promise<boolean>

        /**
         * Removes a style property
         * @param property - The property to remove
         */
        removeStyle(property: string): void

        /**
         * Gets the computed style of the element
         * @param property - The property to get
         * @returns A promise that resolves to the computed value of the property
         */
        getComputedStyle(property: string): Promise<string>

        /**
         * Gets a data attribute (data-* attribute)
         * @param key - The data attribute key (without the data- prefix)
         * @returns A promise that resolves to the data attribute value
         */
        getDataAttribute(key: string): Promise<string | null>

        /**
         * Gets all data attributes (data-* attributes)
         * @returns A promise that resolves to a record of all data attributes
         */
        getDataAttributes(): Promise<Record<string, string>>

        /**
         * Sets a data attribute (data-* attribute)
         * @param key - The data attribute key (without the data- prefix)
         * @param value - The value to set
         */
        setDataAttribute(key: string, value: string): void

        /**
         * Sets the inner HTML
         * @param value - The value to set
         */
        setInnerHTML(value: string): void

        /**
         * Appends the child element
         * @param child - The child element to append
         */
        appendChild(child: DOMElement): void

        /**
         * Removes the child element
         * @param child - The child element to remove
         */
        removeChild(child: DOMElement): void

        /**
         * Removes a data attribute (data-* attribute)
         * @param key - The data attribute key (without the data- prefix)
         */
        removeDataAttribute(key: string): void

        /**
         * Checks if the element has a data attribute
         * @param key - The data attribute key (without the data- prefix)
         * @returns A promise that resolves to true if the data attribute exists
         */
        hasDataAttribute(key: string): Promise<boolean>

        // DOM manipulation
        /**
         * Appends a child to the element
         * @param child - The child to append
         */
        append(child: DOMElement): void

        /**
         * Inserts a sibling before the element
         * @param sibling - The sibling to insert
         */
        before(sibling: DOMElement): void

        /**
         * Inserts a sibling after the element
         * @param sibling - The sibling to insert
         */
        after(sibling: DOMElement): void

        /**
         * Removes the element
         */
        remove(): void

        /**
         * Gets the parent of the element
         * @returns The parent of the element
         */
        getParent(opts?: DOMQueryElementOptions): Promise<DOMElement | null>

        /**
         * Gets the children of the element
         * @returns The children of the element
         */
        getChildren(opts?: DOMQueryElementOptions): Promise<DOMElement[]>

        // Events
        addEventListener(event: string, callback: (event: any) => void): () => void

        /**
         * Queries the DOM for elements that are descendants of this element and match the selector
         * @param selector - The selector to query
         * @returns A promise that resolves to an array of DOM elements
         */
        query(selector: string): Promise<DOMElement[]>

        /**
         * Queries the DOM for a single element that is a descendant of this element and matches the selector
         * @param selector - The selector to query
         * @returns A promise that resolves to a DOM element or null if no element is found
         */
        queryOne(selector: string): Promise<DOMElement | null>
    }

    interface DOMQueryElementOptions {
        /**
         * Whether to include the innerHTML of the element
         */
        withInnerHTML?: boolean

        /**
         * Whether to include the outerHTML of the element
         */
        withOuterHTML?: boolean

        /**
         * Whether to assign plugin-element IDs to all child elements
         * This is useful when you need to interact with child elements directly
         */
        identifyChildren?: boolean
    }

    // DOM interface
    interface DOM {
        /**
         * Queries the DOM for elements matching the selector
         * @param selector - The selector to query
         * @returns A promise that resolves to an array of DOM elements
         */
        query(selector: string, opts?: DOMQueryElementOptions): Promise<DOMElement[]>

        /**
         * Queries the DOM for a single element matching the selector
         * @param selector - The selector to query
         * @returns A promise that resolves to a DOM element or null if no element is found
         */
        queryOne(selector: string, opts?: DOMQueryElementOptions): Promise<DOMElement | null>

        /**
         * Observes changes to the DOM
         * @param selector - The selector to observe
         * @param callback - The callback to call when the DOM changes
         * @returns A tuple containing a function to stop observing the DOM and a function to refetch observed elements
         */
        observe(selector: string, callback: (elements: DOMElement[]) => void, opts?: DOMQueryElementOptions): [() => void, () => void]

        /**
         * Observes changes to the DOM in the viewport
         * @param selector - The selector to observe
         * @param callback - The callback to call when the DOM changes
         * @returns A tuple containing a function to stop observing the DOM and a function to refetch observed elements
         */
        observeInView(selector: string,
            callback: (elements: DOMElement[]) => void,
            opts?: DOMQueryElementOptions & { margin?: string },
        ): [() => void, () => void]

        /**
         * Creates a new DOM element
         * @param tagName - The tag name of the element
         * @returns A promise that resolves to a DOM element
         */
        createElement(tagName: string): Promise<DOMElement>

        /**
         * Returns the DOM element from an element ID
         * Note: No properties are available on this element, only methods, and there is no guarantee that the element exists
         * @param elementId - The ID of the element
         * @returns A DOM element
         */
        asElement(elementId: string): Omit<DOMElement, "tagName" | "attributes" | "innerHTML">

        /**
         * Called when the DOM is ready
         * @param callback - The callback to call when the DOM is ready
         */
        onReady(callback: () => void): void

        /**
         * Called when the main tab is ready or has changed.
         * This is similar to ctx.dom.onReady but is called each time a tab gains "main" status.
         * @param callback - The callback to call
         */
        onMainTabReady(callback: () => void): void

        viewport: {
            /** Returns the current viewport size synchronously (blocking) */
            getSize(): { width: number, height: number }

            /**
             * Observes changes to the viewport size
             * @param cb - The callback to call when the viewport size changes
             * @returns A function to stop observing the viewport size
             */
            onResize(cb: (size: { width: number, height: number }) => void): () => void

        }
    }

    interface Notification {
        /**
         * Sends a system notification
         * @param message - The message to send
         */
        send(message: string): void
    }

    interface Anime {
        /**
         * Get an anime entry
         * @param mediaId - The ID of the anime
         * @returns A promise that resolves to an anime entry
         * @throws Error if the entry is not found
         */
        getAnimeEntry(mediaId: number): Promise<$app.Anime_Entry>

        /**
         * Get raw anime metadata from metadata provider
         * @param from - "anilist" | "mal" | "kitsu" | "anidb"
         * @param mediaId - The ID
         */
        getAnimeMetadata(from: "anilist" | "mal" | "kitsu" | "anidb", mediaId: number): Promise<$app.Metadata_AnimeMetadata | undefined>

        /**
         * Clears episode metadata cache.
         * Note: To clear the anime entry cache, use $anilist.clearCache() (requires 'anilist' permission).
         */
        clearEpisodeMetadataCache(): void
    }

    interface Manga {
        /**
         * Get an manga entry
         * @param mediaId - The ID of the manga
         * @returns A promise that resolves to a manga entry
         * @throws Error if the entry is not found
         */
        getMangaEntry(mediaId: number): Promise<$app.Manga_Entry>

        /**
         * Get a chapter container for a manga.
         * This caches the chapter container if it exists.
         * @param opts - The options for the chapter container
         * @returns A promise that resolves to a chapter container
         * @throws Error if the chapter container is not found or if the manga repository is not found
         */
        getChapterContainer(opts: {
            mediaId: number;
            provider: string;
            titles?: string[];
            year?: number;
        }): Promise<$app.Manga_ChapterContainer | null>

        /**
         * Get the downloaded chapters
         * @returns A promise that resolves to an array of chapters grouped by provider and manga ID
         */
        getDownloadedChapters(): Promise<$app.Manga_ChapterContainer[]>

        /**
         * Get the manga collection
         * @returns A promise that resolves to a manga collection
         */
        getCollection(): Promise<$app.Manga_Collection>

        /**
         * Deletes all cached chapters and refetches them based on the selected provider for that manga.
         *
         * @param selectedProviderMap - A map of manga IDs to provider IDs. Previously cached chapters for providers not in the map will be
         *     deleted.
         * @returns A promise that resolves to void
         */
        refreshChapters(selectedProviderMap: Record<number, string>): Promise<void>

        /**
         * Empties cached chapters for a manga
         * @param mediaId - The ID of the manga
         * @returns A promise that resolves to void
         */
        emptyCache(mediaId: number): Promise<void>

        /**
         * Get the manga providers
         * @returns A map of provider IDs to provider names
         */
        getProviders(): Record<string, string>
    }

    interface Discord {
        /**
         * Set the manga activity
         * @param activity - The manga activity to set
         */
        setMangaActivity(activity: $app.DiscordRPC_MangaActivity): void

        /**
         * Set the anime activity with progress tracking.
         * @param activity - The anime activity to set
         */
        setAnimeActivity(activity: $app.DiscordRPC_AnimeActivity): void

        /**
         * Update the current anime activity progress.
         * Pausing the activity will cancel the activity on discord but retain the it in memory.
         * @param progress - The progress of the anime in seconds
         * @param duration - The duration of the anime in seconds
         * @param paused - Whether the anime is paused
         */
        updateAnimeActivity(progress: number, duration: number, paused: boolean): void

        /**
         * Set the anime activity (no progress tracking)
         * @param activity - The anime activity to set
         */
        setLegacyAnimeActivity(activity: $app.DiscordRPC_LegacyAnimeActivity): void

        /**
         * Cancels the current activity by closing the discord RPC client
         */
        cancelActivity(): void
    }

    interface Continuity {
        /**
         * Get the watch history.
         * The returned object is not in any particular order.
         * @returns A record of media IDs to watch history items
         * @throws Error if something goes wrong
         */
        getWatchHistory(): Record<number, $app.Continuity_WatchHistoryItem>

        /**
         * Delete a watch history item
         * @param mediaId - The ID of the media
         * @throws Error if something goes wrong
         */
        deleteWatchHistoryItem(mediaId: number): void

        /**
         * Update a watch history item
         * @param mediaId - The ID of the media
         * @param watchHistoryItem - The watch history item to update
         * @throws Error if something goes wrong
         */
        updateWatchHistoryItem(mediaId: number, watchHistoryItem: $app.Continuity_WatchHistoryItem): void

        /**
         * Get a watch history item
         * @param mediaId - The ID of the media
         * @returns The watch history item
         * @throws Error if something goes wrong
         */
        getWatchHistoryItem(mediaId: number): $app.Continuity_WatchHistoryItem | undefined
    }

    interface AutoScanner {
        /**
         * Notify the auto scanner to scan the libraries if it is enabled.
         * This is a non-blocking call that simply schedules a scan if one is not already running planned.
         */
        notify(): void
    }

    interface ExternalPlayerLink {
        /**
         * Open a URL in the external player.
         * @param url - The URL to open
         * @param mediaId - The ID of the media (used for the modal)
         * @param episodeNumber - The episode number (used for the modal)
         */
        open(url: string, mediaId: number, episodeNumber: number): void
    }

    interface AutoDownloader {
        /**
         * Run the auto downloader if it is enabled.
         * This is a non-blocking call.
         */
        run(): void
    }

    interface FillerManager {
        /**
         * Get the filler episodes for a media ID
         * @param mediaId - The media ID
         * @returns The filler episodes
         */
        getFillerEpisodes(mediaId: number): string[]

        /**
         * Set the filler episodes for a media ID
         * @param mediaId - The media ID
         * @param fillerEpisodes - The filler episodes
         */
        setFillerEpisodes(mediaId: number, fillerEpisodes: string[]): void

        /**
         * Check if an episode is a filler
         * @param mediaId - The media ID
         * @param episodeNumber - The episode number
         */
        isEpisodeFiller(mediaId: number, episodeNumber: number): boolean

        /**
         * Hydrate the filler data for an anime entry
         * @param e - The anime entry
         */
        hydrateFillerData(e: $app.Anime_Entry): void

        /**
         * Hydrate the filler data for an onlinestream episode
         * @param mId - The media ID
         * @param episodes - The episodes
         */
        hydrateOnlinestreamFillerData(mId: number, episodes: $app.Onlinestream_Episode[]): void

        /**
         * Remove the filler data for a media ID
         * @param mediaId - The media ID
         */
        removeFillerData(mediaId: number): void
    }

    interface TorrentClient {
        /**
         * Get all torrents
         * @returns A promise that resolves to an array of torrents
         */
        getTorrents(): Promise<$app.TorrentClient_Torrent[]>

        /**
         * Get the active torrents
         * @returns A promise that resolves to an array of active torrents
         */
        getActiveTorrents(): Promise<$app.TorrentClient_Torrent[]>

        /**
         * Pause some torrents
         * @param hashes - The hashes of the torrents to pause
         */
        pauseTorrents(hashes: string[]): Promise<void>

        /**
         * Resume some torrents
         * @param hashes - The hashes of the torrents to resume
         */
        resumeTorrents(hashes: string[]): Promise<void>

        /**
         * Deselect some files from a torrent
         * @param hash - The hash of the torrent
         * @param indices - The indices of the files to deselect
         */
        deselectFiles(hash: string, indices: number[]): Promise<void>

        /**
         * Get the files of a torrent
         * @param hash - The hash of the torrent
         * @returns A promise that resolves to an array of files
         */
        getFiles(hash: string): Promise<string[]>
    }

    type Intent =
        "primary"
        | "primary-subtle"
        | "alert"
        | "alert-subtle"
        | "warning"
        | "warning-subtle"
        | "success"
        | "success-subtle"
        | "white"
        | "white-subtle"
        | "gray"
        | "gray-subtle"
}

declare namespace $storage {
    /**
     * Sets a value in the storage.
     * @param key - The key to set
     * @param value - The value to set
     * @throws Error if something goes wrong
     */
    function set(key: string, value: any): void

    /**
     * Gets a value from the storage.
     * @param key - The key to get
     * @returns The value associated with the key
     * @throws Error if something goes wrong
     */
    function get<T = any>(key: string): T | undefined

    /**
     * Removes a value from the storage.
     * @param key - The key to remove
     * @throws Error if something goes wrong
     */
    function remove(key: string): void

    /**
     * Drops the database.
     * @throws Error if something goes wrong
     */
    function drop(): void

    /**
     * Clears all values from the storage.
     * @throws Error if something goes wrong
     */
    function clear(): void

    /**
     * Returns all keys in the storage.
     * @returns An array of all keys in the storage
     * @throws Error if something goes wrong
     */
    function keys(): string[]

    /**
     * Checks if a key exists in the storage.
     * @param key - The key to check
     * @returns True if the key exists, false otherwise
     * @throws Error if something goes wrong
     */
    function has(key: string): boolean
}

declare namespace $anilist {
    /**
     * Deletes all cached data.
     */
    function clearCache(): void

    /**
     * Refresh the anime collection.
     * This will cause the frontend to refetch queries that depend on the anime collection.
     */
    function refreshAnimeCollection(): void

    /**
     * Refresh the manga collection.
     * This will cause the frontend to refetch queries that depend on the manga collection.
     */
    function refreshMangaCollection(): void

    /**
     * Update a media list entry.
     * The anime/manga collection should be refreshed after updating the entry.
     */
    function updateEntry(
        mediaId: number,
        status: $app.AL_MediaListStatus | undefined,
        scoreRaw: number | undefined,
        progress: number | undefined,
        startedAt: $app.AL_FuzzyDateInput | undefined,
        completedAt: $app.AL_FuzzyDateInput | undefined,
    ): void

    /**
     * Update a media list entry's progress.
     * The anime/manga collection should be refreshed after updating the entry.
     */
    function updateEntryProgress(mediaId: number, progress: number, totalCount: number | undefined): void

    /**
     * Update a media list entry's repeat count.
     * The anime/manga collection should be refreshed after updating the entry.
     */
    function updateEntryRepeat(mediaId: number, repeat: number): void

    /**
     * Delete a media list entry.
     * The anime/manga collection should be refreshed after deleting the entry.
     */
    function deleteEntry(mediaId: number): void

    /**
     * Get the user's anime collection.
     * This collection does not include lists with no status (custom lists).
     */
    function getAnimeCollection(bypassCache: boolean): $app.AL_AnimeCollection

    /**
     * Same as [$anilist.getAnimeCollection] but includes lists with no status (custom lists).
     */
    function getRawAnimeCollection(bypassCache: boolean): $app.AL_AnimeCollection

    /**
     * Get the user's manga collection.
     * This collection does not include lists with no status (custom lists).
     */
    function getMangaCollection(bypassCache: boolean): $app.AL_MangaCollection

    /**
     * Same as [$anilist.getMangaCollection] but includes lists with no status (custom lists).
     */
    function getRawMangaCollection(bypassCache: boolean): $app.AL_MangaCollection

    /**
     * Get anime by ID
     */
    function getAnime(id: number): $app.AL_BaseAnime

    /**
     * Get manga by ID
     */
    function getManga(id: number): $app.AL_BaseManga

    /**
     * Get detailed anime info by ID
     */
    function getAnimeDetails(id: number): $app.AL_AnimeDetailsById_Media

    /**
     * Get detailed manga info by ID
     */
    function getMangaDetails(id: number): $app.AL_MangaDetailsById_Media

    /**
     * Get anime collection with relations
     */
    function getAnimeCollectionWithRelations(): $app.AL_AnimeCollectionWithRelations

    /**
     * Add media to collection.
     *
     * This will add the media to the collection with the status "PLANNING".
     *
     * The anime/manga collection should be refreshed after adding the media.
     */
    function addMediaToCollection(mediaIds: number[]): void

    /**
     * Get studio details
     */
    function getStudioDetails(studioId: number): $app.AL_StudioDetails

    /**
     * List anime based on search criteria
     */
    function listAnime(
        page: number | undefined,
        search: string | undefined,
        perPage: number | undefined,
        sort: $app.AL_MediaSort[] | undefined,
        status: $app.AL_MediaStatus[] | undefined,
        genres: string[] | undefined,
        averageScoreGreater: number | undefined,
        season: $app.AL_MediaSeason | undefined,
        seasonYear: number | undefined,
        format: $app.AL_MediaFormat | undefined,
        isAdult: boolean | undefined,
    ): $app.AL_ListAnime

    /**
     * List manga based on search criteria
     */
    function listManga(
        page: number | undefined,
        search: string | undefined,
        perPage: number | undefined,
        sort: $app.AL_MediaSort[] | undefined,
        status: $app.AL_MediaStatus[] | undefined,
        genres: string[] | undefined,
        averageScoreGreater: number | undefined,
        startDateGreater: string | undefined,
        startDateLesser: string | undefined,
        format: $app.AL_MediaFormat | undefined,
        countryOfOrigin: string | undefined,
        isAdult: boolean | undefined,
    ): $app.AL_ListManga

    /**
     * List recent anime
     */
    function listRecentAnime(
        page: number | undefined,
        perPage: number | undefined,
        airingAtGreater: number | undefined,
        airingAtLesser: number | undefined,
        notYetAired: boolean | undefined,
    ): $app.AL_ListRecentAnime

    /**
     * Make a custom GraphQL query
     */
    function customQuery<T = any>(body: Record<string, any>, token: string): T
}

/**
 * Cron
 */

declare namespace $cron {
    /**
     * Adds a cron job
     * @param id - The id of the cron job
     * @param cronExpr - The cron expression
     * @param fn - The function to call
     */
    function add(id: string, cronExpr: string, fn: () => void): void

    /**
     * Removes a cron job
     * @param id - The id of the cron job
     */
    function remove(id: string): void

    /**
     * Removes all cron jobs
     */
    function removeAll(): void

    /**
     * Gets the total number of cron jobs
     * @returns The total number of cron jobs
     */
    function total(): number

    /**
     * Starts the cron jobs, can be paused by calling stop()
     */
    function start(): void

    /**
     * Stops the cron jobs, can be resumed by calling start()
     */
    function stop(): void

    /**
     * Checks if the cron jobs have started
     * @returns True if the cron jobs have started, false otherwise
     */
    function hasStarted(): boolean
}

/**
 * Database
 */

declare namespace $database {

    namespace localFiles {
        /**
         * Gets the local files
         * @returns The local files
         */
        function getAll(): $app.Anime_LocalFile[]

        /**
         * Finds the local files by a filter function
         * @param filterFn - The filter function
         * @returns The local files
         */
        function findBy(filterFn: (file: $app.Anime_LocalFile) => boolean): $app.Anime_LocalFile[]

        /**
         * Saves the modified local files. This only works if the local files are already in the database.
         * @param files - The local files to save
         */
        function save(files: $app.Anime_LocalFile[]): $app.Anime_LocalFile[]

        /**
         * Inserts the local files as a new entry
         * @param files - The local files to insert
         */
        function insert(files: $app.Anime_LocalFile[]): $app.Anime_LocalFile[]
    }

    namespace anilist {
        /**
         * Get the Anilist token
         *
         * Permissions needed: anilist-token
         *
         * @returns The Anilist token
         */
        function getToken(): string

        /**
         * Get the Anilist username
         */
        function getUsername(): string
    }

    namespace autoDownloaderRules {
        /**
         * Gets all auto downloader rules
         */
        function getAll(): $app.Anime_AutoDownloaderRule[]

        /**
         * Gets an auto downloader rule by the database id
         * @param id - The id of the auto downloader rule in the database
         * @returns The auto downloader rule
         */
        function get(id: number): $app.Anime_AutoDownloaderRule | undefined

        /**
         * Gets all auto downloader rules by media id
         * @param mediaId - The id of the media
         * @returns The auto downloader rules
         */
        function getByMediaId(mediaId: number): $app.Anime_AutoDownloaderRule[]

        /**
         * Inserts an auto downloader rule
         * @param rule - The auto downloader rule to insert
         */
        function insert(rule: Omit<$app.Anime_AutoDownloaderRule, "dbId">): void

        /**
         * Updates an auto downloader rule
         * @param id - The id of the auto downloader rule in the database
         * @param rule - The auto downloader rule to update
         */
        function update(id: number, rule: Omit<$app.Anime_AutoDownloaderRule, "dbId">): void

        /**
         * Deletes an auto downloader rule
         * @param id - The id of the auto downloader rule in the database
         */
        function remove(id: number): void
    }

    namespace autoDownloaderItems {
        /**
         * Gets all auto downloader items
         */
        function getAll(): $app.Models_AutoDownloaderItem[]

        /**
         * Gets an auto downloader item by id
         * @param id - The id of the auto downloader item in the database
         */
        function get(id: number): $app.Models_AutoDownloaderItem | undefined

        /**
         * Gets all auto downloader items by media id
         * @param mediaId - The id of the media
         */
        function getByMediaId(mediaId: number): $app.Models_AutoDownloaderItem[]

        /**
         * Inserts an auto downloader item
         * @param item - The auto downloader item to insert
         */
        function insert(item: $app.Models_AutoDownloaderItem): void

        /**
         * Deletes an auto downloader item
         * @param id - The id of the auto downloader item in the database
         */
        function remove(id: number): void
    }

    namespace silencedMediaEntries {
        /**
         * Gets all silenced media entry ids
         */
        function getAllIds(): number[]

        /**
         * Checks if a media entry is silenced
         * @param mediaId - The id of the media
         * @returns True if the media entry is silenced, false otherwise
         */
        function isSilenced(mediaId: number): boolean

        /**
         * Sets a media entry as silenced
         * @param mediaId - The id of the media
         */
        function setSilenced(mediaId: number, silenced: boolean): void
    }

    namespace mediaFillers {
        /**
         * Gets all media fillers
         */
        function getAll(): Record<number, MediaFillerItem>

        /**
         * Gets a media filler by media id
         * @param mediaId - The id of the media
         */
        function get(mediaId: number): MediaFillerItem | undefined

        /**
         * Inserts a media filler
         * @param provider - The provider of the media filler
         * @param mediaId - The id of the media
         * @param slug - The slug of the media filler
         * @param fillerEpisodes - The filler episodes
         */
        function insert(provider: string, mediaId: number, slug: string, fillerEpisodes: string[]): void

        /**
         * Deletes a media filler
         * @param mediaId - The id of the media
         */
        function remove(mediaId: number): void
    }

    interface MediaFillerItem {
        /**
         * The id of the media filler in the database
         */
        dbId: number
        /**
         * The provider of the media filler
         */
        provider: string
        /**
         * The id of the media
         */
        mediaId: number
        /**
         * The slug of the media filler
         */
        slug: string
        /**
         * The filler episodes
         */
        fillerEpisodes: string[]
        /**
         * Date and time the filler data was last fetched
         */
        lastFetchedAt: string
    }
}

declare namespace $app {
    /**
     * Gets the version of the app
     * @returns The version of the app
     */
    function getVersion(): string

    /**
     * Gets the version name of the app
     * @returns The version name of the app
     */
    function getVersionName(): string

    /**
     * Invalidates the queries on the client
     * @param queryKeys - Keys of the queries to invalidate
     */
    function invalidateClientQuery(queryKeys: string[]): void
}

declare namespace $ui {
    // Video Core Types

    type PlayerType = "native" | "web"
    type PlaybackType = "localfile" | "torrent" | "debrid" | "nakama" | "onlinestream"

    type VideoEventType =
        | "video-loaded"
        | "video-loaded-metadata"
        | "video-can-play"
        | "video-paused"
        | "video-resumed"
        | "video-status"
        | "video-completed"
        | "video-fullscreen"
        | "video-pip"
        | "video-subtitle-track"
        | "video-media-caption-track"
        | "video-anime-4k"
        | "video-audio-track"
        | "video-ended"
        | "video-seeked"
        | "video-error"
        | "video-terminated"
        | "video-playback-state"
        | "subtitle-file-uploaded"
        | "video-playlist"

    interface BaseVideoEvent {
        playerType: PlayerType
        playbackType: PlaybackType
        playbackId: string
        clientId: string
    }

    interface VideoLoadedEvent extends BaseVideoEvent {
        clientId: string
        state: PlaybackState
    }

    interface VideoPlaybackStateEvent extends BaseVideoEvent {
        clientId: string
        state: PlaybackState
    }

    interface VideoPausedEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
    }

    interface VideoResumedEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
    }

    interface VideoEndedEvent extends BaseVideoEvent {
        autoNext: boolean
    }

    interface VideoErrorEvent extends BaseVideoEvent {
        error: string
    }

    interface VideoSeekedEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
        paused: boolean
    }

    interface VideoStatusEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
        paused: boolean
    }

    interface VideoLoadedMetadataEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
        paused: boolean
    }

    interface VideoCanPlayEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
        paused: boolean
    }

    interface SubtitleFileUploadedEvent extends BaseVideoEvent {
        filename: string
        content: string
    }

    interface VideoTerminatedEvent extends BaseVideoEvent {
    }

    interface VideoCompletedEvent extends BaseVideoEvent {
        currentTime: number
        duration: number
    }

    interface VideoAudioTrackEvent extends BaseVideoEvent {
        trackNumber: number
        isHLS: boolean
    }

    interface VideoSubtitleTrackEvent extends BaseVideoEvent {
        trackNumber: number
        kind: "file" | "event"
    }

    interface VideoMediaCaptionTrackEvent extends BaseVideoEvent {
        trackIndex: number
    }

    interface VideoFullscreenEvent extends BaseVideoEvent {
        fullscreen: boolean
    }

    interface VideoPipEvent extends BaseVideoEvent {
        pip: boolean
    }

    interface VideoAnime4KEvent extends BaseVideoEvent {
        option: string
    }

    interface VideoPlaylistEvent extends BaseVideoEvent {
        playlist: VideoPlaylistState | null
    }

    interface VideoTextTracksEvent extends BaseVideoEvent {
        textTracks: VideoTextTrack[]
    }

    type VideoEvent =
        | VideoLoadedEvent
        | VideoPlaybackStateEvent
        | VideoPausedEvent
        | VideoResumedEvent
        | VideoEndedEvent
        | VideoErrorEvent
        | VideoSeekedEvent
        | VideoStatusEvent
        | VideoLoadedMetadataEvent
        | VideoCanPlayEvent
        | SubtitleFileUploadedEvent
        | VideoTerminatedEvent
        | VideoCompletedEvent
        | VideoAudioTrackEvent
        | VideoSubtitleTrackEvent
        | VideoMediaCaptionTrackEvent
        | VideoFullscreenEvent
        | VideoPipEvent
        | VideoAnime4KEvent
        | VideoPlaylistEvent
        | VideoTextTracksEvent

    interface VideoSubtitleTrack {
        index: number
        src?: string
        content?: string
        label: string
        language: string
        type?: "srt" | "vtt" | "ass" | "ssa"
        default?: boolean
        useLibassRenderer?: boolean
    }

    interface VideoTextTrack {
        number: number,
        type: "subtitles" | "captions",
        label: string,
        language: string,
    }

    interface VideoSource {
        index: number
        resolution: string
        url?: string
        label?: string
        moreInfo?: string
    }

    interface VideoInitialState {
        currentTime?: number
        paused?: boolean
    }

    interface OnlinestreamParams {
        mediaId: number
        episodeNumber: number
        provider: string
        server: string
        quality: string
        dubbed: boolean
    }

    export type MkvTrackType = "video" | "audio" | "subtitle" | "logo" | "buttons" | "complex" | "unknown"

    export type MkvAttachmentType = "font" | "subtitle" | "other"

    export interface MkvTrackInfo {
        number: number
        uid: number
        type: MkvTrackType
        codecID: string
        name?: string
        language?: string
        languageIETF?: string
        default: boolean
        forced: boolean
        enabled: boolean
        codecPrivate?: string
        video?: any
        audio?: any
        contentEncodings?: any
        defaultDuration?: number
    }

    export interface MkvChapterInfo {
        uid: number
        start: number
        end?: number
        text?: string
        languages?: string[]
        languagesIETF?: string[]
        editionUID?: number
    }

    export interface MkvAttachmentInfo {
        uid: number
        filename: string
        mimetype: string
        size: number
        description?: string
        type?: MkvAttachmentType
        data?: Uint8Array
        isCompressed?: boolean
    }

    export interface MkvMetadata {
        title?: string
        duration: number
        timecodeScale: number
        muxingApp?: string
        writingApp?: string
        tracks: MkvTrackInfo[]
        videoTracks: MkvTrackInfo[]
        audioTracks: MkvTrackInfo[]
        subtitleTracks: MkvTrackInfo[]
        chapters: MkvChapterInfo[]
        attachments: MkvAttachmentInfo[]
        mimeCodec?: string
        error?: Error
    }

    interface VideoPlaybackInfo {
        id: string
        playbackType: PlaybackType
        streamUrl: string
        mkvMetadata?: MkvMetadata
        localFile?: $app.Anime_LocalFile
        onlinestreamParams?: OnlinestreamParams
        subtitleTracks: VideoSubtitleTrack[]
        videoSources: VideoSource[]
        selectedVideoSource?: number
        playlistExternalEpisodeNumbers: number[]
        disableRestoreFromContinuity?: boolean
        initialState?: VideoInitialState
        media?: $app.AL_BaseAnime
        episode?: $app.Anime_Episode
        streamType: "native" | "hls" | "unknown"
        isNakamaWatchParty?: boolean
    }

    interface VideoPlaylistState {
        type: PlaybackType
        episodes: $app.Anime_Episode[]
        previousEpisode?: $app.Anime_Episode
        nextEpisode?: $app.Anime_Episode
        currentEpisode: $app.Anime_Episode
        animeEntry?: $app.Anime_Entry
    }

    interface PlaybackStatus {
        id: string
        clientId: string
        paused: boolean
        currentTime: number
        duration: number
    }

    interface PlaybackState {
        clientId: string
        playerType: PlayerType
        playbackInfo: VideoPlaybackInfo
    }

    interface VideoCore {
        /**
         * Adds an event listener for video-loaded events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-loaded", callback: (event: VideoLoadedEvent) => void): void

        /**
         * Adds an event listener for video-playback-state events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-playback-state", callback: (event: VideoPlaybackStateEvent) => void): void

        /**
         * Adds an event listener for video-paused events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-paused", callback: (event: VideoPausedEvent) => void): void

        /**
         * Adds an event listener for video-resumed events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-resumed", callback: (event: VideoResumedEvent) => void): void

        /**
         * Adds an event listener for video-ended events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-ended", callback: (event: VideoEndedEvent) => void): void

        /**
         * Adds an event listener for video-error events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-error", callback: (event: VideoErrorEvent) => void): void

        /**
         * Adds an event listener for video-seeked events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-seeked", callback: (event: VideoSeekedEvent) => void): void

        /**
         * Adds an event listener for video-status events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-status", callback: (event: VideoStatusEvent) => void): void

        /**
         * Adds an event listener for video-loaded-metadata events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-loaded-metadata", callback: (event: VideoLoadedMetadataEvent) => void): void

        /**
         * Adds an event listener for video-can-play events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-can-play", callback: (event: VideoCanPlayEvent) => void): void

        /**
         * Adds an event listener for subtitle-file-uploaded events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "subtitle-file-uploaded", callback: (event: SubtitleFileUploadedEvent) => void): void

        /**
         * Adds an event listener for video-terminated events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-terminated", callback: (event: VideoTerminatedEvent) => void): void

        /**
         * Adds an event listener for video-completed events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-completed", callback: (event: VideoCompletedEvent) => void): void

        /**
         * Adds an event listener for video-audio-track events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-audio-track", callback: (event: VideoAudioTrackEvent) => void): void

        /**
         * Adds an event listener for video-subtitle-track events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-subtitle-track", callback: (event: VideoSubtitleTrackEvent) => void): void

        /**
         * Adds an event listener for video-media-caption-track events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-media-caption-track", callback: (event: VideoMediaCaptionTrackEvent) => void): void

        /**
         * Adds an event listener for video-fullscreen events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-fullscreen", callback: (event: VideoFullscreenEvent) => void): void

        /**
         * Adds an event listener for video-pip events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-pip", callback: (event: VideoPipEvent) => void): void

        /**
         * Adds an event listener for video-anime-4k events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-anime-4k", callback: (event: VideoAnime4KEvent) => void): void

        /**
         * Adds an event listener for video-playlist events
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: "video-playlist", callback: (event: VideoPlaylistEvent) => void): void

        /**
         * Adds an event listener for any video event (fallback)
         * @param eventType - The event type to listen for
         * @param callback - The callback function to execute when the event is triggered
         */
        addEventListener(eventType: VideoEventType, callback: (event: VideoEvent) => void): void

        /**
         * Removes an event listener for the specified event type
         * @param eventType - The event type to stop listening for
         */
        removeEventListener(eventType: VideoEventType): void

        // Playback control methods

        /**
         * Pauses the video playback
         */
        pause(): void

        /**
         * Resumes the video playback
         */
        resume(): void

        /**
         * Seeks forward or backward by the specified number of seconds
         * @param seconds - Number of seconds to seek (positive for forward, negative for backward)
         */
        seek(seconds: number): void

        /**
         * Seeks to an absolute position in the video
         * @param seconds - The absolute position in seconds
         */
        seekTo(seconds: number): void

        /**
         * Terminates the current video playback
         */
        terminate(): void

        /**
         * Plays the specified playlist episode
         * @param which - "next", "previous", or the AniDB Episode ID
         */
        playEpisodeFromPlaylist(which: string): void

        // UI control methods

        /**
         * Sets the fullscreen state of the video player
         * @param fullscreen - Whether to enable fullscreen
         */
        setFullscreen(fullscreen: boolean): void

        /**
         * Sets the picture-in-picture state of the video player
         * @param pip - Whether to enable picture-in-picture
         */
        setPip(pip: boolean): void

        /**
         * Shows a message in the video player
         * @param message - The message to display
         * @param milliseconds - The duration of the message in milliseconds (Default: 2000)
         */
        showMessage(message: string, milliseconds?: number): void

        // Track control methods

        /**
         * Sets the active subtitle track
         * @param trackNumber - The track number to activate
         */
        setSubtitleTrack(trackNumber: number): void

        /**
         * Adds a subtitle track to the video player.
         * @important Use addExternalSubtitleTrack instead.
         * @param track - The subtitle track information
         */
        addSubtitleTrack(track: any): void

        /**
         * Adds an external subtitle track to the video player
         * @param track - The external subtitle track information
         */
        addExternalSubtitleTrack(track: Omit<VideoSubtitleTrack, "index" | "useLibassRenderer">): void

        /**
         * Sets the active media caption track
         * @param trackIndex - The track index to activate
         */
        setMediaCaptionTrack(trackIndex: number): void

        /**
         * Adds a media caption track to the video player
         * @important Use addExternalSubtitleTrack instead.
         * @param track - The media caption track information
         */
        addMediaCaptionTrack(track: any): void

        /**
         * Sets the active audio track
         * @param trackNumber - The track number to activate
         */
        setAudioTrack(trackNumber: number): void

        // State request methods

        /**
         * Requests the current fullscreen state from the player
         */
        sendGetFullscreen(): void

        /**
         * Requests the current picture-in-picture state from the player
         */
        sendGetPip(): void

        /**
         * Requests the current Anime4K state from the player
         */
        sendGetAnime4K(): void

        /**
         * Requests the current subtitle track from the player
         */
        sendGetSubtitleTrack(): void

        /**
         * Requests the current audio track from the player
         */
        sendGetAudioTrack(): void

        /**
         * Requests the current media caption track from the player
         */
        sendGetMediaCaptionTrack(): void

        /**
         * Requests the current playback state from the player
         */
        sendGetPlaybackState(): void

        // Async getters

        /**
         * Gets the current text tracks
         * @returns A promise that resolves to the text tracks or undefined
         */
        getTextTracks(): Promise<VideoTextTrack[] | undefined>

        /**
         * Gets the current playlist state
         * @returns A promise that resolves to the playlist state or undefined
         */
        getPlaylist(): Promise<VideoPlaylistState | undefined>

        /**
         * Pulls the current playback status from the player
         * @returns A promise that resolves to the video status event
         */
        pullStatus(): Promise<VideoStatusEvent | undefined>

        // Sync getters

        /**
         * Gets the current playback status
         * @returns The playback status or undefined
         */
        getPlaybackStatus(): PlaybackStatus | undefined

        /**
         * Gets the current playback state
         * @returns The playback state or undefined
         */
        getPlaybackState(): PlaybackState | undefined

        /**
         * Gets the current playback information
         * @returns The playback information or undefined
         */
        getCurrentPlaybackInfo(): VideoPlaybackInfo | undefined

        /**
         * Gets the current media being played
         * @returns The media information or undefined
         */
        getCurrentMedia(): $app.AL_BaseAnime | undefined

        /**
         * Gets the current client ID
         * @returns The client ID or empty string
         */
        getCurrentClientId(): string

        /**
         * Gets the current player type
         * @returns The player type or empty string
         */
        getCurrentPlayerType(): PlayerType | ""

        /**
         * Gets the current playback type
         * @returns The playback type or empty string
         */
        getCurrentPlaybackType(): PlaybackType | ""
    }
}

