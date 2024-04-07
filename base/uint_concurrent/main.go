package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	var g uint32
	runtime.GOMAXPROCS(8)

	for i := 0; i < 1000000; i++ {
		var wg sync.WaitGroup
		// 协程 1
		wg.Add(1)
		go func() {
			defer wg.Done()
			g = 2
			fmt.Printf("g=%d\n", g)
		}()

		// 协程 2
		wg.Add(1)
		go func() {
			defer wg.Done()
			g = 3
			fmt.Printf("g=%d\n", g)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			g = 4
			fmt.Printf("g=%d\n", g)
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			g = 5
			fmt.Printf("g=%d\n", g)
		}()
		wg.Wait()
	}
	fmt.Printf("done\n")
}
