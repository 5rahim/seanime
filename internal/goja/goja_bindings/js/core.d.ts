/**
 * Buffer
 */

declare class Buffer extends ArrayBuffer {
    static poolSize: number

    constructor(arg?: string | ArrayBuffer | ArrayLike<number>, encoding?: string);

    static from(arrayBuffer: ArrayBuffer): Buffer;
    static from(array: ArrayLike<number>): Buffer;
    static from(string: string, encoding?: string): Buffer;

    static alloc(size: number, fill?: string | number, encoding?: string): Buffer;

    equals(other: Buffer | Uint8Array): boolean;

    toString(encoding?: string): string;
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
 * Torrent
 */

declare function getMagnetLinkFromTorrentData(b64: string): string
