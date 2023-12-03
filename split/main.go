package main

import (
	"fmt"
	"sync"
)

func Split(source <-chan int, n int) []<-chan int {
	dest := make([]<-chan int, 0) // Create the dests slice

	for ; n > 0; n-- { // Create n destination channels
		ch := make(chan int)
		dest = append(dest, ch)

		go func() { // Each channel gets a dedicated goroutine that competes for reads
			defer close(ch)

			for v := range source {
				ch <- v
			}
		}()
	}

	return dest
}

func main() {
	source := make(chan int, 10) // The input channel
	dests := Split(source, 5)    // Retrieve 5 output channels

	go func() { // Send the number 1..10 to source and close it when we're done
		defer close(source)
		for i := 1; i < 10; i++ {
			source <- i
		}
	}()

	var wg sync.WaitGroup // Use WaitGroup to wait until the output channels all close
	wg.Add(len(dests))

	worker := func(i int, ch <-chan int) {
		defer wg.Done()
		for val := range ch {
			fmt.Printf("#%d got %d\n", i, val)
		}
	}
	for i, ch := range dests {
		/*
			go func(i int, d <-chan int) {
				defer wg.Done()

				for val := range d {
					fmt.Printf("#%d got %d\n", i, val)
				}
			}(i, ch)
		*/
		go worker(i, ch)
	}

	wg.Wait()

}
