package main

import (
	"context"
	"fmt"
	"time"
)

type Value struct{}


func Stream(ctx context.Context, out chan<- Value) error {
	// Create a derived Context with a 10s timeout; dctx
	// will be canceledd upon timeout, but ctx will not
	// cancel is a function that will explicitly cancel dctx.

	dctx, cancel := context.WithTimeout(ctx, time.Second * 10)
	

	// Release resources if SlowOperation completes before timeout
	defer cancel()


	res, err := SlowOperation(dctx)
	if err != nil { // True if dctx tines out
		return err
	}

	for {
		select {
		case out <- res: // Read form res; send to out


		case <-ctx.Done(): // Triggered if ctx is canncelled
			return ctx.Err()
		}
	}
}

func SlowOperation(ctx context.Context) (Value, error){
	fmt.Println(ctx.Deadline())
	return Value{}, nil
}