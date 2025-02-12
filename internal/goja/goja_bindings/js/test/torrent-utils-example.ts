/// <reference path="../crypto.d.ts" />
/// <reference path="../buffer.d.ts" />
/// <reference path="../torrent.d.ts" />

async function run() {
    try {

        console.log("\nTesting torrent file to magnet link")

        const url = "https://animetosho.org/storage/torrent/da9aad67b6f8bb82757bb3ef95235b42624c34f7/%5BSubsPlease%5D%20Make%20Heroine%20ga%20Oosugiru%21%20-%2011%20%281080p%29%20%5B58B3496A%5D.torrent"

        const data = await (await fetch(url)).text()
        
        const magnetLink = getMagnetLinkFromTorrentData(data)

        console.log("Magnet link:", magnetLink)
    }
    catch (e) {
        console.error(e)
    }
}
