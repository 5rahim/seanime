import type { Config } from "tailwindcss"

const config: Config = {
    darkMode: "class",
    content: [
        "./index.html",
        "./src/app/**/*.{ts,tsx,mdx}",
        "./src/pages/**/*.{ts,tsx,mdx}",
        "./src/components/**/*.{ts,tsx,mdx}",
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
                    50: "#f2f0ff",
                    100: "#eeebff",
                    200: "#d4d0ff",
                    300: "#c7c2ff",
                    400: "#9f92ff",
                    500: "#6152df",
                    600: "#5243cb",
                    700: "#3f2eb2",
                    800: "#312887",
                    900: "#231c6b",
                    950: "#1a144f",
                    DEFAULT: "#6152df",
                },
                gray: {
                    50: "#FAFAFA",
                    100: "#F5F5F5",
                    200: "#E5E5E5",
                    300: "#D4D4D4",
                    400: "#A3A3A3",
                    500: "#737373",
                    600: "#525252",
                    700: "#404040",
                    800: "#262626",
                    900: "#171717",
                    950: "#101010",
                    DEFAULT: "#737373",
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
    plugins: [require("@tailwindcss/typography"), require("@tailwindcss/forms"), require("@headlessui/tailwindcss"),
        require("tailwind-scrollbar-hide"), require("tailwindcss-animate")],
}
export default config
