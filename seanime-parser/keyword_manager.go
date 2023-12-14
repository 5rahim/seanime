package seanime_parser

import (
	"slices"
	"strings"

	"golang.org/x/text/unicode/norm"
)

type keywordCategory string

const (
	keywordCatNone               keywordCategory = ""
	keywordCatSeasonPrefix                       = "seasonPrefix"
	keywordCatAnimeType                          = "animeType"
	keywordCatYear                               = "year"
	keywordCatAudioTerm                          = "audioTerm"
	keywordCatDeviceCompat                       = "deviceCompat"
	keywordCatEpisodePrefix                      = "episodePrefix"
	keywordCatPartPrefix                         = "partPrefix"
	keywordCatVolumePrefix                       = "volumePrefix"
	keywordCatFileChecksum                       = "fileChecksum"
	keywordCatFileExtension                      = "fileExtension"
	keywordCatLanguage                           = "language"
	keywordCatReleaseGroup                       = "releaseGroup"
	keywordCatReleaseInformation                 = "releaseInfo"
	keywordCatReleaseVersion                     = "releaseVersion"
	keywordCatSource                             = "source"
	keywordCatSubtitles                          = "subtitles"
	keywordCatVideoResolution                    = "videoRes"
	keywordCatVideoTerm                          = "videoTerm"
)

type keywordKind uint8

const (
	keywordKindCombinedWithNumber keywordKind = iota
	keywordKindSeparatedWithNumber
	keywordKindOrdinalSuffix
	keywordKindStandalone
)

type (
	keyword struct {
		value    string
		category keywordCategory
		kind     keywordKind
	}

	keywordParts struct {
		category keywordCategory
		prefix   string
		seqParts []string
	}

	keywordManager struct {
		keywords          []*keyword
		keywordParts      []*keywordParts
		ambiguousKeywords []string
	}
)

func newKeywordManager() *keywordManager {
	km := keywordManager{
		keywords: make([]*keyword, 0),
	}

	// The parser will treat these keywords as standalone ambiguous tokens and will not attempt to identify them if
	// they are found in the first half of the filename.
	km.ambiguousKeywords = []string{
		"SP", "ANDROID", "ITA", "ESP", "FR", "JP", "EN", "RU", "CH",
	}

	// Season
	// Order matters

	km.addGroup( // First look for ordinal suffix (e.g. "2nd Season")
		keywordCatSeasonPrefix,
		keywordKindOrdinalSuffix,
		[]string{"SEASON", "SAISON", "SEASONS", "SAISONS"},
	)
	km.addGroup( // Then look for combined with number (e.g. "S01")
		keywordCatSeasonPrefix,
		keywordKindCombinedWithNumber,
		[]string{"S"},
	)
	km.addGroup( // Then look for separated with number (e.g. "Season 01")
		keywordCatSeasonPrefix,
		keywordKindSeparatedWithNumber,
		[]string{"SEASON", "SAISON", "SEASONS", "SAISONS"},
	)

	// Episode

	km.addGroup(
		keywordCatEpisodePrefix,
		keywordKindCombinedWithNumber,
		[]string{"E", "#", "\x7B2C", "EP", "EPS", "EPISODE", "EPISODES", "CAPITULO", "EPISODIO", "FOLDE"},
	)
	km.addGroup(
		keywordCatEpisodePrefix,
		keywordKindSeparatedWithNumber,
		[]string{"EP", "EPS", "EPISODE", "EPISODES", "CAPITULO", "EPISODIO", "FOLDE"},
	)

	// Volume

	km.addGroup(
		keywordCatVolumePrefix,
		keywordKindCombinedWithNumber,
		[]string{"VOL", "VOLUME", "VOLUMES"},
	)
	km.addGroup(
		keywordCatVolumePrefix,
		keywordKindSeparatedWithNumber,
		[]string{"VOL", "VOLUME", "VOLUMES"},
	)

	// Part

	km.addGroup(
		keywordCatPartPrefix,
		keywordKindCombinedWithNumber,
		[]string{"PART", "PARTS", "COUR"},
	)
	km.addGroup(
		keywordCatPartPrefix,
		keywordKindSeparatedWithNumber,
		[]string{"PART", "PARTS", "COUR"},
	)
	km.addGroup(
		keywordCatPartPrefix,
		keywordKindOrdinalSuffix,
		[]string{"COUR"},
	)

	// Anime Type

	km.addGroup(
		keywordCatAnimeType,
		keywordKindCombinedWithNumber,
		[]string{"SP"},
	)
	km.addGroup(
		keywordCatAnimeType,
		keywordKindCombinedWithNumber,
		[]string{"SP", "MOVIE", "OAD", "OAV", "ONA", "OVA", "SPECIAL", "SPECIALS", "ED", "ENDING", "NCED", "NCOP", "OPED", "OP", "OPENING",
			"番外編", "總集編", "映像特典", "特典", "特典アニメ"},
	)
	km.addGroup(
		keywordCatAnimeType,
		keywordKindSeparatedWithNumber,
		[]string{"SP", "MOVIE", "OAD", "OAV", "ONA", "OVA", "SPECIAL", "SPECIALS", "ED", "ENDING", "NCED", "NCOP", "OPED", "OP", "OPENING",
			"番外編", "總集編", "映像特典", "特典", "特典アニメ"},
	)
	km.addGroup(
		keywordCatAnimeType,
		keywordKindStandalone,
		[]string{
			"MOVIE", "GEKIJOUBAN", "ONA", "OVA", "OAV", "OAD", "SPECIALS", "TV",
			"ED", "ENDING", "NCED", "NCOP", "OPED", "OP", "OPENING", "PREVIEW", "PV", "EVENT", "TOKUTEN", "LOGO", "CM", "SPOT", "MENU"},
	)

	km.addGroupParts(
		keywordCatReleaseGroup,
		[]*keywordParts{
			{prefix: "CENTRAL", seqParts: []string{" ", "ANIME"}},
		},
	)

	// Audio Term

	km.addGroup(
		keywordCatAudioTerm,
		keywordKindStandalone,
		[]string{
			"2CH", "DTS", "DTS-ES", "DTS5.1", "TRUEHD5.1", "5.1", "2.0",
			"AAC", "AACX2", "AACX3", "AACX4", "AC3", "EAC3", "E-AC-3", "FLAC",
			"FLACX2", "FLACX3", "FLACX4", "LOSSLESS", "MP3", "OGG", "VORBIS",
			"DD2", "DD2.0", "DDP", "DD", "HDMA", "DTSHD",
			"DUALAUDIO", "DUAL-AUDIO",
		},
	)

	km.addGroupParts(
		keywordCatAudioTerm,
		[]*keywordParts{
			{prefix: "2", seqParts: []string{".", "0CH"}},                   // 2.0CH
			{prefix: "5", seqParts: []string{".", "1"}},                     // 5.1
			{prefix: "5", seqParts: []string{".", "1", "+", "2", ".", "0"}}, // 5.1
			{prefix: "5", seqParts: []string{".", "1ch"}},                   // 5.1ch
			{prefix: "DTS", seqParts: []string{"-", "ES"}},                  // DTS-ES
			{prefix: "DTS", seqParts: []string{"-", "HD"}},                  // DTS-HD
			{prefix: "DTS", seqParts: []string{"-", "HDMA"}},                // DTS-HDMA
			{prefix: "DTS5", seqParts: []string{".", "1"}},                  // DTS5.1
			{prefix: "TRUEHD5", seqParts: []string{".", "1"}},               // TRUEHD5.1
			{prefix: "DUAL", seqParts: []string{"-", "AUDIO"}},              // DUAL-AUDIO
			{prefix: "DUAL", seqParts: []string{".", "AUDIO"}},              // DUAL.AUDIO
			{prefix: "DUAL", seqParts: []string{" ", "AUDIO"}},              // DUAL AUDIO
			{prefix: "DD2", seqParts: []string{".", "0"}},                   // DD2.0
			{prefix: "E", seqParts: []string{"-", "AC", "-", "3"}},          // E-AC-3
		},
	)

	// Video Term

	km.addGroup(
		keywordCatVideoTerm,
		keywordKindStandalone,
		[]string{
			// Frame rate
			"24FPS", "30FPS", "60FPS", "120FPS",
			// Video codec
			"8BIT", "10BIT", "10BITS",
			"HI10", "HI10P", "HI444", "HI444P", "HI444PP",
			"H264", "H265", "X264", "X265",
			"AVC", "HEVC", "HEVC2", "DIVX", "DIVX5", "DIVX6", "XVID",
			"AV1",
			"HDR", "DV",
			// Video format
			"AVI", "RMVB", "WMV", "WMV3", "WMV9",
			// Video quality
			"HQ", "LQ",
			// Video resolution
			"HD", "SD", "4K",
		},
	)

	km.addGroupParts(
		keywordCatVideoTerm,
		[]*keywordParts{
			{prefix: "23", seqParts: []string{".", "976FPS"}},    // 23.976FPS
			{prefix: "29", seqParts: []string{".", "97FPS"}},     // 29.97FPS
			{prefix: "8", seqParts: []string{".", "BIT"}},        // 8.BIT
			{prefix: "8", seqParts: []string{"-", "BIT"}},        // 8-BIT
			{prefix: "10", seqParts: []string{" ", "BIT"}},       // 10 BIT
			{prefix: "10", seqParts: []string{".", "BIT"}},       // 10.BIT
			{prefix: "10", seqParts: []string{"-", "BIT"}},       // 10-BIT
			{prefix: "10", seqParts: []string{"-", "BITS"}},      // 10-BITS
			{prefix: "10", seqParts: []string{" ", "BITS"}},      // 10 BITS
			{prefix: "10", seqParts: []string{".", "BITS"}},      // 10.BITS
			{prefix: "H", seqParts: []string{".", "264"}},        // H.264
			{prefix: "H", seqParts: []string{" ", "264"}},        // H 264
			{prefix: "H", seqParts: []string{".", "265"}},        // H.265
			{prefix: "H", seqParts: []string{" ", "265"}},        // H 265
			{prefix: "X", seqParts: []string{".", "264"}},        // X.264
			{prefix: "DOLBY", seqParts: []string{" ", "VISION"}}, // DOLBY VISION
		},
	)

	// Release version

	km.addGroup(
		keywordCatReleaseVersion,
		keywordKindStandalone,
		[]string{"V0", "V1", "V2", "V3", "V4", "V5"},
	)

	// Device Compat

	km.addGroup(
		keywordCatDeviceCompat,
		keywordKindStandalone,
		[]string{"IPAD3", "IPHONE5", "IPOD", "PS3", "XBOX", "XBOX360", "ANDROID"},
	)

	// File Extension
	// should be last

	km.addGroup(
		keywordCatFileExtension,
		keywordKindStandalone,
		[]string{"3GP", "AVI", "DIVX", "FLV", "M2TS", "MKV", "MOV", "MP4", "MPG",
			"OGM", "RM", "RMVB", "TS", "WEBM", "WMV"},
	)

	// Language
	// should be enclosed

	km.addGroup(
		keywordCatLanguage,
		keywordKindStandalone,
		[]string{"ENG", "ENGLISH", "ESPANOL", "JAP", "JP", "EN", "JPN", "FR", "PT-BR", "SPANISH", "VOSTFR", "ESP", "ITA", "RU", "CHT", "CHS", "CH"},
	)

	km.addGroupParts(
		keywordCatLanguage,
		[]*keywordParts{
			{prefix: "PT", seqParts: []string{"-", "BR"}},
			{prefix: "PT", seqParts: []string{".", "BR"}},
			{prefix: "PT", seqParts: []string{" ", "BR"}},
		},
	)

	// Release info

	km.addGroup(
		keywordCatReleaseInformation,
		keywordKindStandalone,
		[]string{"REMASTER", "REMASTERED", "UNCENSORED", "UNCUT", "TS", "VFR",
			"WIDESCREEN", "WS", "BATCH", "COMPLETE", "PATCH", "REMUX", "FINAL"},
	)

	km.addGroup(
		keywordCatSubtitles,
		keywordKindStandalone,
		[]string{"ASS", "BIG5", "DUB", "DUBBED", "HARDSUB", "HARDSUBS", "RAW",
			"SOFTSUB", "SOFTSUBS", "SUB", "SUBBED", "SUBTITLED", "MULTISUB", "MULTIAUDIO"},
	)

	km.addGroupParts(
		keywordCatSubtitles,
		[]*keywordParts{
			{prefix: "MULTI", seqParts: []string{"_", "SUB"}},
			{prefix: "MULTI", seqParts: []string{" ", "SUB"}},
			{prefix: "MULTI", seqParts: []string{"-", "SUB"}},
			{prefix: "MULTI", seqParts: []string{".", "SUB"}},
			{prefix: "MULTI", seqParts: []string{"-", "SUBS"}},
			{prefix: "MULTI", seqParts: []string{" ", "SUBS"}},
			{prefix: "MULTI", seqParts: []string{".", "SUBS"}},
			{prefix: "MULTI", seqParts: []string{"-", "AUDIO"}},
			{prefix: "MULTI", seqParts: []string{" ", "AUDIO"}},
			{prefix: "MULTI", seqParts: []string{".", "AUDIO"}},
		},
	)

	// Source

	km.addGroup(
		keywordCatSource,
		keywordKindStandalone,
		[]string{"BD", "ASF", "BDRIP", "BLURAY", "BLU-RAY", "DVD", "DVD5", "DVD9",
			"DVD-R2J", "DVDRIP", "DVD-RIP", "R2DVD", "R2J", "R2JDVD",
			"R2JDVDRIP", "HDTV", "HDTVRIP", "TVRIP", "TV-RIP",
			"WEBCAST", "WEBRIP"},
	)

	km.addGroupParts(
		keywordCatSource,
		[]*keywordParts{
			{prefix: "BLU", seqParts: []string{"-", "RAY"}},
			{prefix: "BLU", seqParts: []string{" ", "RAY"}},
			{prefix: "DVD", seqParts: []string{"-", "R2J"}},
			{prefix: "DVD", seqParts: []string{"-", "RIP"}},
			{prefix: "DVD", seqParts: []string{" ", "RIP"}},
			{prefix: "TV", seqParts: []string{"-", "RIP"}},
			{prefix: "TV", seqParts: []string{" ", "RIP"}},
		},
	)

	return &km
}

func (km *keywordManager) addGroup(category keywordCategory, kind keywordKind, group []string) {
	for _, value := range group {
		km.keywords = append(km.keywords, &keyword{
			value:    value,
			category: category,
			kind:     kind,
		})
	}
}

func (km *keywordManager) addGroupParts(category keywordCategory, group []*keywordParts) {
	for _, value := range group {
		km.keywordParts = append(km.keywordParts, &keywordParts{
			category: category,
			prefix:   value.prefix,
			seqParts: value.seqParts,
		})
	}
}

func normalize(text string) string {
	f := norm.Form(3)

	return strings.ToUpper(string(f.Bytes([]byte(text))))
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (km *keywordManager) findKeywordPartGroups(s string) ([]*keywordParts, bool) {
	partsToTest := make([]*keywordParts, 0)
	for _, kwParts := range km.keywordParts {
		if normalize(s) == kwParts.prefix {
			partsToTest = append(partsToTest, kwParts)
		}
	}

	if len(partsToTest) == 0 {
		return partsToTest, false
	}

	return partsToTest, true
}

func (km *keywordManager) findStandaloneKeywordByValue(s string) (*keyword, bool) {
	var keyword *keyword
	for _, kw := range km.keywords {
		if normalize(s) == kw.value && kw.kind == keywordKindStandalone {
			keyword = kw
			break
		}
	}

	if keyword == nil {
		return nil, false
	}

	return keyword, true
}

func (km *keywordManager) findKeywordsBy(pred func(kw *keyword) bool) ([]*keyword, bool) {

	var kws []*keyword
	for _, kw := range km.keywords {
		if pred(kw) {
			kws = append(kws, kw)
		}
	}

	if len(kws) == 0 {
		return nil, false
	}

	return kws, true
}
func (km *keywordManager) isKeywordAmbiguous(kw *keyword) bool {

	return slices.Contains(km.ambiguousKeywords, kw.value)

}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

//func (k *keyword) getMetadataKindFromKeywordCategory(cat keywordCategory) metadataCategory {
//	switch cat {
//	case keywordCatSource:
//		return metadataKindSource
//	case keywordCatReleaseInformation:
//		return metadataKindReleaseInformation
//	case keywordCatReleaseVersion:
//		return metadataKindReleaseVersion
//	case keywordCatReleaseGroup:
//		return metadataKindReleaseGroup
//	case keywordCatLanguage:
//		return metadataKindLanguage
//	case keywordCatFileExtension:
//		return metadataKindFileExtension
//	case keywordCatVolumePrefix:
//
//	}
//}
