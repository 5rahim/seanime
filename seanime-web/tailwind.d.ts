declare module "tailwindcss/lib/util/flattenColorPalette" {
    export default function flattenColorPalette(
        pallette: Record<string, string>,
    ): Record<string, string>;
}
