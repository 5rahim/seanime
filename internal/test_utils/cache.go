package test_utils

import "fmt"

func GetTestDataPath(name string) string {
	return fmt.Sprintf("%s/%s.json", TwoLevelDeepTestDataPath, name)
}

func GetDataPath(name string) string {
	return fmt.Sprintf("%s/%s.json", TwoLevelDeepDataPath, name)
}
