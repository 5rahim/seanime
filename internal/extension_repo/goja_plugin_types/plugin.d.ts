/// <reference path="app.d.ts" />

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
         * Discord
         */
        discord: Discord

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
         * Registers a field reference for field components.
         * @param fieldName - The name of the field
         * @returns A field reference object
         */
        fieldRef<T extends any = string>(): FieldRef<T>

        /**
         * Creates a new tray icon.
         * @param options - The options for the tray icon.
         * @returns A tray icon object.
         */
        newTray(options: TrayOptions): Tray

        /**
         * Creates a new command palette.
         * @param options - The options for the command palette
         * @returns A command palette object
         */
        newCommandPalette(options: CommandPaletteOptions): CommandPalette
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
        input: InputComponentFunction
        select: SelectComponentFunction
        checkbox: CheckboxComponentFunction
        radioGroup: RadioGroupComponentFunction
        switch: SwitchComponentFunction

        /** Invoked when the tray icon is clicked */
        onClick(cb: () => void): void

        /** Invoked when the tray icon is opened */
        onOpen(cb: () => void): void

        /** Invoked when the tray icon is closed */
        onClose(cb: () => void): void

        /** Registers the render function for the tray content */
        render(fn: () => void): void

        /** Schedules a re-render of the tray content */
        update(): void

        /** Opens the tray */
        open(): void

        /** Closes the tray */
        close(): void

        /** Updates the badge number of the tray icon. 0 = no badge. Default intent is "info". */
        updateBadge(options: { number: number, intent?: "success" | "error" | "warning" | "info" }): void
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
         * Note that this does not stop/close the media player.
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

    interface Action {
        /**
         * Creates a new button for the anime page
         * @param props - Button properties
         */
        newAnimePageButton(props: { label: string, intent?: Intent, style?: Record<string, string> }): ActionObject<{ media: $app.AL_BaseAnime }>

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
        newMangaPageButton(props: { label: string, intent?: Intent, style?: Record<string, string> }): ActionObject<{ media: $app.AL_BaseManga }>
    }

    interface ActionObject<E extends any = {}> {
        /** Mounts the action to make it visible */
        mount(): void

        /** Unmounts the action to hide it */
        unmount(): void

        /** Sets the label of the action */
        setLabel(label: string): void

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
     * @default size="md"
     */
    type InputComponentFunction = {
        (props: { label?: string, placeholder?: string } & FieldComponentProps): void
        (label: string, placeholder: string, props?: FieldComponentProps): void
    }
    /**
     * @default size="md"
     */
    type SelectComponentFunction = {
        (props: { label?: string, placeholder?: string, options: { label: string, value: string }[] } & FieldComponentProps): void
        (label: string, options: { placeholder?: string, value?: string }[], props?: FieldComponentProps): void
    }
    /**
     * @default size="md"
     */
    type CheckboxComponentFunction = {
        (props: { label?: string } & FieldComponentProps<boolean>): void
        (label: string, props?: FieldComponentProps<boolean>): void
    }
    /**
     * @default size="md"
     */
    type RadioGroupComponentFunction = {
        (props: { label?: string, options: { label: string, value: string }[] } & FieldComponentProps): void
        (label: string, options: { label: string, value: string }[], props?: FieldComponentProps): void
    }
    /**
     * @default side="right"
     * @default size="sm"
     */
    type SwitchComponentFunction = {
        (props: { label?: string, side?: "left" | "right" } & FieldComponentProps<boolean>): void
        (label: string, props?: { side?: "left" | "right" } & FieldComponentProps<boolean>): void
    }

    // DOM Element interface
    interface DOMElement {
        id: string
        tagName: string
        attributes: Record<string, string>
        // children: DOMElement[]
        textContent?: string
        innerHTML?: string

        // Properties
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
         * @param className - The class to add
         */
        addClass(className: string): void

        /**
         * Checks if the element has a class
         * @param className - The class to check
         * @returns A promise that resolves to true if the class exists
         */
        hasClass(className: string): Promise<boolean>

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
    }

    interface Notification {
        /**
         * Sends a notification
         * @param message - The message to send
         */
        send(message: string): void
    }

    interface Manga {
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
    }

    interface Discord {
        /**
         * Set the manga activity
         */
        setMangaActivity(opts: $app.DiscordRPC_MangaActivity): void

        /**
         * Set the anime activity
         */
        setAnimeActivity(opts: $app.DiscordRPC_AnimeActivity): void

        /**
         * Cancels the current activity
         */
        cancelActivity(): void
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

declare namespace $store {
    /**
     * Sets a value in the store.
     * @param key - The key to set
     * @param value - The value to set
     */
    function set(key: string, value: any): void

    /**
     * Gets a value from the store.
     * @param key - The key to get
     * @returns The value associated with the key
     */
    function get<T = any>(key: string): T

    /**
     * Checks if a key exists in the store.
     * @param key - The key to check
     * @returns True if the key exists, false otherwise
     */
    function has(key: string): boolean

    /**
     * Gets a value from the store or sets it if it doesn't exist.
     * @param key - The key to get or set
     * @param setFunc - The function to set the value
     * @returns The value associated with the key
     */
    function getOrSet<T = any>(key: string, setFunc: () => T): T

    /**
     * Sets a value in the store if it's less than the limit.
     * @param key - The key to set
     * @param value - The value to set
     * @param maxAllowedElements - The maximum allowed elements
     */
    function setIfLessThanLimit<T = any>(key: string, value: T, maxAllowedElements: number): boolean

    /**
     * Unmarshals a JSON string.
     * @param data - The JSON string to unmarshal
     */
    function unmarshalJSON(data: string): void

    /**
     * Marshals a value to a JSON string.
     * @param value - The value to marshal
     * @returns The JSON string
     */
    function marshalJSON(value: any): string

    /**
     * Resets the store.
     */
    function reset(): void

    /**
     * Gets all values from the store.
     * @returns An array of all values in the store
     */
    function values(): any[]
}

/**
 * Replaces the reference of the value with the new value.
 * @param value - The value to replace
 * @param newValue - The new value
 */
declare function $replace<T = any>(value: T, newValue: T): void

/**
 * Creates a deep copy of the value.
 * @param value - The value to copy
 * @returns A deep copy of the value
 */
declare function $clone<T = any>(value: T): T

/**
 * Converts a value to a string
 * @param value - The value to convert
 * @returns The string representation of the value
 */
declare function $toString(value: any): string

/**
 * Sleeps for a specified amount of time
 * @param milliseconds - The amount of time to sleep in milliseconds
 */
declare function $sleep(milliseconds: number): void

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

declare namespace $habari {

    interface Metadata {
        season_number?: string[];
        part_number?: string[];
        title?: string;
        formatted_title?: string;
        anime_type?: string[];
        year?: string;
        audio_term?: string[];
        device_compatibility?: string[];
        episode_number?: string[];
        other_episode_number?: string[];
        episode_number_alt?: string[];
        episode_title?: string;
        file_checksum?: string;
        file_extension?: string;
        file_name?: string;
        language?: string[];
        release_group?: string;
        release_information?: string[];
        release_version?: string[];
        source?: string[];
        subtitles?: string[];
        video_resolution?: string;
        video_term?: string[];
        volume_number?: string[];
    }

    /**
     * Parses a filename and returns the metadata
     * @param filename - The filename to parse
     * @returns The metadata
     */
    function parse(filename: string): Metadata
}
