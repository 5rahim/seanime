# Seanime-Mihon Connector

**Normalized intent:** Create a Mihon-compatible extension that connects to the user's Seanime Docker instance and serves its downloaded manga library to the Mihon Android app — the same pattern Suwayomi uses (server exposes API, Tachiyomi/Mihon extension consumes it).

**Complexity:** Standard (two codebases: Go server endpoints + Kotlin Android extension)
**Risk:** Low (additive — new API endpoints, no changes to existing code)

---

## Architecture

```
┌─────────────┐     HTTPS/HTTP      ┌────────────────────┐
│  Mihon App   │ ◄─────────────────► │  Seanime Server    │
│  (Android)   │                     │  (Docker container) │
│              │   GET /mihon/...    │                    │
│  ┌─────────┐ │                     │  ┌──────────────┐  │
│  │Seanime  │ │  JSON responses    │  │ Mihon API    │  │
│  │Extension│ │ ◄────────────────► │  │ Handler      │  │
│  │(APK)    │ │                     │  │ (new Go code)│  │
│  └─────────┘ │   Image files      │  └──────┬───────┘  │
│              │ ◄────────────────► │         │          │
└─────────────┘  /manga-downloads/  │  ┌──────▼───────┐  │
                                    │  │ Download Dir │  │
                                    │  │ /config/manga│  │
                                    │  └──────────────┘  │
                                    └────────────────────┘
```

### Why this approach (vs alternatives)

| Approach | Pros | Cons | Verdict |
|----------|------|------|---------|
| **Custom REST API + HttpSource extension** | Simple, works with existing Seanime data, no external deps | Need to build extension APK | **Chosen** |
| Add Komga-compatible API to Seanime | Mihon already has Komga extension | Complex spec to implement, different data model | Rejected |
| Add OPDS feed to Seanime | Standard protocol | Dated XML format, poor image serving, Mihon OPDS support limited | Rejected |
| GraphQL like Suwayomi | Proven pattern | Massive overkill — Apollo codegen, schema design | Rejected |

---

## Part 1: Seanime Server — Mihon API Endpoints (Go)

### New file: `internal/handlers/mihon.go`

Add lightweight GET endpoints under `/api/v1/mihon/` that scan the download directory and return JSON. These endpoints are **read-only wrappers** over the existing download directory — no DB writes, no state changes.

#### Endpoints

| Method | Path | Returns | Source |
|--------|------|---------|--------|
| `GET` | `/api/v1/mihon/library` | List of manga with cover URLs, titles (from AniList cache or download dir scan) | Scan download dir, group by `{provider}_{mediaId}`, query AniList for metadata |
| `GET` | `/api/v1/mihon/manga/:id` | Single manga details (title, author, description, cover, status) | AniList API or cached DB data |
| `GET` | `/api/v1/mihon/manga/:id/chapters` | Array of chapters with number, title, date | Scan download dir for `*_{mediaId}_*` directories, parse chapter numbers |
| `GET` | `/api/v1/mihon/chapter/:dir/pages` | Array of page objects with image URLs | Read `registry.json` from the chapter directory |

#### Data flow

1. **Library listing**: Scan `/config/manga/` directory → extract unique `{provider}_{mediaId}` pairs → for each mediaId, fetch AniList manga metadata (title, cover, description, author, status) via Seanime's existing AniList client → return JSON array of manga objects
2. **Chapter listing**: Filter download directories by mediaId → parse chapter numbers from dir names → sort by chapter number → return JSON array
3. **Page listing**: Read `registry.json` from chapter dir → build image URLs as `/manga-downloads/{dirName}/{filename}` → return ordered array

#### JSON response shapes

```json
// GET /api/v1/mihon/library
[{
  "id": 163824,
  "title": "Revenge of the Iron-Blooded Sword Hound",
  "author": "Author Name",
  "artist": "Artist Name",
  "description": "Synopsis text...",
  "cover_url": "https://s4.anilist.co/file/anilistcdn/media/manga/cover/large/...",
  "status": 1,
  "chapter_count": 120
}]

// GET /api/v1/mihon/manga/:id/chapters
[{
  "dir": "asurascans_163824_..._100_100",
  "number": 100.0,
  "title": "Chapter 100",
  "page_count": 16,
  "date": 1713600000
}]

// GET /api/v1/mihon/chapter/:dir/pages
[{
  "index": 0,
  "url": "/manga-downloads/asurascans_163824_.../01.webp"
}]
```

#### Authentication

The existing `/manga-downloads/` static route is served **outside auth middleware** (raw Echo static mount). The new `/api/v1/mihon/` endpoints will use `OptionalAuthMiddleware` — if a password is set, the extension must send `X-Seanime-Token` header.

#### Route registration

In `internal/handlers/routes.go`, add a new group:

```go
mihonGroup := e.Group("/api/v1/mihon")
mihonGroup.Use(h.OptionalAuthMiddleware)
mihonGroup.GET("/library", h.HandleMihonLibrary)
mihonGroup.GET("/manga/:id", h.HandleMihonMangaDetails)
mihonGroup.GET("/manga/:id/chapters", h.HandleMihonMangaChapters)
mihonGroup.GET("/chapter/:dir/pages", h.HandleMihonChapterPages)
```

### New file: `internal/handlers/mihon_repo.go`

Serve the extension repo directly from Seanime:

| Method | Path | Returns |
|--------|------|---------|
| `GET` | `/api/v1/mihon/repo/index.min.json` | Mihon extension repo manifest |
| `GET` | `/api/v1/mihon/repo/apk/:name` | Extension APK file |

This lets users add `https://seanime.example.com/api/v1/mihon/repo` as a Mihon extension repo URL.

---

## Part 2: Mihon Extension (Kotlin)

### New project: `mihon-extension/` (in seanime repo root)

Based on the Suwayomi/Keiyoushi extension build infrastructure.

#### Project structure

```
mihon-extension/
├── buildSrc/
│   ├── build.gradle.kts                # kotlin-dsl plugin
│   └── src/main/kotlin/AndroidConfig.kt
├── gradle/
│   ├── libs.versions.toml              # version catalog
│   └── wrapper/
├── core/
│   ├── build.gradle.kts                # android library (manifest + icon)
│   ├── AndroidManifest.xml             # tachiyomi.extension meta-data
│   └── res/mipmap-*/ic_launcher.png    # Seanime icon
├── src/all/seanime/
│   ├── build.gradle                    # extension metadata
│   ├── AndroidManifest.xml             # empty
│   └── src/eu/kanade/tachiyomi/extension/all/seanime/
│       ├── Seanime.kt                  # Main HttpSource implementation
│       ├── SeanimeApi.kt               # API response DTOs
│       └── SeanimeSettings.kt          # ConfigurableSource preferences
├── build.gradle.kts                    # root buildscript
├── settings.gradle.kts
├── common.gradle
└── gradle.properties
```

#### Extension class: `Seanime.kt`

Extends `HttpSource` + implements `ConfigurableSource`:

```kotlin
class Seanime : HttpSource(), ConfigurableSource {
    override val name = "Seanime"
    override val lang = "all"
    override val supportsLatest = false
    override val baseUrl: String  // from SharedPreferences

    // Library = popular manga (Mihon's browse tab)
    override fun popularMangaRequest(page: Int) =
        GET("$baseUrl/api/v1/mihon/library")

    override fun popularMangaParse(response: Response): MangasPage {
        // Parse JSON array → List<SManga>
    }

    // Search filters by title substring
    override fun searchMangaRequest(page: Int, query: String, filters: FilterList) =
        GET("$baseUrl/api/v1/mihon/library?q=$query")

    // Manga details
    override fun mangaDetailsRequest(manga: SManga) =
        GET("$baseUrl/api/v1/mihon/manga/${manga.url}")

    override fun mangaDetailsParse(response: Response): SManga { ... }

    // Chapters
    override fun chapterListRequest(manga: SManga) =
        GET("$baseUrl/api/v1/mihon/manga/${manga.url}/chapters")

    override fun chapterListParse(response: Response): List<SChapter> { ... }

    // Pages
    override fun pageListRequest(chapter: SChapter) =
        GET("$baseUrl/api/v1/mihon/chapter/${chapter.url}/pages")

    override fun pageListParse(response: Response): List<Page> { ... }

    override fun imageUrlParse(response: Response) = "" // not needed, direct URLs

    // ConfigurableSource — server URL setting
    override fun setupPreferenceScreen(screen: PreferenceScreen) {
        // EditTextPreference for server URL
        // EditTextPreference for auth token (optional)
    }
}
```

#### Key design decisions

1. **`lang = "all"`** — Seanime manga aren't language-specific
2. **`supportsLatest = false`** — no "latest updates" concept for a personal library
3. **Popular = library listing** — Mihon's "Browse > Popular" tab shows the full Seanime library
4. **Search = client-side filter** — query parameter filters manga titles server-side
5. **Auth token in preferences** — if Seanime has a password, user enters the token in extension settings
6. **Image URLs are absolute** — pointing to `/manga-downloads/...` on the server (no auth needed)

---

## Part 3: Extension Distribution

### Option A (chosen): Serve from Seanime itself

The Seanime server hosts the extension repo at `/api/v1/mihon/repo/`:
- `index.min.json` — repo manifest pointing to the APK
- `apk/tachiyomi-all.seanime-v1.4.1.apk` — the extension file

**User setup in Mihon:**
1. Open Mihon → Browse → Extensions → Repos
2. Add repo: `https://seanime.nobara.mg-dev25.com/api/v1/mihon/repo`
3. Extension "Seanime" appears in the extension list
4. Install → configure server URL + optional auth token
5. Browse → Seanime → see your manga library

### Option B (backup): GitHub Pages repo

If self-hosting the APK from Seanime is impractical, create a GitHub repo that hosts the `index.min.json` + APK.

---

## Implementation order

1. **Go endpoints** (`internal/handlers/mihon.go`, `internal/handlers/mihon_repo.go`, route registration)
2. **Kotlin extension** (full project in `mihon-extension/`)
3. **Build APK** (Gradle assembleRelease)
4. **Embed APK** in Seanime's repo-serving endpoint
5. **Test E2E** — install extension in Mihon emulator or real device via browser automation

---

## Risks and mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| AniList rate limiting on library endpoint | Slow first load | Cache manga metadata in memory after first fetch; AniList data rarely changes |
| Android SDK not available for building extension | Can't produce APK | Use Android SDK command-line tools (sdkmanager), available via package manager |
| Extension signing | Mihon may reject unsigned APK | Generate a debug keystore for signing; sideloaded extensions don't need official signing |
| Download dir naming with URL-encoded chars | Parsing complexity | Well-defined format: `{provider}_{mediaId}_{chapterId}_{chapterNumber}` — regex extraction |

---

## Out of scope

- Real-time sync (Syncyomi-style bidirectional progress sync)
- Writing back to Seanime from Mihon (reading progress, etc.)
- Streaming non-downloaded manga (only downloaded chapters are served)
- Publishing to the official Keiyoushi extension repo
