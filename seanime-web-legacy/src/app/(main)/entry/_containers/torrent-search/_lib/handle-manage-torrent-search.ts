// export function useManageTorrentSearch() {
//
//
//     const [selectionType, setSelection] = useAtom(__torrentSearch_selectionAtom)
//     const [episodeNumber, setTorrentSearchEpisodeNumber] = useAtom(__torrentSearch_selectionEpisodeAtom)
//
//     function openTorrentSearch(type: TorrentSelectionType) {
//         setSelection(type)
//     }
//     function openTorrentSearchForEpisode(type: TorrentSelectionType) {
//         setSelection(type)
//     }
//
//     return {
//         selectionType,
//         episodeNumber,
//         openTorrentSearch,
//         closeTorrentSearch: () => setSelection(undefined),
//     }
// }
