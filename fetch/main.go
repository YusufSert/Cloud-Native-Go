package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Promise struct {
	once sync.Once
	wg   sync.WaitGroup

	res *http.Response
	err error

	resCh <-chan *http.Response
	errCh <-chan error
}

func (p *Promise) Result() (*http.Response, error) {
	p.once.Do(func() {
		p.wg.Add(1)
		defer p.wg.Done()
		p.res = <-p.resCh
		p.err = <-p.errCh
	})
	p.wg.Wait()

	return p.res, p.err
}

func (p *Promise) Then(f func(*http.Response)) *Promise {
	res, _ := p.Result()
	f(res)
	return p
}

func Fetch(addr string) Promise {
	resCh := make(chan *http.Response)
	errCh := make(chan error)
	go func() {
		defer close(resCh)
		defer close(errCh)
		time.Sleep(1 * time.Second)
		res, err := http.Get(addr)
		resCh <- res
		errCh <- err
	}()

	return Promise{resCh: resCh, errCh: errCh}
}

func main() {
	p := Fetch("https://www.google.com")
	p.Then(func(r *http.Response) {
		fmt.Println(r.Status)
	})
}
