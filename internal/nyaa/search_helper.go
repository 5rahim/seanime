package nyaa

import (
	"bytes"
	"fmt"
	"github.com/seanime-app/seanime-server/internal/anilist"
	"strings"
)

// (jjk|jujutsu kaisen)
func getTitleGroup(titles []string) string {
	return fmt.Sprintf("(%s)", strings.Join(titles, "|"))
}

func getAbsoluteGroup(title string, opts *BuildSearchQueryOptions) string {
	return fmt.Sprintf("(%s(%d))", title, *opts.EpisodeNumber+*opts.AbsoluteOffset)
}

// (s01e01)
func getSeasonAndEpisodeGroup(season int, ep int) string {
	if season == 0 {
		season = 1
	}
	return fmt.Sprintf(`"s%se%s"`, zeropad(season), zeropad(ep))
}

// (01|e01|e01v|ep01|ep1)
func getEpisodeGroup(ep int) string {
	pEp := zeropad(ep)
	//return fmt.Sprintf(`("%s"|"e%s"|"e%sv"|"%sv"|"ep%s"|"ep%d")`, pEp, pEp, pEp, pEp, pEp, ep)
	return fmt.Sprintf(`(%s|e%s|e%sv|%sv|ep%s|ep%d)`, pEp, pEp, pEp, pEp, pEp, ep)
}

// (season 1|season 01|s1|s01)
func getSeasonGroup(season int) string {
	// Season section
	seasonBuff := bytes.NewBufferString("")
	// e.g. S1, season 1, season 01
	if season != 0 {
		seasonBuff.WriteString(fmt.Sprintf(`("%s%d"|`, "season ", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%s"|`, "season ", zeropad(season)))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%d"|`, "s", season))
		seasonBuff.WriteString(fmt.Sprintf(`"%s%s")`, "s", zeropad(season)))
		//seasonBuff.WriteString(fmt.Sprintf(`(%s%d|`, "season ", season))
		//seasonBuff.WriteString(fmt.Sprintf(`%s%s|`, "season ", zeropad(season)))
		//seasonBuff.WriteString(fmt.Sprintf(`%s%d|`, "s", season))
		//seasonBuff.WriteString(fmt.Sprintf(`%s%s)`, "s", zeropad(season)))
	}
	return seasonBuff.String()
}
func getPartGroup(part int) string {
	partBuff := bytes.NewBufferString("")
	if part != 0 {
		partBuff.WriteString(fmt.Sprintf(`("%s%d")`, "part ", part))
	}
	return partBuff.String()
}

func getBatchGroup(m *anilist.BaseMedia) string {
	buff := bytes.NewBufferString("")
	buff.WriteString("(")
	// e.g. 01-12
	s1 := fmt.Sprintf(`"%s%s%s"`, zeropad("1"), " - ", zeropad(m.GetTotalEpisodeCount()))
	buff.WriteString(s1)
	buff.WriteString("|")
	// e.g. 01~12
	s2 := fmt.Sprintf(`"%s%s%s"`, zeropad("1"), " ~ ", zeropad(m.GetTotalEpisodeCount()))
	buff.WriteString(s2)
	buff.WriteString("|")
	// e.g. 01~12
	buff.WriteString(`"Batch"|`)
	buff.WriteString(`"Complete"|`)
	buff.WriteString(`"+ OVA"|`)
	buff.WriteString(`"+ Specials"|`)
	buff.WriteString(`"+ Special"|`)
	buff.WriteString(`"Seasons"|`)
	buff.WriteString(`"Parts"`)
	buff.WriteString(")")
	return buff.String()
}

func zeropad(v interface{}) string {
	switch i := v.(type) {
	case int:
		return fmt.Sprintf("%02d", i)
	case string:
		return fmt.Sprintf("%02s", i)
	default:
		return ""
	}
}
