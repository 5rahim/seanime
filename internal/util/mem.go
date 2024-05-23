package util

import "fmt"

func GetMemAddrStr(v interface{}) string {
	return fmt.Sprintf("%p", v)
}
