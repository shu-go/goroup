package goroup

import "context"

// Done returns receive-only chan that is sent when f() is done.
func Done(f func()) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		f()
		close(doneChan)
	}()
	return doneChan
}

func ContextDone(c context.Context) bool {
	select {
	case <-c.Done():
		return true
	default:
		return false
	}
}
