package eu.kanade.tachiyomi.extension.all.seanime

import android.app.Application
import android.content.SharedPreferences
import androidx.preference.EditTextPreference
import androidx.preference.PreferenceScreen
import eu.kanade.tachiyomi.network.GET
import eu.kanade.tachiyomi.source.ConfigurableSource
import eu.kanade.tachiyomi.source.model.FilterList
import eu.kanade.tachiyomi.source.model.MangasPage
import eu.kanade.tachiyomi.source.model.Page
import eu.kanade.tachiyomi.source.model.SChapter
import eu.kanade.tachiyomi.source.model.SManga
import eu.kanade.tachiyomi.source.online.HttpSource
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.float
import kotlinx.serialization.json.int
import kotlinx.serialization.json.jsonArray
import kotlinx.serialization.json.jsonObject
import kotlinx.serialization.json.jsonPrimitive
import okhttp3.Headers
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.Response
import uy.kohesive.injekt.Injekt
import uy.kohesive.injekt.api.get

class Seanime : HttpSource(), ConfigurableSource {

    override val name = "Seanime"
    override val lang = "all"
    override val supportsLatest = false

    private val preferences: SharedPreferences by lazy {
        Injekt.get<Application>().getSharedPreferences("source_$id", 0x0000)
    }

    override val baseUrl: String
        get() = preferences.getString(PREF_SERVER_URL, DEFAULT_SERVER_URL)!!.trimEnd('/')

    override val client: OkHttpClient = network.cloudflareClient

    private val json = Json { ignoreUnknownKeys = true }

    private val authToken: String
        get() = preferences.getString(PREF_AUTH_TOKEN, "")!!

    override fun headersBuilder(): Headers.Builder = super.headersBuilder().apply {
        val token = authToken
        if (token.isNotBlank()) {
            add("X-Seanime-Token", token)
        }
    }

    // -- Popular (= full library) --

    override fun popularMangaRequest(page: Int): Request =
        GET("$baseUrl/api/v1/mihon/library", headers)

    override fun popularMangaParse(response: Response): MangasPage {
        val wrapper = json.parseToJsonElement(response.body.string()).jsonObject
        val data = wrapper["data"]?.jsonArray ?: return MangasPage(emptyList(), false)
        val mangas = data.map { parseManga(it.jsonObject) }
        return MangasPage(mangas, false)
    }

    // -- Search --

    override fun searchMangaRequest(page: Int, query: String, filters: FilterList): Request =
        GET("$baseUrl/api/v1/mihon/library?q=$query", headers)

    override fun searchMangaParse(response: Response): MangasPage = popularMangaParse(response)

    // -- Latest (not supported) --

    override fun latestUpdatesRequest(page: Int): Request =
        throw UnsupportedOperationException("Not supported")

    override fun latestUpdatesParse(response: Response): MangasPage =
        throw UnsupportedOperationException("Not supported")

    // -- Manga details --

    override fun mangaDetailsRequest(manga: SManga): Request =
        GET("$baseUrl/api/v1/mihon/manga/${manga.url}", headers)

    override fun mangaDetailsParse(response: Response): SManga {
        val wrapper = json.parseToJsonElement(response.body.string()).jsonObject
        val data = wrapper["data"]?.jsonObject ?: return SManga.create()
        return parseManga(data)
    }

    // -- Chapters --

    override fun chapterListRequest(manga: SManga): Request =
        GET("$baseUrl/api/v1/mihon/manga/${manga.url}/chapters", headers)

    override fun chapterListParse(response: Response): List<SChapter> {
        val wrapper = json.parseToJsonElement(response.body.string()).jsonObject
        val data = wrapper["data"]?.jsonArray ?: return emptyList()
        return data.map { item ->
            val obj = item.jsonObject
            SChapter.create().apply {
                url = obj["dir"]!!.jsonPrimitive.content
                name = obj["title"]?.jsonPrimitive?.content ?: "Chapter"
                chapter_number = obj["number"]?.jsonPrimitive?.float ?: -1f
            }
        }.sortedByDescending { it.chapter_number }
    }

    // -- Pages --

    override fun pageListRequest(chapter: SChapter): Request =
        GET("$baseUrl/api/v1/mihon/chapter/${chapter.url}/pages", headers)

    override fun pageListParse(response: Response): List<Page> {
        val wrapper = json.parseToJsonElement(response.body.string()).jsonObject
        val data = wrapper["data"]?.jsonArray ?: return emptyList()
        return data.map { item ->
            val obj = item.jsonObject
            val index = obj["index"]!!.jsonPrimitive.int
            val url = obj["url"]!!.jsonPrimitive.content
            Page(index, "", "$baseUrl$url")
        }
    }

    override fun imageUrlParse(response: Response): String =
        throw UnsupportedOperationException("Not used")

    // -- Settings --

    override fun setupPreferenceScreen(screen: PreferenceScreen) {
        EditTextPreference(screen.context).apply {
            key = PREF_SERVER_URL
            title = "Seanime Server URL"
            summary = "The URL of your Seanime instance (e.g. https://seanime.example.com)"
            setDefaultValue(DEFAULT_SERVER_URL)
            dialogTitle = "Server URL"
        }.let(screen::addPreference)

        EditTextPreference(screen.context).apply {
            key = PREF_AUTH_TOKEN
            title = "Auth Token (optional)"
            summary = "If your Seanime instance has a password, enter the token here"
            setDefaultValue("")
            dialogTitle = "Auth Token"
        }.let(screen::addPreference)
    }

    // -- Helpers --

    private fun parseManga(obj: JsonObject): SManga = SManga.create().apply {
        url = obj["id"]!!.jsonPrimitive.content
        title = obj["title"]?.jsonPrimitive?.content ?: "Unknown"
        author = obj["author"]?.jsonPrimitive?.content
        artist = obj["artist"]?.jsonPrimitive?.content
        description = obj["description"]?.jsonPrimitive?.content
        thumbnail_url = obj["cover_url"]?.jsonPrimitive?.content
        genre = obj["genres"]?.jsonPrimitive?.content
        status = obj["status"]?.jsonPrimitive?.int ?: SManga.UNKNOWN
    }

    companion object {
        private const val PREF_SERVER_URL = "server_url"
        private const val PREF_AUTH_TOKEN = "auth_token"
        private const val DEFAULT_SERVER_URL = "http://localhost:43211"
    }
}
