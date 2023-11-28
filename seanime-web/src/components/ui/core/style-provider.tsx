"use client"
import React from "react"

/* -------------------------------------------------------------------------------------------------
 * Locale
 * -----------------------------------------------------------------------------------------------*/

/**
 * @internal UI Folder scope
 */
type Lng = "fr" | "en" // DEVNOTE Add new lang keywords to maintain type safety
export type UILocaleConfig = {
    locale: Lng,
    countryLocale: string,
    country: string
}
const __LocaleConfigDefaultValue: UILocaleConfig = { locale: "en", countryLocale: "en-US", country: "us" }
const __LocaleConfigContext = React.createContext<UILocaleConfig>(__LocaleConfigDefaultValue)

/**
 * @internal UI Folder scope
 */
export const useUILocaleConfig = () => {
    return React.useContext(__LocaleConfigContext)
}

useUILocaleConfig.displayName = "useUILocaleConfig"

/* -------------------------------------------------------------------------------------------------
 * UI Provider
 * -----------------------------------------------------------------------------------------------*/

export interface UIProviderProps {
    children?: React.ReactNode
    config?: {
        locale?: Lng,
        countryLocale?: string,
        country?: string
    },
}

/**
 * @example
 * <UIProvider config={{ locale: 'en', countryLocale: 'en-US', country: 'us' }}>
 *    <App/>
 * </UIProvider>
 * @param children
 * @param config
 * @constructor
 */
export const UIProvider: React.FC<UIProviderProps> = ({ children, config }) => {

    const localeConfig: UILocaleConfig = {
        ...__LocaleConfigDefaultValue,
        ...config,
    }

    return (
        <__LocaleConfigContext.Provider value={localeConfig}>
            {children}
        </__LocaleConfigContext.Provider>
    )
}

UIProvider.displayName = "UIProvider"
