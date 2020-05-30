// reference: https://studygolang.com/articles/04189
package main

import (
	"fmt"
	"net/http"

	"golang.org/x/sync/errgroup"
	"sync"
)

func TestWaitGroup() {
	var wg sync.WaitGroup
	var urls = [...]string{
		"http://www.126.com/",
		"http://www.baidu.com/",
		"http://www.qq.com",
	}
	for _, url := range urls {
		// Increment the WaitGroup counter.
		wg.Add(1)
		// Launch a goroutine to fetch the URL.
		go func(url string) {
			// Decrement the counter when the goroutine completes.
			defer wg.Done()
			// Fetch the URL.
			http.Get(url)
		}(url)
	}
	// Wait for all HTTP fetches to complete.
	wg.Wait()
}

func TestErrorGroup() {
	var g errgroup.Group
	var err error
	var urls = []string{
		"http://www.126.com/",
		"http://www.baidu.com/",
		"http://www.qq33daffdaer.com/",
	}
	for _, url := range urls {
		// Launch a goroutine to fetch the URL.
		url := url // https://golang.org/doc/faq#closures_and_goroutines
		g.Go(func() error {
			// Fetch the URL.
			resp, err := http.Get(url)
			if err == nil {
				resp.Body.Close()
			}
			return err
		})
	}

	// Wait for all HTTP fetches to complete.
	if err = g.Wait(); err == nil {
		fmt.Println("Successfully fetched all URLs.")
	} else {
		fmt.Printf("Failed to fetch urls, error=%v", err)
	}
}

func main() {
	TestErrorGroup()
	TestWaitGroup()
}
