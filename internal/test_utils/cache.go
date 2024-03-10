package test_utils

import "fmt"

func GetTestDataPath(name string) string {
	return fmt.Sprintf("%s/%s.json", TwoLevelDeepTestDataPath, name)
}
