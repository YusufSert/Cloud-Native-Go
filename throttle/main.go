package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Effector func(ctx context.Context) (string, error)

func Throttle(e Effector, max uint, refill uint, d time.Duration) Effector {
	var tokens = max
	var once sync.Once

	return func(ctx context.Context) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		once.Do(func() {
			ticker := time.NewTicker(d)

			fmt.Println("once")
			go func() {
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return

					case <-ticker.C:
						t := tokens + refill
						if t > max {
							t = max
						}
						tokens = t
					}
				}
			}()
		})

		if tokens <= 0 {
			return "", fmt.Errorf("too many calls")
		}

		tokens--
		return e(ctx)
	}
}

func main() {
	e := func(ctx context.Context) (string, error) {
		return ctx.Value("name").(string), nil
	}
	t := Throttle(e, 50, 50, 1*time.Second)

	ctx := context.WithValue(context.Background(), "name", "Kudiş")
	fmt.Println(ctx, t)
}
