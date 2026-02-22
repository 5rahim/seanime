/**
 * Is offline
 */
declare const __isOffline__: boolean

/**
 * Fetch
 */
declare function fetch(url: string, options?: FetchOptions): Promise<FetchResponse>

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
 * Converts a value to a bytes array
 * @param value - The value to convert
 * @returns The bytes array
 */
declare function $toBytes(value: any): Uint8Array

/**
 * Sleeps for a specified amount of time
 * @param milliseconds - The amount of time to sleep in milliseconds
 */
declare function $sleep(milliseconds: number): void

declare function $await<T>(promise: Promise<T>): void

/**
 *
 * @param model
 */
declare function $arrayOf<T>(model: T): T[]

/**
 * Marshals and unmarshals a value to a JSON string
 * @param data - The value to marshal
 * @param dst - The destination to unmarshal the value to. Must be a reference.
 * @throws If unmarshalling fails
 */
declare function $unmarshalJSON(data: any, dst: any): void

/**
 * Get a user preference
 * @param key The key of the preference
 * @returns The value of the preference set by the user, the default value if it is not set, or undefined.
 */
declare function $getUserPreference(key: string): string | undefined;

/**
 * Habari
 */

declare namespace $habari {

    interface Metadata {
        season_number?: string[]
        part_number?: string[]
        title?: string
        formatted_title?: string
        anime_type?: string[]
        year?: string
        audio_term?: string[]
        device_compatibility?: string[]
        episode_number?: string[]
        other_episode_number?: string[]
        episode_number_alt?: string[]
        episode_title?: string
        file_checksum?: string
        file_extension?: string
        file_name?: string
        language?: string[]
        release_group?: string
        release_information?: string[]
        release_version?: string[]
        source?: string[]
        subtitles?: string[]
        video_resolution?: string
        video_term?: string[]
        volume_number?: string[]
    }

    /**
     * Parses a filename and returns the metadata
     * @param filename - The filename to parse
     * @returns The metadata
     */
    function parse(filename: string): Metadata
}

/**
 * GoFeed
 */

declare namespace $goFeed {
    function parse(str: string): Record<string, any>
}

/**
 * Buffer
 */

declare class Buffer extends ArrayBuffer {
    static poolSize: number

    constructor(arg?: string | ArrayBuffer | ArrayLike<number>, encoding?: string)

    static from(arrayBuffer: ArrayBuffer): Buffer
    static from(array: ArrayLike<number>): Buffer
    static from(string: string, encoding?: string): Buffer

    static alloc(size: number, fill?: string | number, encoding?: string): Buffer

    equals(other: Buffer | Uint8Array): boolean

    toString(encoding?: string): string
}


/**
 * Crypto
 */

declare class WordArray {
    toString(encoder?: CryptoJSEncoder): string;
}

// CryptoJS supports AES-128, AES-192, and AES-256. It will pick the variant by the size of the key you pass in. If you use a passphrase,
// then it will generate a 256-bit key.
declare class CryptoJS {
    static AES: {
        encrypt: (message: string, key: string | Uint8Array, cfg?: AESConfig) => WordArray;
        decrypt: (message: string | WordArray, key: string | Uint8Array, cfg?: AESConfig) => WordArray;
    }
    static enc: {
        Utf8: CryptoJSEncoder;
        Base64: CryptoJSEncoder;
        Hex: CryptoJSEncoder;
        Latin1: CryptoJSEncoder;
        Utf16: CryptoJSEncoder;
        Utf16LE: CryptoJSEncoder;
    }
}

declare interface AESConfig {
    iv?: Uint8Array;
}

declare class CryptoJSEncoder {
    stringify(input: Uint8Array): string;

    parse(input: string): Uint8Array;
}


/**
 * Doc
 */

declare class DocSelection {
    // Retrieves the value of the specified attribute for the first element in the DocSelection.
    // To get the value for each element individually, use a looping construct such as each or map.
    attr(name: string): string | undefined;

    // Returns an object containing the attributes of the first element in the DocSelection.
    attrs(): { [key: string]: string };

    // Gets the child elements of each element in the DocSelection, optionally filtered by a selector.
    children(selector?: string): DocSelection;

    // For each element in the DocSelection, gets the first ancestor that matches the selector by testing the element itself
    // and traversing up through its ancestors in the DOM tree.
    closest(selector?: string): DocSelection;

    // Gets the children of each element in the DocSelection, including text and comment nodes.
    contents(): DocSelection;

    // Gets the children of each element in the DocSelection, filtered by the specified selector.
    contentsFiltered(selector: string): DocSelection;

    // Gets the value of a data attribute for the first element in the DocSelection.
    // If no name is provided, returns an object containing all data attributes.
    data<T extends string | undefined>(name?: T): T extends string ? (string | undefined) : { [key: string]: string };

    // Iterates over each element in the DocSelection, executing a function for each matched element.
    each(callback: (index: number, element: DocSelection) => void): DocSelection;

    // Ends the most recent filtering operation in the current chain and returns the set of matched elements to its previous state.
    end(): DocSelection;

    // Reduces the set of matched elements to the one at the specified index. If a negative index is given, it counts backwards starting at the end
    // of the set.
    eq(index: number): DocSelection;

    // Filters the set of matched elements to those that match the selector.
    filter(selector: string | ((index: number, element: DocSelection) => boolean)): DocSelection;

    // Gets the descendants of each element in the DocSelection, filtered by a selector.
    find(selector: string): DocSelection;

    // Reduces the set of matched elements to the first element in the DocSelection.
    first(): DocSelection;

    // Reduces the set of matched elements to those that have a descendant that matches the selector.
    has(selector: string): DocSelection;

    // Gets the combined text contents of each element in the DocSelection, including their descendants.
    text(): string;

    // Gets the HTML contents of the first element in the DocSelection.
    html(): string | null;

    // Checks the set of matched elements against a selector and returns true if at least one of these elements matches.
    is(selector: string | ((index: number, element: DocSelection) => boolean)): boolean;

    // Reduces the set of matched elements to the last element in the DocSelection.
    last(): DocSelection;

    // Gets the number of elements in the DocSelection.
    length(): number;

    // Passes each element in the current matched set through a function, producing an array of the return values.
    map<T>(callback: (index: number, element: DocSelection) => T): T[];

    // Gets the next sibling of each element in the DocSelection, optionally filtered by a selector.
    next(selector?: string): DocSelection;

    // Gets all following siblings of each element in the DocSelection, optionally filtered by a selector.
    nextAll(selector?: string): DocSelection;

    // Gets the next siblings of each element in the DocSelection, up to but not including the element matched by the selector.
    nextUntil(selector: string, until?: string): DocSelection;

    // Removes elements from the DocSelection that match the selector.
    not(selector: string | ((index: number, element: DocSelection) => boolean)): DocSelection;

    // Gets the parent of each element in the DocSelection, optionally filtered by a selector.
    parent(selector?: string): DocSelection;

    // Gets the ancestors of each element in the DocSelection, optionally filtered by a selector.
    parents(selector?: string): DocSelection;

    // Gets the ancestors of each element in the DocSelection, up to but not including the element matched by the selector.
    parentsUntil(selector: string, until?: string): DocSelection;

    // Gets the previous sibling of each element in the DocSelection, optionally filtered by a selector.
    prev(selector?: string): DocSelection;

    // Gets all preceding siblings of each element in the DocSelection, optionally filtered by a selector.
    prevAll(selector?: string): DocSelection;

    // Gets the previous siblings of each element in the DocSelection, up to but not including the element matched by the selector.
    prevUntil(selector: string, until?: string): DocSelection;

    // Gets the siblings of each element in the DocSelection, optionally filtered by a selector.
    siblings(selector?: string): DocSelection;
}

declare class Doc extends DocSelection {
    constructor(html: string);
}

declare function LoadDoc(html: string): DocSelectionFunction;

declare interface DocSelectionFunction {
    (selector: string): DocSelection;
}

/**
 * Torrent utils
 */

declare interface $torrentUtils {
    /**
     * Get a magnet link from a base64 encoded torrent data
     * @param b64 - The base64 encoded torrent data
     * @returns The magnet link
     */
    getMagnetLinkFromTorrentData(b64: string): string
}

/**
 * Media utils
 */

declare namespace $scannerUtils {
    interface NormalizedTitle {
        original: string
        normalized: string
        cleanBaseTitle: string
        denoisedTitle: string
        tokens: string[]
        season: number
        part: number
        year: number
        isMain: boolean
    }

    interface SmartSearchTitlesResult {
        /** Cleaned, deduplicated, search-ready title variants */
        titles: string[]
        /** Detected season number (-1 if none) */
        season: number
        /** Detected part number (-1 if none) */
        part: number
    }

    /**
     * Normalizes a title for matching. Handles macrons, possessives, separators, format words, etc.
     * @param title - The title to normalize
     * @returns A NormalizedTitle object with tokens, season, part, year extracted
     */
    function normalizeTitle(title: string): NormalizedTitle

    /**
     * Extracts a part number from a title string.
     * Handles: "Part 2", "Cour 2", "Part II", "2nd Part", "2nd Cour"
     * @param title - The title to extract from
     * @returns The part number, or -1 if not found
     */
    function extractPartNumber(title: string): number

    /**
     * Extracts a season number from a title string.
     * Handles: "Season 2", "S02", "2nd Season", roman numerals (II, III),
     * trailing numbers (Konosuba 2), Japanese patterns (2期)
     * @param title - The title to extract from
     * @returns The season number, or -1 if not found
     */
    function extractSeasonNumber(title: string): number

    /**
     * Extracts a year from a title string.
     * @param title - The title to extract from
     * @returns The year, or -1 if not found
     */
    function extractYear(title: string): number

    /**
     * Compares two titles using weighted token matching after normalization.
     * @param title1 - First title
     * @param title2 - Second title
     * @returns Match ratio (0.0 - 1.0)
     */
    function compareTitles(title1: string, title2: string): number

    /**
     * Finds the best matching title from a list of candidates.
     * @param target - The target title to match against
     * @param candidates - Array of candidate titles
     * @returns The best matching candidate string
     */
    function findBestMatch(target: string, candidates: string[]): string

    /**
     * Returns significant (non-noise) tokens from a title string.
     * @param title - The title to tokenize
     * @returns Array of significant tokens
     */
    function getSignificantTokens(title: string): string[]

    /**
     * Builds a clean search query from a title by normalizing and removing noise words.
     * @param title - The title to build a query from
     * @returns A compact search query string
     */
    function buildSearchQuery(title: string): string

    /**
     * Builds an advanced boolean query grouping multiple alternative titles with OR syntax.
     * @param titles - Array of alternative titles
     * @returns A query string like "(title1 | title2 | title3)"
     */
    function buildAdvancedQuery(titles: string[]): string

    /**
     * Strips special search syntax characters from a string.
     * @param query - The raw query string to sanitize
     * @returns The sanitized query string
     */
    function sanitizeQuery(query: string): string

    /**
     * Builds a query string with multiple season identifier formats.
     * e.g. buildSeasonQuery("Overlord", 2) → "(Overlord S02 | Overlord Season 2 | Overlord 2nd Season)"
     * @param title - The base title
     * @param season - The season number
     * @returns A query string with season alternatives, or just the base title for season 0/1
     */
    function buildSeasonQuery(title: string, season: number): string

    /**
     * Builds a query string with multiple part identifier formats.
     * e.g. buildPartQuery("Re:Zero", 2) → "(Re Zero Part 2 | Re Zero Part II | Re Zero 2nd Cour)"
     * @param title - The base title
     * @param part - The part number
     * @returns A query string with part alternatives, or just the base title for part 0/1
     */
    function buildPartQuery(title: string, part: number): string

    /**
     * Processes all media titles and returns cleaned, deduplicated search variants
     * along with extracted season and part numbers.
     *
     * Performs: normalization (macrons, possessives, separators), season/part/roman numeral
     * extraction, colon/dash splitting for shortened variants, deduplication.
     *
     * @param titles - Array of all media titles (romaji, english, synonyms)
     * @returns An object with titles (cleaned variants), season, and part
     */
    function buildSmartSearchTitles(titles: string[]): SmartSearchTitlesResult
}

/**
 * ChromeDP
 */

declare interface ChromeBrowserOptions {
    /** Timeout in seconds, defaults to 30 */
    timeout?: number;
    /** CSS selector to wait for after page load */
    waitSelector?: string;
    /** Milliseconds to wait after page load */
    waitDuration?: number;
    /** Custom user agent */
    userAgent?: string;
    /** Run in headless mode, defaults to true */
    headless?: boolean;
}

declare interface NewChromeBrowserOptions {
    /** Timeout in seconds, defaults to 30 */
    timeout?: number;
    /** Custom user agent */
    userAgent?: string;
    /** Run in headless mode, defaults to true */
    headless?: boolean;
}

declare interface ChromeBrowser {
    /**
     * Navigate to a URL
     * @param url - The URL to navigate to
     */
    navigate(url: string): Promise<void>;

    /**
     * Wait for a selector to be visible
     * @param selector - CSS selector
     */
    waitVisible(selector: string): Promise<void>;

    /**
     * Wait for a selector to be ready
     * @param selector - CSS selector
     */
    waitReady(selector: string): Promise<void>;

    /**
     * Click on an element
     * @param selector - CSS selector
     */
    click(selector: string): Promise<void>;

    /**
     * Type text into an element
     * @param selector - CSS selector
     * @param keys - Text to type
     */
    sendKeys(selector: string, keys: string): Promise<void>;

    /**
     * Evaluate JavaScript in the browser context
     * @param jsCode - JavaScript code to evaluate
     * @returns The result of the evaluation
     */
    evaluate(jsCode: string): Promise<any>;

    /**
     * Get the inner HTML of an element
     * @param selector - CSS selector
     * @returns The inner HTML
     */
    innerHTML(selector: string): Promise<string>;

    /**
     * Get the outer HTML of an element
     * @param selector - CSS selector
     * @returns The outer HTML
     */
    outerHTML(selector: string): Promise<string>;

    /**
     * Get the text content of an element
     * @param selector - CSS selector
     * @returns The text content
     */
    text(selector: string): Promise<string>;

    /**
     * Get an attribute value of an element
     * @param selector - CSS selector
     * @param attributeName - Name of the attribute
     * @returns The attribute value or null if not found
     */
    attribute(selector: string, attributeName: string): Promise<string | null>;

    /**
     * Capture a screenshot of a specific element
     * @param selector - CSS selector
     * @returns The screenshot as a byte array
     */
    screenshot(selector: string): Promise<Uint8Array>;

    /**
     * Capture a full page screenshot
     * @returns The screenshot as a byte array
     */
    fullScreenshot(): Promise<Uint8Array>;

    /**
     * Sleep for a duration
     * @param milliseconds - Duration in milliseconds
     */
    sleep(milliseconds: number): Promise<void>;

    /**
     * Close the browser instance
     */
    close(): Promise<void>;
}

declare class ChromeDP {
    /**
     * Create a new browser instance.
     * The default timeout is 30 seconds.
     * @param options - Browser options
     * @returns A browser instance
     */
    static newBrowser(options?: NewChromeBrowserOptions): Promise<ChromeBrowser>;

    /**
     * Navigate to a URL and return the HTML content
     * @param url - The URL to scrape
     * @param options - Scraping options
     * @returns The HTML content
     */
    static scrape(url: string, options?: ChromeBrowserOptions): Promise<string>;

    /**
     * Capture a screenshot of a webpage
     * @param url - The URL to screenshot
     * @param options - Screenshot options
     * @returns The screenshot as a byte array
     */
    static screenshot(url: string, options?: ChromeBrowserOptions): Promise<Uint8Array>;

    /**
     * Run JavaScript code in the browser context and return the result
     * @param url - The URL to navigate to
     * @param jsCode - JavaScript code to evaluate
     * @param options - Evaluation options
     * @returns The result of the evaluation
     */
    static evaluate(url: string, jsCode: string, options?: ChromeBrowserOptions): Promise<any>;
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

    /**
     * Watches a key in the store.
     * @param key - The key to watch
     * @param callback - The callback to call when the key changes
     */
    function watch<T = any>(key: string, callback: (value: T) => void): void
}
