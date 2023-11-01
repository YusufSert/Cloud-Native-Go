package main

import (
	"fmt"
	"sync"
	"time"
)

type Circuit func(s string) (string, error)

func DebounceFirst(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var result string
	var err error
	var mu sync.Mutex

	return func(s string) (string, error) {
		mu.Lock()

		defer func() {
			threshold = time.Now().Add(d)
			mu.Unlock()
		}()

		if time.Now().Before(threshold) {
			return result, err
		}
		result, err = circuit(s)

		return result, err
	}
}

func main() {
	var wg sync.WaitGroup

	t := func(s string) (string, error) {
		return s, nil
	}

	t = DebounceFirst(t, 1*time.Nanosecond)

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(i int) {
			defer wg.Done()
			response, _ := t(fmt.Sprintf("%s\t%d", "debounce_first", i))
			fmt.Println(response)
		}(i)
	}

	wg.Wait()
}
