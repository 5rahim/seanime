/// <reference path="./manga-provider.d.ts" />

class Provider {

    private api = "https://api.comick.fun"

    getSettings(): Settings {
        return {
            supportsMultiLanguage: true,
            supportsMultiScanlator: false,
        }
    }

    async search(opts: QueryOptions): Promise<SearchResult[]> {
        console.log(this.api, opts.query)

        const requestRes = await fetch(`${this.api}/v1.0/search?q=${encodeURIComponent(opts.query)}&limit=25&page=1`, {
            method: "get",
        })
        const comickRes = await requestRes.json() as ComickSearchResult[]

        const ret: SearchResult[] = []

        for (const res of comickRes) {

            let cover: any = res.md_covers ? res.md_covers[0] : null
            if (cover && cover.b2key != undefined) {
                cover = "https://meo.comick.pictures/" + cover.b2key
            }

            ret.push({
                id: res.hid,
                title: res.title ?? res.slug,
                synonyms: res.md_titles?.map(t => t.title) ?? {},
                year: res.year ?? 0,
                image: cover,
            })
        }

        console.log(ret[0])

        console.error("test", ret[0].id)

        return ret
    }

    async findChapters(id: string): Promise<ChapterDetails[]> {

        console.log("Fetching chapters", id)

        const chapterList: ChapterDetails[] = []

        const data = (await (await fetch(`${this.api}/comic/${id}/chapters?lang=en&page=0&limit=1000000`))?.json()) as { chapters: ComickChapter[] }

        const chapters: ChapterDetails[] = []

        for (const chapter of data.chapters) {

            if (!chapter.chap) {
                continue
            }

            let title = "Chapter " + this.padNum(chapter.chap, 2) + " "

            if (title.length === 0) {
                if (!chapter.title) {
                    title = "Oneshot"
                } else {
                    title = chapter.title
                }
            }

            let canPush = true
            for (let i = 0; i < chapters.length; i++) {
                if (chapters[i].title?.trim() === title?.trim()) {
                    canPush = false
                }
            }

            if (canPush) {
                if (chapter.lang === "en") {
                    chapters.push({
                        url: `${this.api}/comic/${id}/chapter/${chapter.hid}`,
                        index: 0,
                        id: chapter.hid,
                        title: title?.trim(),
                        chapter: chapter.chap,
                        rating: chapter.up_count - chapter.down_count,
                        updatedAt: chapter.updated_at,
                    })
                }
            }
        }

        chapters.reverse()

        for (let i = 0; i < chapters.length; i++) {
            chapters[i].index = i
        }

        console.log(chapters.map(c => c.chapter))

        return chapters
    }

    async findChapterPages(id: string): Promise<ChapterPage[]> {

        const data = (await (await fetch(`${this.api}/chapter/${id}`))?.json()) as {
            chapter: { md_images: { vol: any; w: number; h: number; b2key: string }[] }
        }

        const pages: ChapterPage[] = []

        data.chapter.md_images.map((image, index: number) => {
            pages.push({
                url: `https://meo.comick.pictures/${image.b2key}?width=${image.w}`,
                index: index,
                headers: {},
            })
        })

        return pages
    }

    padNum(number: string, places: number): string {
        let range = number.split("-")
        range = range.map((chapter) => {
            chapter = chapter.trim()
            const digits = chapter.split(".")[0].length
            return "0".repeat(Math.max(0, places - digits)) + chapter
        })
        return range.join("-")
    }

}

interface ComickSearchResult {
    title: string;
    id: number;
    hid: string;
    slug: string;
    year?: number;
    rating: string;
    rating_count: number;
    follow_count: number;
    user_follow_count: number;
    content_rating: string;
    created_at: string;
    demographic: number;
    md_titles: { title: string }[];
    md_covers: { vol: any; w: number; h: number; b2key: string }[];
    highlight: string;
}

interface Comic {
    id: number;
    hid: string;
    title: string;
    country: string;
    status: number;
    links: {
        al: string;
        ap: string;
        bw: string;
        kt: string;
        mu: string;
        amz: string;
        cdj: string;
        ebj: string;
        mal: string;
        raw: string;
    };
    last_chapter: any;
    chapter_count: number;
    demographic: number;
    hentai: boolean;
    user_follow_count: number;
    follow_rank: number;
    comment_count: number;
    follow_count: number;
    desc: string;
    parsed: string;
    slug: string;
    mismatch: any;
    year: number;
    bayesian_rating: any;
    rating_count: number;
    content_rating: string;
    translation_completed: boolean;
    relate_from: Array<any>;
    mies: any;
    md_titles: { title: string }[];
    md_comic_md_genres: { md_genres: { name: string; type: string | null; slug: string; group: string } }[];
    mu_comics: {
        licensed_in_english: any;
        mu_comic_categories: {
            mu_categories: { title: string; slug: string };
            positive_vote: number;
            negative_vote: number;
        }[];
    };
    md_covers: { vol: any; w: number; h: number; b2key: string }[];
    iso639_1: string;
    lang_name: string;
    lang_native: string;
}

interface ComickChapter {
    id: number;
    chap: string;
    title: string;
    vol: string | null;
    lang: string;
    created_at: string;
    updated_at: string;
    up_count: number;
    down_count: number;
    group_name: any;
    hid: string;
    identities: any;
    md_chapter_groups: { md_groups: { title: string; slug: string } }[];
}
