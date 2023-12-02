package main

import (
	"fmt"
	"sync"
	"time"
)

func Funnel(sources ...<-chan int) <-chan int {
	dest := make(chan int) // The shared output channel

	var wg sync.WaitGroup // Used to automatically close dest  When all sources are closed

	wg.Add(len(sources))

	for _, ch := range sources { // Start a goroutine for each source
		go func(ch <-chan int) {
			defer wg.Done() // Notify WaitGroup when c closes

			for v := range ch {
				dest <- v
			}
		}(ch)
	}

	go func() { // Start a goroutine to close dest after all sources close
		wg.Wait()
		close(dest)
	}()

	return dest
}

func main() {
	sources := make([]<-chan int, 0) // Create an empty channel slice

	for i := 0; i < 3; i++ {
		ch := make(chan int)
		sources = append(sources, ch) // Create a channel; add o sources

		go func() {
			defer close(ch)

			for i := 1; i <= 5; i++ {
				ch <- i
				time.Sleep(time.Second)
			}
		}()
	}

	dest := Funnel(sources...)
	for d := range dest {
		fmt.Println(d)
	}
}

// Fan-In
func fanIn(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	multiplexStream := make(chan interface{})

	multiplex := func(c <-chan interface{}) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case multiplexStream <- i:
			}
		}
	}

	// Select form all the channels
	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	// Wait for all the reads to complete
	go func() {
		wg.Wait()
		close(multiplexStream)
	}()

	return multiplexStream
}
