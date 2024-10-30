package util

import (
	"fmt"
	"github.com/kr/pretty"
	"strings"
)

func Spew(v interface{}) {
	fmt.Printf("%# v\n", pretty.Formatter(v))
}

func SpewMany(v ...interface{}) {
	fmt.Println("Spewing values:")
	for _, val := range v {
		Spew(val)
	}
}

func SpewT(v interface{}) string {
	return fmt.Sprintf("%# v\n", pretty.Formatter(v))
}

func InlineSpewT(v interface{}) string {
	return strings.ReplaceAll(SpewT(v), "\n", "")
}
