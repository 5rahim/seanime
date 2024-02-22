package util

import (
	"fmt"
	"time"
)

func TimestampToDateStr(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	return fmt.Sprintf("%v", tm)
}
