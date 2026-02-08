// import { HibikeManga_ChapterDetails } from "@/api/generated/types"
// import { getChapterNumberFromChapter } from "@/app/(main)/manga/_lib/handle-manga-utils"
// import { LANGUAGES_LIST } from "@/app/(main)/manga/_lib/language-map"
// import { IconButton } from "@/components/ui/button"
// import { Checkbox } from "@/components/ui/checkbox"
// import { cn } from "@/components/ui/core/styling"
// import { NumberInput } from "@/components/ui/number-input"
// import { Pagination, PaginationItem, PaginationTrigger } from "@/components/ui/pagination"
// import { Select } from "@/components/ui/select"
// import { RowSelectionState } from "@tanstack/react-table"
// import React from "react"
// import { BiChevronRight } from "react-icons/bi"
// import { MdOutlineDownloadForOffline, MdOutlineOfflinePin } from "react-icons/md"
// import { RiDownloadLine } from "react-icons/ri"

// type ChapterListTableProps = {
//     chapters: HibikeManga_ChapterDetails[]

//     // Selection
//     rowSelection: RowSelectionState
//     setRowSelection: (value: RowSelectionState) => void
//     setSelectedChapters: (value: HibikeManga_ChapterDetails[]) => void

//     // Actions
//     onChapterClick: (chapter: HibikeManga_ChapterDetails) => void
//     onDownloadChapter: (chapter: HibikeManga_ChapterDetails) => void

//     // Utils
//     isChapterQueued: (chapter: HibikeManga_ChapterDetails) => boolean
//     isChapterDownloaded: (chapter: HibikeManga_ChapterDetails) => boolean
//     isChapterLocal: (chapter: HibikeManga_ChapterDetails) => boolean
// }

// export function ChapterListTable(props: ChapterListTableProps) {
//     const {
//         chapters,
//         rowSelection,
//         setRowSelection,
//         setSelectedChapters,
//         isChapterDownloaded,
//         isChapterLocal,
//         isChapterQueued,
//         onDownloadChapter,
//         onChapterClick,
//     } = props

//     // Pagination state
//     const [perPage, setPerPage] = React.useState(10)
//     const [currentPage, setCurrentPage] = React.useState(1)

//     const totalPages = Math.ceil(chapters.length / perPage)

//     const displayedChapters = React.useMemo(() => {
//         return chapters.slice((currentPage - 1) * perPage, currentPage * perPage)
//     }, [chapters, currentPage, perPage])

//     React.useEffect(() => {
//         setCurrentPage(1)
//     }, [chapters.length])

//     const [lastSelectedId, setLastSelectedId] = React.useState<string | null>(null)

//     const handleSelect = React.useCallback((chapter: HibikeManga_ChapterDetails, checked: boolean, shiftKey: boolean) => {
//         let newSelection = { ...rowSelection }

//         if (isChapterDownloaded(chapter) || isChapterQueued(chapter)) return

//         if (shiftKey && lastSelectedId) {
//             const lastIndex = chapters.findIndex(c => c.id === lastSelectedId)
//             const currentIndex = chapters.findIndex(c => c.id === chapter.id)

//             if (lastIndex !== -1 && currentIndex !== -1) {
//                 const start = Math.min(lastIndex, currentIndex)
//                 const end = Math.max(lastIndex, currentIndex)

//                 for (let i = start; i <= end; i++) {
//                     const ch = chapters[i]
//                     if (!isChapterDownloaded(ch) && !isChapterQueued(ch)) {
//                         newSelection[ch.id] = true
//                     }
//                 }
//             }
//         } else {
//             if (checked) {
//                 newSelection[chapter.id] = true
//                 setLastSelectedId(chapter.id)
//             } else {
//                 delete newSelection[chapter.id]
//             }
//         }

//         setRowSelection(newSelection)

//         const selectedIds = new Set(Object.keys(newSelection))

//         const newSelectedChapters = chapters.filter(c => selectedIds.has(c.id))
//         setSelectedChapters(newSelectedChapters)

//     }, [chapters, rowSelection, lastSelectedId, isChapterDownloaded, isChapterQueued, setRowSelection, setSelectedChapters])

//     // Select All
//     const handleSelectAll = (checked: boolean) => {
//         let newSelection = { ...rowSelection }
//         if (checked) {
//             chapters.forEach(ch => {
//                 if (!isChapterDownloaded(ch) && !isChapterQueued(ch)) {
//                     newSelection[ch.id] = true
//                 }
//             })
//         } else {
//             chapters.forEach(ch => {
//                 delete newSelection[ch.id]
//             })
//         }
//         setRowSelection(newSelection)
//         const selectedIds = new Set(Object.keys(newSelection))
//         const newSelectedChapters = chapters.filter(c => selectedIds.has(c.id))
//         setSelectedChapters(newSelectedChapters)
//     }

//     const areAllSelected = chapters.length > 0 && chapters.every(ch =>
//         (isChapterQueued(ch) || isChapterDownloaded(ch)) || rowSelection[ch.id],
//     ) && chapters.some(ch => !isChapterQueued(ch) && !isChapterDownloaded(ch))

//     return (
//         <div className="space-y-4">
//             <div className="flex items-center gap-2 px-2 pb-2 border-b border-[--border] text-sm text-[--muted]">
//                 <div className="flex items-center gap-2">
//                     <Checkbox
//                         value={areAllSelected}
//                         onValueChange={(v) => handleSelectAll(v as boolean)}
//                         label="Select all"
//                         disabled={chapters.every(ch => isChapterQueued(ch) || isChapterDownloaded(ch))}
//                     />
//                 </div>
//                 <div className="flex-1 text-right">
//                     {chapters.length} chapters
//                 </div>
//             </div>

//             {chapters.length === 0 && (
//                 <div className="p-4 text-center text-[--muted]">
//                     No chapters found
//                 </div>
//             )}

//             <div className="space-y-2">
//                 {displayedChapters.map(chapter => (
//                     <ChapterListTableRow
//                         key={chapter.id}
//                         chapter={chapter}
//                         isSelected={!!rowSelection[chapter.id]}
//                         onSelect={(checked, shift) => handleSelect(chapter, checked, shift)}
//                         isQueued={isChapterQueued(chapter)}
//                         isDownloaded={isChapterDownloaded(chapter)}
//                         isLocal={isChapterLocal(chapter)}
//                         onDownload={() => onDownloadChapter(chapter)}
//                         onClick={() => onChapterClick(chapter)}
//                     />
//                 ))}
//             </div>

//             {totalPages > 1 && <div className="flex flex-col md:flex-row justify-center items-center gap-4 mt-4">
//                 <Pagination>
//                     <PaginationTrigger
//                         direction="previous"
//                         isChevrons
//                         isDisabled={currentPage === 1}
//                         onClick={() => setCurrentPage(1)}
//                     />
//                     <PaginationTrigger
//                         direction="previous"
//                         isDisabled={currentPage === 1}
//                         onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
//                     />
//                     {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
//                         let p = currentPage - 2 + i
//                         if (currentPage < 3) p = 1 + i
//                         if (currentPage > totalPages - 2) p = totalPages - 4 + i
//                         if (p < 1 || p > totalPages) return null
//                         return (
//                             <PaginationItem
//                                 key={p}
//                                 value={p}
//                                 data-selected={p === currentPage}
//                                 onClick={() => setCurrentPage(p)}
//                             />
//                         )
//                     })}
//                     <PaginationTrigger
//                         direction="next"
//                         isDisabled={currentPage === totalPages}
//                         onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
//                     />
//                     <PaginationTrigger
//                         direction="next"
//                         isChevrons
//                         isDisabled={currentPage === totalPages}
//                         onClick={() => setCurrentPage(totalPages)}
//                     />
//                 </Pagination>

//                 <div className="flex-1"></div>

//                 <div className="flex items-center gap-2 text-sm">
//                     <NumberInput
//                         value={currentPage}
//                         min={1}
//                         max={totalPages}
//                         onValueChange={(v) => setCurrentPage(v)}
//                         size="sm"
//                         className="w-16"
//                         hideControls
//                     />
//                     <Select
//                         options={[
//                             { value: "10", label: "10" },
//                             { value: "20", label: "20" },
//                             { value: "50", label: "50" },
//                             { value: "100", label: "100" },
//                         ]}
//                         value={String(perPage)}
//                         onValueChange={(v) => setPerPage(Number(v))}
//                         size="sm"
//                         className="w-16"
//                     />
//                 </div>
//             </div>}
//         </div>
//     )
// }

// type ChapterListTableRowProps = {
//     chapter: HibikeManga_ChapterDetails
//     isSelected: boolean
//     onSelect: (checked: boolean, shiftKey: boolean) => void
//     isQueued: boolean
//     isDownloaded: boolean
//     isLocal: boolean
//     onDownload: () => void
//     onClick: () => void
// }

// function ChapterListTableRow(props: ChapterListTableRowProps) {
//     const { chapter, isSelected, onSelect, isQueued, isDownloaded, isLocal, onDownload, onClick } = props

//     const chapterNumber = getChapterNumberFromChapter(chapter.chapter)
//     const canSelect = !isQueued && !isDownloaded

//     return (
//         <div
//             className={cn(
//                 "group relative overflow-hidden rounded-[--radius] bg-gray-950 hover:bg-gray-800/50 transition-colors",
//                 isSelected && "ring-2 ring-[--border]",
//                 // (isQueued) && "opacity-80",
//             )}
//             onClick={onClick}
//         >
//             <div className="flex justify-between py-2 px-3 gap-3 items-center cursor-pointer">

//                 <div className="flex items-center gap-3 w-full overflow-hidden">
//                     <div className="flex-shrink-0" onClick={(e) => e.stopPropagation()}>
//                         <Checkbox
//                             value={isSelected}
//                             onValueChange={(ch) => onSelect(ch as boolean, false)}
//                             disabled={!canSelect}
//                             className={cn(
//                                 (!canSelect) && "invisible",
//                             )}
//                         />
//                         <div
//                             className="absolute inset-y-0 left-0 w-10 z-10 cursor-pointer"
//                             onClick={(e) => {
//                                 if (!canSelect) return
//                                 e.preventDefault()
//                                 e.stopPropagation()
//                                 onSelect(!isSelected, e.shiftKey)
//                             }}
//                             style={{ opacity: 0 }}
//                         />
//                     </div>

//                     <div className="space-y-1 w-full overflow-hidden">
//                         <p className={cn("font-medium text-base tracking-wide line-clamp-1 text-gray-200")}>
//                             <span className="font-medium pr-1">{chapter.title || `Chapter ${chapter.chapter}`}</span>
//                         </p>
//                         <div className="text-sm text-gray-400 line-clamp-1 flex space-x-2 items-center divide-x divide-gray-700 [&>span]:pl-2">
//                             <span>#{chapterNumber}</span>
//                             {chapter.scanlator && <span>{chapter.scanlator}</span>}
//                             {chapter.language && <span>{LANGUAGES_LIST[chapter.language]?.nativeName || chapter.language}</span>}
//                         </div>
//                     </div>
//                 </div>

//                 <div className="flex items-center gap-2 flex-shrink-0">
//                     {isQueued && (
//                         <span className="text-xs font-medium text-[--muted] flex items-center gap-1">
//                             <RiDownloadLine className="animate-pulse" /> Queued
//                         </span>
//                     )}

//                     {isDownloaded && (
//                         <span className="text-green-500 text-lg" title="Downloaded">
//                             <MdOutlineOfflinePin />
//                         </span>
//                     )}

//                     {!isDownloaded && !isQueued && !isLocal && (
//                         <IconButton
//                             intent="gray-basic"
//                             size="sm"
//                             icon={<MdOutlineDownloadForOffline className="text-xl" />}
//                             onClick={(e) => {
//                                 e.stopPropagation()
//                                 onDownload()
//                             }}
//                             className="opacity-0 group-hover:opacity-100 transition-opacity"
//                         />
//                     )}

//                     <IconButton
//                         intent="white-basic"
//                         icon={<BiChevronRight />}
//                         size="sm"
//                         onClick={(e) => {
//                             e.stopPropagation()
//                             onClick()
//                         }}
//                     />
//                 </div>
//             </div>
//         </div>
//     )
// }
