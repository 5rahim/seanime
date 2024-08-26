/// <reference path="../doc.d.ts" />

class Provider {
    async test() {
        try {
            const data = await fetch("https://cryptojs.gitbook.io/docs")

            const $ = new Doc(await data.text())

            console.log($.find("header h1").text())
        }
        catch (e) {
            console.error(e)
        }
    }
}
