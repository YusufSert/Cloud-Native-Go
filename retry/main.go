package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"
)

type Effector func(ctx context.Context) (string, error)

func Retry(e Effector, retries uint, d time.Duration) Effector {
	return func(ctx context.Context) (string, error) {
		for r := 0; ; r++ {
			response, err := e(ctx)
			if r >= int(retries) || err == nil {
				return response, err
			}

			log.Printf("Attempt %d failed; retrying in %v", r+1, d)

			select {
			case <-time.After(d):
				d = d << 1
			case <-ctx.Done():
				return "", ctx.Err()
			}

		}
	}
}

func main() {
	var count int

	EmulateTransientError := func(ctx context.Context) (string, error) {
		count++

		if count <= 3 {
			return "intentional fail", errors.New("error")
		}
		return "success", nil
	}

	r := Retry(EmulateTransientError, 5, 500*time.Millisecond)

	res, err := r(context.Background())
	fmt.Println(res, err)

	f := func(i int) func() int {
		return func() int {
			i++
			return i
		}
	}

	i := f(4)
	fmt.Println(i())
	fmt.Println(i())

}
