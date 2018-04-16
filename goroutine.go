package goroup

import (
	"fmt"
	"reflect"
)

// Done returns receive-only chan that is sent when f() is done.
func Done(f func()) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		f()
		doneChan <- struct{}{}
	}()
	return doneChan
}

func Select(chans ...interface{}) (c interface{}, value interface{}, recvOK bool) {
	if len(chans) == 0 {
		return nil, nil, false
	}

	var cases []reflect.SelectCase
	for i, c := range chans {
		cv := reflect.ValueOf(c)
		if cv.Kind() != reflect.Chan || cv.Type().ChanDir()&reflect.RecvDir == 0 {
			panic(fmt.Sprintf("%dth element is not a receivable chan (%v)", i, cv.Type().String()))
		}

		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: cv,
		})
	}

	chosen, recv, recvOK := reflect.Select(cases)
	if !recvOK {
		return nil, nil, false
	}

	return chans[chosen], recv.Interface(), true
}

func TrySelect(chans ...interface{}) (c interface{}, value interface{}, recvOK bool) {
	if len(chans) == 0 {
		return nil, nil, false
	}

	for i, c := range chans {
		cv := reflect.ValueOf(c)
		if cv.Kind() != reflect.Chan || cv.Type().ChanDir()&reflect.RecvDir == 0 {
			panic(fmt.Sprintf("%dth element is not a receivable chan (%v)", i, cv.Type().String()))
		}

		if x, ok := cv.TryRecv(); ok {
			return c, x.Interface(), true
		} else if x == reflect.Zero(cv.Type().Elem()) {
			// closed
			return c, nil, false
		}
	}

	return nil, nil, false
}
