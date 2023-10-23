package matching

import "testing"

func TestFindBestMatchWithLevenstein(t *testing.T) {

	title := "jujutsu kaisen 2"
	titles := []string{"JJK", "Jujutsu Kaisen", "Jujutsu Kaisen 2"}
	expected := "Jujutsu Kaisen 2"

	if res, ok := FindBestMatchWithLevenstein(title, titles); ok {
		if res.Value != expected {
			t.Errorf("expected %s for %s, got %s", expected, title, res.Value)
		}
		if res.Distance != 0 {
			t.Errorf("expected a distance of 0, got %d", res.Distance)
		}
		t.Logf("value: %s, distance: %d", res.Value, res.Distance)
	} else {
		t.Error("expected result, got nil")
	}

}
