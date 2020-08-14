package main

import (
	"errors"
	"fmt"
	"github.com/leisheyoufu/golangstudy/base/retry/wait"
	"time"
)

const (
	maxTimes = 10
)

var (
	times int
)

// A test function only return true when call times reachs.
// parameter is only for test
func Caller(message string) error {
	if times == maxTimes {
		times = 0
		return nil
	}
	times++
	fmt.Printf("Hello %s %d\n", message, times)
	return errors.New("Caller error")
}

func main() {
	err := wait.Poll(1*time.Second, 3*time.Second, func() (bool, error) {
		fmt.Printf("retry %d\n", times)
		if err := Caller("World"); err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
	fmt.Printf("Wait caller function succesfully\n")
	err = wait.Poll(1*time.Second, 15*time.Second, func() (bool, error) {
		fmt.Printf("retry %d\n", times)
		if err := Caller("World"); err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
	}
	fmt.Printf("Wait caller function succesfully\n")
}
