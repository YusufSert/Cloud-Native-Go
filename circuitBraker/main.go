package circuitBraker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	var consecutiveFailures int = 0
	var lastAttempt = time.Now()
	var m sync.RWMutex

	return func(ctx context.Context) (string, error) {
		// use readLock when there is only reading the shared mem addr
		m.RLock()

		d := consecutiveFailures - int(failureThreshold)

		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d)
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unreachable")
			}
		}

		m.RUnlock()

		response, err := circuit(ctx)

		// use lock while there can be writing the shared mem addr
		m.Lock()
		defer m.Unlock()
		//

		lastAttempt = time.Now() //access to shared mem addr

		if err != nil {
			consecutiveFailures++ // access to shared mem addr
			return response, err
		}

		consecutiveFailures = 0 // access to shared mem addr

		//
		return response, nil
	}
}

func test(wg *sync.WaitGroup) func() int {
	var i int = 0
	var m sync.RWMutex

	//closure
	return func() int {
		defer wg.Done()
		// read
		m.RLock()
		fmt.Println(i)
		m.RUnlock()
		//

		//write
		m.Lock()
		i++
		m.Unlock()
		//
		return i
	}
}

func main() {
	var wg sync.WaitGroup
	inc := test(&wg)

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go inc()
	}
	wg.Wait()
}
