type LanguageSource = {
    label?: string;
    language?: string;
};

const LANG_MAP: Record<string, string> = {
    en: "en", eng: "en", english: "en",
    ru: "ru", rus: "ru", russian: "ru",
    fr: "fr", fre: "fr", fra: "fr", french: "fr",
    de: "de", ger: "de", deu: "de", german: "de",
    es: "es", spa: "es", spanish: "es",
    it: "it", ita: "it", italian: "it",
    pt: "pt", por: "pt", portuguese: "pt",
    nl: "nl", dut: "nl", nld: "nl", dutch: "nl",
    pl: "pl", pol: "pl", polish: "pl",
    uk: "uk", ukr: "uk", ukrainian: "uk",
    ja: "ja", jpn: "ja", japanese: "ja",
    zh: "zh", chi: "zh", zho: "zh", chinese: "zh",
    ko: "ko", kor: "ko", korean: "ko",
    ar: "ar", ara: "ar", arabic: "ar",
    hi: "hi", hin: "hi", hindi: "hi",
    tr: "tr", tur: "tr", turkish: "tr",
    vi: "vi", vie: "vi", vietnamese: "vi",
    sv: "sv", swe: "sv", swedish: "sv",
    da: "da", dan: "da", danish: "da",
    fi: "fi", fin: "fi", finnish: "fi",
    no: "no", nor: "no", norwegian: "no",
    el: "el", gre: "el", greek: "el",
    he: "he", heb: "he", hebrew: "he",
    id: "id", ind: "id", indonesian: "id",
    th: "th", tha: "th", thai: "th",

    deutsch: "de",
    español: "es", castellano: "es",
    français: "fr",
    italiano: "it",
    português: "pt", brazilian: "pt",
    polski: "pl",
    nederlands: "nl",
    dansk: "da",
    suomi: "fi",
    norsk: "no",
    svenska: "sv",
}

const SCRIPT_CHECKS = [
    { code: "ko", re: /[\uAC00-\uD7AF]/ },        // Hangul
    { code: "ja", re: /[\u3040-\u30FF]/ },        // Hiragana/Katakana
    { code: "zh", re: /[\u4E00-\u9FFF]/ },        // Han (Chinese)
    { code: "uk", re: /[ЄІЇҐєіїґ]/ },             // Specific Ukrainian Cyrillic
    { code: "ru", re: /[\u0400-\u04FF]/ },        // General Cyrillic
    { code: "th", re: /[\u0E00-\u0E7F]/ },        // Thai
    { code: "ar", re: /[\u0600-\u06FF]/ },        // Arabic
    { code: "he", re: /[\u0590-\u05FF]/ },        // Hebrew
    { code: "el", re: /[\u0370-\u03FF]/ },        // Greek
    { code: "hi", re: /[\u0900-\u097F]/ },        // Devanagari
    { code: "bn", re: /[\u0980-\u09FF]/ },        // Bengali
    { code: "vi", re: /[àáạảãâầấậẩẫăằắặẳẵèéẹẻẽêềếệểễìíịỉĩòóọỏõôồốộổỗơờớợởỡùúụủũưừứựửữỳýỵỷỹđ]/ }, // Viet specific
]

const HAS_NON_ASCII = /[^\u0000-\u007f]/

const TOKENIZER = /[a-z0-9\u00C0-\u00FF]+/g

export function detectTrackLanguage(source: LanguageSource): string | null {
    if (source.language) {
        const cleanLang = source.language.trim().toLowerCase()
        if (LANG_MAP[cleanLang]) return LANG_MAP[cleanLang]
        if (cleanLang.length === 2) return cleanLang // Fallback for valid ISO codes
    }

    if (!source.label) return null
    const label = source.label

    if (HAS_NON_ASCII.test(label)) {
        for (let i = 0; i < SCRIPT_CHECKS.length; i++) {
            if (SCRIPT_CHECKS[i].re.test(label)) {
                return SCRIPT_CHECKS[i].code
            }
        }
    }

    const lowerLabel = label.toLowerCase()
    const tokens = lowerLabel.match(TOKENIZER)

    if (tokens) {
        for (let i = 0; i < tokens.length; i++) {
            const match = LANG_MAP[tokens[i]]
            if (match) return match
        }
    }

    return null
}
