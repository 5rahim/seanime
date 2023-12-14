package seanime_parser

func (k *keyword) getValue() string {
	return k.value
}

func (k *keyword) getCategory() keywordCategory {
	return k.category
}

func (k *keyword) getKind() keywordKind {
	return k.kind
}

func (k *keyword) isNone() bool {
	return k.category == keywordCatNone
}

func (k *keyword) isStandalone() bool {
	return k.kind == keywordKindStandalone
}

func (k *keyword) isCombinedWithNumber() bool {
	return k.kind == keywordKindCombinedWithNumber
}

func (k *keyword) isSeparatedWithNumber() bool {
	return k.kind == keywordKindSeparatedWithNumber
}

func (k *keyword) isOrdinalSuffix() bool {
	return k.kind == keywordKindOrdinalSuffix
}

func (k *keyword) isSeasonPrefix() bool {
	return k.category == keywordCatSeasonPrefix
}

func (k *keyword) isAnimeType() bool {
	return k.category == keywordCatAnimeType
}

func (k *keyword) isYear() bool {
	return k.category == keywordCatYear
}

func (k *keyword) isAudioTerm() bool {
	return k.category == keywordCatAudioTerm
}

func (k *keyword) isDeviceCompat() bool {
	return k.category == keywordCatDeviceCompat
}

func (k *keyword) isEpisodePrefix() bool {
	return k.category == keywordCatEpisodePrefix
}

func (k *keyword) isPartPrefix() bool {
	return k.category == keywordCatPartPrefix
}

func (k *keyword) isVolumePrefix() bool {
	return k.category == keywordCatVolumePrefix
}

func (k *keyword) isFileChecksum() bool {
	return k.category == keywordCatFileChecksum
}

func (k *keyword) isFileExtension() bool {
	return k.category == keywordCatFileExtension
}

func (k *keyword) isLanguage() bool {
	return k.category == keywordCatLanguage
}

func (k *keyword) isReleaseGroup() bool {
	return k.category == keywordCatReleaseGroup
}

func (k *keyword) isReleaseInformation() bool {
	return k.category == keywordCatReleaseInformation
}

func (k *keyword) isReleaseVersion() bool {
	return k.category == keywordCatReleaseVersion
}

func (k *keyword) isSource() bool {
	return k.category == keywordCatSource
}

func (k *keyword) isSubtitles() bool {
	return k.category == keywordCatSubtitles
}

func (k *keyword) isVideoResolution() bool {
	return k.category == keywordCatVideoResolution
}

func (k *keyword) isVideoTerm() bool {
	return k.category == keywordCatVideoTerm
}
