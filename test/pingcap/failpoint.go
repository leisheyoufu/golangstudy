package main

import (
	"fmt"

	"github.com/pingcap/failpoint"
)

func main() {
	failpoint.Inject("testPanic", func() {
		panic("failpoint triggerd")
	})
	fmt.Println("Hello World")
}
