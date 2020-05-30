package ctx

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func cancelConsume(ctx context.Context) <-chan int {
	c := make(chan int)
	n := 0
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Got %d, cancel\n", n)
				return
			case c <- n:
				incr := rand.Intn(5)
				n += incr
				if n >= 10 {
					n = 10
				}
				fmt.Printf("Consume %d \n", n)
			}
		}
	}()

	return c
}
func Cancel() {
	ctx, cancel := context.WithCancel(context.Background())
	num := cancelConsume(ctx)
	for n := range num {
		if n >= 10 {
			cancel()
			time.Sleep(1 * time.Second)
			break
		}
	}
}
