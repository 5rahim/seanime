package transcoder

import (
	"errors"
	"fmt"
)

func ParseSegment(segment string) (int32, error) {
	var ret int32
	_, err := fmt.Sscanf(segment, "segment-%d.ts", &ret)
	if err != nil {
		return 0, errors.New("could not parse segment")
	}
	return ret, nil
}
