package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Circuit func(ctx context.Context) (string, error)

func DebounceLast(circuit Circuit, d time.Duration) Circuit {
	var timer *time.Timer = time.NewTimer(d)
	var result string
	var err error
	var once sync.Once
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		timer.Reset(d)

		once.Do(func() {
			var wg sync.WaitGroup
			wg.Add(1)
			timer = time.NewTimer(d)

			go func() {
				wg.Done()
				fmt.Println("goroutine")
				defer func() {
					m.Lock()
					timer.Stop()
					once = sync.Once{}
					m.Unlock()
				}()

				for {
					select {
					case <-timer.C:
						m.Lock()
						result, err = circuit(ctx)
						m.Unlock()
						return
					case <-ctx.Done():
						m.Lock()
						result, err = "", ctx.Err()
						m.Unlock()
						return
					}
				}
			}()
			wg.Wait()
		})
		return result, err
	}
}

func main() {

	ctx := context.WithValue(context.Background(), "name", "Kudiş")
	//ctx2 := context.WithValue(context.Background(), "name", "Yusuf")
	f := func(ctx context.Context) (string, error) {
		return ctx.Value("name").(string), nil
	}

	f = DebounceLast(f, 1*time.Nanosecond)

	var wg sync.WaitGroup
	wg.Add(200)

	for i := 0; i < 100; i++ {
		go f(ctx)
	}

}
