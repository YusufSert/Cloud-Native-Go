package main

import (
	"fmt"
	"sync"
	"time"
)

type Promise struct {
	once sync.Once
	wg   sync.WaitGroup

	ok    any
	wrong any

	chOk    chan any
	chWrong chan any
}

func (p *Promise) Ok(v any) {
	p.chWrong <- v
}
func (p *Promise) Wrong(v any) {
	p.chOk <- v
}

func (p *Promise) Then(f func(any)) *Promise {
	p.once.Do(func() {
		p.wg.Add(1)
		defer p.wg.Done()
		p.ok = <-p.chOk
		p.wrong = <-p.chWrong
	})
	p.wg.Wait()
	f(p.ok)
	return p
}

func NewPromise(f func(ok func(any), wrong func(any))) *Promise {
	chOk := make(chan any)
	chWrong := make(chan any)

	p := new(Promise)
	go func() {
		defer close(chOk)
		defer close(chWrong)
		f(p.Ok, p.Wrong)
	}()
	p.chOk = chOk
	p.chWrong = chWrong
	return p
}

func main() {
	p := NewPromise(func(ok func(any), wrong func(any)) {
		//Blocking-function

		ok("ok")
		wrong("error")
	})
	time.Sleep(1 * time.Second)
	fmt.Println(p.ok, p.wrong)

	b := struct {
		money int
	}{money: 0}

	var wg sync.WaitGroup
	c := make(chan int)

	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		go func() {
			defer wg.Done()
			c <- 1
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()

	for v := range c {
		b.money += v
	}
	fmt.Println(b.money)

}
