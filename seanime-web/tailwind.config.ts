import type { Config } from "tailwindcss"

const config: Config = {
    darkMode: "class",
    content: [
        "./index.html",
        "./src/app/**/*.{ts,tsx,mdx}",
        "./src/pages/**/*.{ts,tsx,mdx}",
        "./src/components/**/*.{ts,tsx,mdx}",
    ],
    safelist: [
        "bg-amber-900", "bg-amber-800", "bg-amber-700", "bg-amber-600", "bg-amber-500", "bg-amber-400", "bg-amber-400", "bg-amber-300",
        "text-amber-300",
        "text-amber-200",
        "bg-green-900", "bg-green-800", "bg-green-700", "bg-green-600", "bg-green-500", "bg-green-400", "bg-green-400", "bg-green-300",
        "text-green-300",
        "text-green-200",
        "bg-gray-900", "bg-gray-800", "bg-gray-700", "bg-gray-600", "bg-gray-500", "bg-gray-400", "bg-gray-400", "bg-gray-300", "text-gray-300",
        "bg-indigo-900", "bg-indigo-800", "bg-indigo-700", "bg-indigo-600", "bg-indigo-500", "bg-indigo-400", "bg-indigo-400", "bg-indigo-300",
        "text-indigo-300", "text-indigo-200",
        "bg-lime-900", "bg-lime-800", "bg-lime-700", "bg-lime-600", "bg-lime-500", "bg-lime-400", "bg-lime-400", "bg-lime-300", "text-lime-300",
        "text-lime-200",
        "text-lime-400",
        "text-lime-500",
        "bg-red-900", "bg-red-800", "bg-red-700", "bg-red-600", "bg-red-500", "bg-red-400", "bg-red-400", "bg-red-300", "text-red-300",
        "text-red-200",
        "bg-emerald-900", "bg-emerald-800", "bg-emerald-700", "bg-emerald-600", "bg-emerald-500", "bg-emerald-400", "bg-emerald-400",
        "bg-emerald-300", "text-emerald-300", "text-emerald-200", "text-emerald-400", "text-emerald-500",
        "bg-purple-900", "bg-purple-800", "bg-purple-700", "bg-purple-600", "bg-purple-500", "bg-purple-400", "bg-purple-400",
        "bg-green-300", "text-green-300", "text-green-200", "text-green-400", "text-green-500",
        "bg-opacity-70",
        "bg-opacity-80",
        "bg-opacity-70",
        "bg-opacity-60",
        "bg-opacity-50",
        "bg-opacity-30",
        "bg-opacity-20",
        "bg-opacity-10",
        "bg-opacity-5",
        "text-audienceScore-100", "text-audienceScore-200", "text-audienceScore-300", "text-audienceScore-400", "text-audienceScore-500",
        "text-audienceScore-600", "text-audienceScore-700", "text-audienceScore-800", "text-audienceScore-900",
        "drop-shadow-sm",
        "-top-10 top-10",

        {
            pattern: /bg-(red|green|blue|gray|brand|orange|yellow)-(100|200|300|400|500|600|700|800|900|950)/,
            variants: ["hover"],
        },
        // {
        //     pattern: /text-(red|green|blue|gray|brand|orange|yellow)-(100|200|300|400|500|600|700|800|900|950)/,
        //     variants: ['hover'],
        // },
        // {
        //     pattern: /border-(red|green|blue|gray|brand|orange|yellow)-(100|200|300|400|500|600|700|800|900|950)/,
        // },

        {
            pattern: /p-[0-9]+/,
        },
        {
            pattern: /m-[0-9]+/,
        },
        {
            pattern: /gap-[0-9]+/,
        },
        {
            pattern: /(px|py|pt|pb|pl|pr)-[0-9]+/,
        },
        {
            pattern: /(mx|my|mt|mb|ml|mr)-[0-9]+/,
        },

        {
            pattern: /grid-cols-[1-9]+/,
            variants: ["lg"],
        },
        {
            pattern: /col-span-[1-9]+/,
            variants: ["lg"],
        },
        "flex", "inline-flex", "grid", "inline-grid",
        "flex-row", "flex-col", "flex-row-reverse", "flex-col-reverse",
        "flex-wrap", "flex-nowrap", "flex-wrap-reverse",
        "items-start", "items-center", "items-end", "items-baseline", "items-stretch",
        "justify-start", "justify-center", "justify-end", "justify-between", "justify-around", "justify-evenly",

        {
            pattern: /flex|inline-flex|grid|inline-grid|flex-row|flex-col|flex-row-reverse|flex-col-reverse|flex-wrap|flex-nowrap|flex-wrap-reverse|items-start|items-center|items-end|items-baseline|items-stretch|justify-start|justify-center|justify-end|justify-between|justify-around|justify-evenly/,
            variants: ["lg", "md"],
        },


        // {
        //     pattern: /w-[0-9]+/,
        //     variants: ['lg', 'md', 'sm', 'xl', '2xl'],
        // },
        // {
        //     pattern: /h-[0-9]+/,
        //     variants: ['lg', 'md', 'sm', 'xl', '2xl'],
        // },
        "w-full", "h-full", "w-screen", "h-screen", "w-auto", "h-auto",
        "min-w-0", "min-h-0", "max-w-none", "max-h-none",

        {
            pattern: /text-xs|text-sm|text-base|text-lg|text-xl|text-2xl|text-3xl/,
            variants: ["lg", "md"],
        },

        {
            pattern: /font-thin|font-light|font-normal|font-medium|font-semibold|font-bold/,
            variants: ["lg", "md"],
        },

        {
            pattern: /text-left|text-center|text-right|text-justify/,
            variants: ["lg", "md"],
        },


        "uppercase", "lowercase", "capitalize", "normal-case",
        "truncate", "overflow-ellipsis", "overflow-clip",

        "rounded-none", "rounded-sm", "rounded", "rounded-md", "rounded-lg", "rounded-xl", "rounded-2xl", "rounded-3xl", "rounded-full",
        "border", "border-0", "border-2", "border-4", "border-8",
        // {
        //     pattern: /border-[0-9]+/,
        //     variants: ['lg', 'md', 'sm', 'xl', '2xl', 'hover', 'focus'],
        // },

        "shadow-sm", "shadow", "shadow-md", "shadow-lg", "shadow-xl", "shadow-2xl", "shadow-inner", "shadow-none",
        "opacity-0", "opacity-25", "opacity-50", "opacity-75", "opacity-100",


        // "transition", "transition-all", "transition-colors", "transition-opacity", "transition-shadow", "transition-transform",
        // "duration-75", "duration-100", "duration-150", "duration-200", "duration-300", "duration-500", "duration-700", "duration-1000",
        // "ease-linear", "ease-in", "ease-out", "ease-in-out",
        // "scale-0", "scale-50", "scale-75", "scale-90", "scale-95", "scale-100", "scale-105", "scale-110", "scale-125", "scale-150",
        // "rotate-0", "rotate-45", "rotate-90", "rotate-180", "-rotate-45", "-rotate-90", "-rotate-180",
        // "translate-x-0", "translate-x-1", "translate-x-2", "translate-x-4", "translate-x-8",
        // "translate-y-0", "translate-y-1", "translate-y-2", "translate-y-4", "translate-y-8",

        "cursor-pointer", "cursor-not-allowed", "cursor-wait", "cursor-text", "cursor-move", "cursor-help",
        "select-none", "select-text", "select-all", "select-auto",
        "pointer-events-none", "pointer-events-auto",
        "resize", "resize-none", "resize-y", "resize-x",

        "static", "fixed", "absolute", "relative", "sticky",
        "top-0", "right-0", "bottom-0", "left-0",
        "z-0", "z-10", "z-20", "z-30", "z-40", "z-50", "z-auto",

        "block", "inline-block", "inline", "hidden",
        {
            pattern: /hidden|block|inline|inline-block/,
            variants: ["lg", "md"],
        },

        "overflow-auto", "overflow-hidden", "overflow-visible", "overflow-scroll",
        "overflow-x-auto", "overflow-y-auto", "overflow-x-hidden", "overflow-y-hidden",
    ],
    theme: {
        container: {
            center: true,
            padding: {
                DEFAULT: "1rem",
                sm: "2rem",
                lg: "4rem",
                xl: "5rem",
                "2xl": "6rem",
            },
            screens: {
                "2xl": "1400px",
                "3xl": "1600px",
                "4xl": "1800px",
                "5xl": "2000px",
                "6xl": "2200px",
                "7xl": "2400px",
            },
        },
        data: {
            checked: "checked",
            selected: "selected",
            disabled: "disabled",
            highlighted: "highlighted",
        },
        extend: {
            screens: {
                "3xl": "1600px",
                "4xl": "1800px",
                "5xl": "2000px",
                "6xl": "2200px",
                "7xl": "2400px",
            },
            animationDuration: {
                DEFAULT: "0.25s",
            },
            keyframes: {
                "accordion-down": {
                    from: { height: "0" },
                    to: { height: "var(--radix-accordion-content-height)" },
                },
                "accordion-up": {
                    from: { height: "var(--radix-accordion-content-height)" },
                    to: { height: "0" },
                },
                "slide-down": {
                    from: { transform: "translateY(-1rem)", opacity: "0" },
                    to: { transform: "translateY(0)", opacity: "1" },
                },
                "slide-up": {
                    from: { transform: "translateY(0)", opacity: "1" },
                    to: { transform: "translateY(-1rem)", opacity: "0" },
                },
                "indeterminate-progress": {
                    "0%": { transform: " translateX(0) scaleX(0)" },
                    "40%": { transform: "translateX(0) scaleX(0.4)" },
                    "100%": { transform: "translateX(100%) scaleX(0.5)" },
                },
            },
            animation: {
                "accordion-down": "accordion-down 0.15s linear",
                "accordion-up": "accordion-up 0.15s linear",
                "slide-down": "slide-down 0.15s ease-in-out",
                "slide-up": "slide-up 0.15s ease-in-out",
                "indeterminate-progress": "indeterminate-progress 1s infinite ease-out",
            },
            transformOrigin: {
                "left-right": "0% 100%",
            },
            boxShadow: {
                "md": "0 1px 3px 0 rgba(0, 0, 0, 0.1),0 1px 2px 0 rgba(0, 0, 0, 0.06)",
            },
            colors: {
                brand: {
                    50: "rgb(var(--color-brand-50) / <alpha-value>)",
                    100: "rgb(var(--color-brand-100) / <alpha-value>)",
                    200: "rgb(var(--color-brand-200) / <alpha-value>)",
                    300: "rgb(var(--color-brand-300) / <alpha-value>)",
                    400: "rgb(var(--color-brand-400) / <alpha-value>)",
                    500: "rgb(var(--color-brand-500) / <alpha-value>)",
                    600: "rgb(var(--color-brand-600) / <alpha-value>)",
                    700: "rgb(var(--color-brand-700) / <alpha-value>)",
                    800: "rgb(var(--color-brand-800) / <alpha-value>)",
                    900: "rgb(var(--color-brand-900) / <alpha-value>)",
                    950: "rgb(var(--color-brand-950) / <alpha-value>)",
                    DEFAULT: "rgb(var(--color-brand-500) / <alpha-value>)",
                },
                gray: {
                    50: "rgb(var(--color-gray-50) / <alpha-value>)",
                    100: "rgb(var(--color-gray-100) / <alpha-value>)",
                    200: "rgb(var(--color-gray-200) / <alpha-value>)",
                    300: "rgb(var(--color-gray-300) / <alpha-value>)",
                    400: "rgb(var(--color-gray-400) / <alpha-value>)",
                    500: "rgb(var(--color-gray-500) / <alpha-value>)",
                    600: "rgb(var(--color-gray-600) / <alpha-value>)",
                    700: "rgb(var(--color-gray-700) / <alpha-value>)",
                    800: "rgb(var(--color-gray-800) / <alpha-value>)",
                    900: "rgb(var(--color-gray-900) / <alpha-value>)",
                    950: "rgb(var(--color-gray-950) / <alpha-value>)",
                    DEFAULT: "rgb(var(--color-gray-500) / <alpha-value>)",
                },
                green: {
                    50: "#e6f7ea",
                    100: "#cfead6",
                    200: "#7bd0a7",
                    300: "#68b695",
                    400: "#57a181",
                    500: "#258c60",
                    600: "#1a6444",
                    700: "#154f37",
                    800: "#103b29",
                    900: "#0a2318",
                    950: "#05130d",
                    DEFAULT: "#258c60",
                },
                audienceScore: {
                    300: "#b45d5d",
                    500: "#9d8741",
                    600: "#a0b974",
                    700: "#57a181",
                },
                background: {
                    500: "rgb(var(--background) / <alpha-value>)",
                    DEFAULT: "rgb(var(--background) / <alpha-value>)",
                },
            },
        },
    },
    plugins: [
        require("@tailwindcss/typography"),
        require("@tailwindcss/forms"),
        require("@headlessui/tailwindcss"),
        require("tailwind-scrollbar-hide"),
        require("tailwindcss-animate"),
        addVariablesForColors,
        function ({ addVariant }: { addVariant: (variant: string, selector: string) => void }) {
            addVariant("firefox", ":-moz-any(&)")
        },
    ],
}
export default config


function addVariablesForColors({ addBase, theme }: any) {
    let allColors = flattenColorPalette(theme("colors"))
    let newVars = Object.fromEntries(
        Object.entries(allColors).map(([key, val]) => [`--${key}`, val]),
    )

    addBase({
        ":root": newVars,
    })
}

type Colors = {
    [key: string | number]: string | Colors
}

function flattenColorPalette(colors: Colors) {
    let result: Record<string, string> = {}

    for (let [root, children] of Object.entries(colors ?? {})) {
        if (root === "__CSS_VALUES__") continue
        if (typeof children === "object" && children !== null) {
            for (let [parent, value] of Object.entries(flattenColorPalette(children))) {
                result[`${root}${parent === "DEFAULT" ? "" : `-${parent}`}`] = value
            }
        } else {
            result[root] = children
        }
    }

    if ("__CSS_VALUES__" in colors) {
        for (let [key, value] of Object.entries(colors.__CSS_VALUES__)) {
            if ((Number(value) & 1 << 2) === 0) {
                result[key] = colors[key] as string
            }
        }
    }

    return result
}
