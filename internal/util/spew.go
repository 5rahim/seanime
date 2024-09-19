package util

import (
	"fmt"
	"github.com/kr/pretty"
)

func Spew(v interface{}) {
	fmt.Printf("%# v\n", pretty.Formatter(v))
}
