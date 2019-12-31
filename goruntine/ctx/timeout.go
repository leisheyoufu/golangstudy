package ctx

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func Timeout() {
	// ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	timeConsume(ctx)
	defer cancel()
}

func timeConsume(ctx context.Context) {
	n := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Stop timeConsumer due to timeout event\n")
			return
		default:
			incr := rand.Intn(5)
			n += incr
			fmt.Printf("Consume %d\n", n)
		}
		time.Sleep(time.Second)
	}
}
