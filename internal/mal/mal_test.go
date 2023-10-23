package mal

import "testing"

func TestSearchWithMAL(t *testing.T) {

	res, err := SearchWithMAL("bungo stray dogs", 4)

	if err != nil {
		t.Fatalf("error while fetching media, %v", err)
	}

	for _, m := range res {
		t.Log(m.Name)
	}

}

func TestAdvancedSearchWithMal(t *testing.T) {

	res, err := AdvancedSearchWithMAL("sousou no frieren")

	if err != nil {
		t.Fatal("expected result, got error: ", err)
	}

	t.Log(res.Name)

}
