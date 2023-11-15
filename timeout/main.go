package timeout

import (
	"context"
	"errors"
	"time"
)

type Result struct {
	Response string
	Error    error
}
type SlowFunction func(string) (string, error)

type WithContext func(context.Context, string) (string, error)

func Timeout(f SlowFunction) WithContext {
	return func(ctx context.Context, args string) (string, error) {
		chres := make(chan Result)

		go func() {
			close(chres)
			res, err := f(args)
			chres <- Result{res, err}
		}()

		select {
		case res := <-chres:
			return res.Response, res.Error
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func TimeoutAfterTime(f SlowFunction) func(arg string) (string, error) {
	return func(arg string) (string, error) {
		chres := make(chan string)
		cherr := make(chan error)

		go func() {
			res, err := f(arg)
			chres <- res
			cherr <- err
		}()

		select {
		case res := <-chres:
			return res, <-cherr
		case <-time.After(time.Second * 10):
			return "", errors.New("time out")
		}
	}
}
