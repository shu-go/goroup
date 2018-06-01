package goroup

import (
	"context"
)

type Routine struct {
	*rawRoutine
}

type GoFunc func(context.Context, ...interface{})

type rawRoutine struct {
	run     GoFunc
	context context.Context
	cancel  context.CancelFunc
}

func Go(run GoFunc, parent context.Context, params ...interface{}) Routine {
	r := Routine{
		rawRoutine: &rawRoutine{
			run: run,
		},
	}

	if parent == nil {
		parent = context.Background()
	}
	r.context, r.cancel = context.WithCancel(parent)

	go func() {
		r.run(r.context, params...)
		r.cancel()
	}()

	return r
}

func (r Routine) Wait() {
	if r.rawRoutine == nil {
		return
	}

	<-r.context.Done()
}

func (r Routine) Done() <-chan struct{} {
	return r.context.Done()
}

func (r Routine) HasDone() bool {
	select {
	case <-r.context.Done():
		return true
	default:
	}
	return false
}

func (r Routine) Cancel() {
	if r.rawRoutine == nil {
		return
	}

	r.cancel()
}
