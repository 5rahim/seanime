import type { Config } from "tailwindcss"
// @ts-ignore
import flattenColorPalette from "tailwindcss/lib/util/flattenColorPalette"


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
        "bg-green-900", "bg-green-800", "bg-green-700", "bg-green-600", "bg-green-500", "bg-green-400", "bg-green-400", "bg-green-300",
        "text-green-300",
        "bg-gray-900", "bg-gray-800", "bg-gray-700", "bg-gray-600", "bg-gray-500", "bg-gray-400", "bg-gray-400", "bg-gray-300", "text-gray-300",
        "bg-indigo-900", "bg-indigo-800", "bg-indigo-700", "bg-indigo-600", "bg-indigo-500", "bg-indigo-400", "bg-indigo-400", "bg-indigo-300",
        "text-indigo-300",
        "bg-lime-900", "bg-lime-800", "bg-lime-700", "bg-lime-600", "bg-lime-500", "bg-lime-400", "bg-lime-400", "bg-lime-300", "text-lime-300",
        "bg-red-900", "bg-red-800", "bg-red-700", "bg-red-600", "bg-red-500", "bg-red-400", "bg-red-400", "bg-red-300", "text-red-300",
        "bg-emerald-900", "bg-emerald-800", "bg-emerald-700", "bg-emerald-600", "bg-emerald-500", "bg-emerald-400", "bg-emerald-400",
        "bg-emerald-300", "text-emerald-300",
        "bg-purple-900", "bg-purple-800", "bg-purple-700", "bg-purple-600", "bg-purple-500", "bg-purple-400", "bg-purple-400",
        "bg-opacity-70",
        "bg-opacity-80",
        "bg-opacity-70",
        "bg-opacity-60",
        "bg-opacity-50",
        "bg-opacity-30",
        "bg-opacity-20",
        "bg-opacity-10",
        "bg-opacity-5",
        "drop-shadow-sm",
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
            },
        },
        data: {
            checked: "checked",
            selected: "selected",
            disabled: "disabled",
            highlighted: "highlighted",
        },
        extend: {
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
