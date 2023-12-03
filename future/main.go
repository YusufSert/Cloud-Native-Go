package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Future interface {
	Result() (string, error)
	Then(func(string, error)) Future
}

type InnerFuture struct {
	once sync.Once
	wg   sync.WaitGroup

	res   string
	err   error
	resCh <-chan string
	errCh <-chan error
}

func (f *InnerFuture) Result() (string, error) {
	f.once.Do(func() {
		f.wg.Add(1)
		defer f.wg.Done()
		f.res = <-f.resCh
		f.err = <-f.errCh
	})
	f.wg.Wait()

	return f.res, f.err
}

func (f *InnerFuture) Then(fu func(res string, err error)) Future {
	res, err := f.Result()
	fu(res, err)
	return f
}

func SlowFunction(ctx context.Context) Future {
	resCh := make(chan string)
	errCh := make(chan error)

	go func() {
		select {
		case <-time.After(time.Second * 2):
			resCh <- "I slept for 2 seconds"
			errCh <- nil
		case <-ctx.Done():
			resCh <- ""
			errCh <- ctx.Err()
		}
	}()

	return &InnerFuture{resCh: resCh, errCh: errCh}
}

func main() {
	ctx := context.Background()
	future := SlowFunction(ctx)

	future.Then(func(data string, err error) {
		if err != nil {
			fmt.Println("error", err)
			return
		}
		fmt.Println(data)
	})

	/*
		res, err := future.Result()
		if err != nil {
			fmt.Println("error", err)
			return
		}
		fmt.Println(res)

	*/

}

// examples

type Matrix int

func BlockingInverse(m Matrix) Matrix {
	time.Sleep(2 * time.Second)
	return Matrix(0)
}

func ConcurrentInverse(m Matrix) <-chan Matrix {
	out := make(chan Matrix)

	go func() {
		out <- BlockingInverse(m)
		close(out)
	}()

	return out
}

func Product(a, b Matrix) Matrix {
	return Matrix(a * b)
}

func InverseProduct(a, b Matrix) Matrix {
	inva := ConcurrentInverse(a)
	invb := ConcurrentInverse(b)

	return Product(<-inva, <-invb)
}
