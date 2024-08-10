package notifier

import (
	"fmt"
	"seanime/internal/database/models"
	"seanime/internal/test_utils"
	"seanime/internal/util"
	"testing"
	"time"
)

func TestNotifier(t *testing.T) {
	test_utils.SetTwoLevelDeep()
	test_utils.InitTestProvider(t)

	GlobalNotifier = NewNotifier()

	GlobalNotifier.SetSettings(test_utils.ConfigData.Path.DataDir, &models.NotificationSettings{}, util.NewLogger())

	GlobalNotifier.Notify(
		AutoDownloader,
		fmt.Sprintf("%d %s %s been downloaded or added to the queue.", 1, util.Pluralize(1, "episode", "episodes"), util.Pluralize(1, "has", "have")),
	)

	GlobalNotifier.Notify(
		AutoScanner,
		fmt.Sprintf("%d %s %s been downloaded or added to the queue.", 1, util.Pluralize(1, "episode", "episodes"), util.Pluralize(1, "has", "have")),
	)

	time.Sleep(1 * time.Second)

}
