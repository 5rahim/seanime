declare class Buffer extends Uint8Array {
    static poolSize: number

    constructor(arg?: string | ArrayBuffer | ArrayLike<number>, encoding?: string);

    static from(arrayBuffer: ArrayBuffer): Buffer;
    static from(array: ArrayLike<number>): Buffer;
    static from(string: string, encoding?: string): Buffer;

    static alloc(size: number, fill?: string | number, encoding?: string): Buffer;

    equals(other: Buffer | Uint8Array): boolean;

    toString(encoding?: string): string;
}
