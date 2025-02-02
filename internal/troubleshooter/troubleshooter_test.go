package troubleshooter

import (
	"path/filepath"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
)

func TestAnalyze(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	analyzer := NewAnalyzer(NewTroubleshooterOptions{
		LogsDir: filepath.Join(test_utils.ConfigData.Path.DataDir, "logs"),
	})

	res, err := analyzer.Analyze()
	if err != nil {
		t.Fatalf("Error analyzing logs: %v", err)
	}

	util.Spew(res)
}
