package anizip

import (
	"fmt"
	"testing"
)

func TestOffsetEpisode(t *testing.T) {

	inputStrings := []string{"S1", "OP1", "1", "OP"}

	for _, s := range inputStrings {
		modifiedStr := OffsetEpisode(s, 1)
		fmt.Printf("%s -> %s\n", s, modifiedStr)
	}

}
