package anilist

func (m *BaseMedia) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}

func (m *BasicMedia) GetTitleSafe() string {
	if m.GetTitle().GetEnglish() != nil {
		return *m.GetTitle().GetEnglish()
	}
	if m.GetTitle().GetRomaji() != nil {
		return *m.GetTitle().GetRomaji()
	}
	return "N/A"
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

var EdgeNarrowFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort}
var EdgeBroaderFormats = []MediaFormat{MediaFormatTv, MediaFormatTvShort, MediaFormatOna, MediaFormatOva, MediaFormatMovie, MediaFormatSpecial}

func (m *BaseMedia) FindEdge(relation string, formats []MediaFormat) (*BasicMedia, bool) {
	if m.GetRelations() == nil {
		return nil, false
	}

	edges := m.GetRelations().GetEdges()

	for _, edge := range edges {

		if edge.GetRelationType().String() == relation {
			for _, fm := range formats {
				if fm.String() == edge.GetNode().GetFormat().String() {
					return edge.GetNode(), true
				}
			}
		}

	}
	return nil, false
}

func (e *BaseMedia_Relations_Edges) IsBroadRelationFormat() bool {
	if e.GetNode() == nil {
		return false
	}
	if e.GetNode().GetFormat() == nil {
		return false
	}
	for _, fm := range EdgeBroaderFormats {
		if fm.String() == e.GetNode().GetFormat().String() {
			return true
		}
	}
	return false
}
func (e *BaseMedia_Relations_Edges) IsNarrowRelationFormat() bool {
	if e.GetNode() == nil {
		return false
	}
	if e.GetNode().GetFormat() == nil {
		return false
	}
	for _, fm := range EdgeNarrowFormats {
		if fm.String() == e.GetNode().GetFormat().String() {
			return true
		}
	}
	return false
}
