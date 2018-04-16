package goroup

type Goer interface {
	Go()
	Cancel()
	Wait()
	Done() <-chan struct{}

	HasDone() bool // at that moment
}

var _ Goer = Routine{}
var _ Goer = Group{}
