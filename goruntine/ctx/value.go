package ctx

import (
	"context"
	"fmt"
)

func WithValue() {
	ctx := context.WithValue(context.Background(), "trace_id", "88888888")
	ctx = context.WithValue(ctx, "session", 1)

	process(ctx)
}

func process(ctx context.Context) {
	session, ok := ctx.Value("session").(int)
	if !ok {
		fmt.Println("Session is not found")
		return
	}

	if session != 1 {
		fmt.Println("Session is not correct")
		return
	}

	traceID := ctx.Value("trace_id").(string)
	fmt.Println("traceID:", traceID, "session:", session)
	// emptyCtx will return nil in Value method
	no_value := ctx.Value("no_value")
	if no_value != nil {
		fmt.Println("No value exist, wrong!")
	}
}
