package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Circuit func(ctx context.Context) (string, error)

func DebounceLast(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var ticker *time.Ticker
	var result string
	var err error
	var once sync.Once
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		fmt.Println("lock")
		defer m.Unlock()

		threshold = time.Now().Add(d)

		once.Do(func() {
			ticker = time.NewTicker(time.Millisecond * 100)

			go func() {
				defer func() {
					m.Lock()
					ticker.Stop()
					once = sync.Once{}
					m.Unlock()
				}()
				for {
					select {
					case <-ticker.C:
						fmt.Println("ticker")
						m.Lock()
						if time.Now().After(threshold) {
							fmt.Println("debounce")
							result, err = circuit(ctx)
							m.Unlock()
							return
						}
						m.Unlock()
					case <-ctx.Done():
						m.Lock()
						result, err = "", ctx.Err()
						m.Unlock()
						return
					}
				}
			}()
		})
		return result, err
	}
}

func main() {
	ctx := context.WithValue(context.Background(), "name", "Kudiş")
	ctx2 := context.WithValue(context.Background(), "name", "Yusuf")
	f := func(ctx context.Context) (string, error) {
		return ctx.Value("name").(string), nil
	}

	f = DebounceLast(f, 2*time.Second)

	fmt.Println(f(ctx))
	time.Sleep(1 * time.Second)
	fmt.Println(f(ctx2))

	time.Sleep(5 * time.Second)

}
