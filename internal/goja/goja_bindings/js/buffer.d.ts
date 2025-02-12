declare class Buffer extends SUint8Array {
    static poolSize: number

    constructor(arg?: string | ArrayBuffer | ArrayLike<number>, encoding?: string);

    static from(arrayBuffer: ArrayBuffer): Buffer;
    static from(array: ArrayLike<number>): Buffer;
    static from(string: string, encoding?: string): Buffer;

    static alloc(size: number, fill?: string | number, encoding?: string): Buffer;

    equals(other: Buffer | Uint8Array): boolean;

    toString(encoding?: string): string;
}

declare class SUint8Array {
    length: number

    constructor(arrayBuffer: ArrayBuffer);

    constructor(length: number);

    [index: number]: number;

    static of(...items: number[]): SUint8Array;

    static from(arrayLike: ArrayLike<number>): SUint8Array;
    static from(arrayBuffer: ArrayBuffer): SUint8Array;

    set(array: SUint8Array, offset?: number): void;
}
