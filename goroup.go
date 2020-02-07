package goroup

import (
	"context"
	"runtime"
	"sync"
)

//type GoroupFunc func(context.Context, ...interface{})
type GoroupFunc func(context.Context)
type Goroup struct {
	funcs chan GoroupFunc
	wg    sync.WaitGroup
	maxG  uint
	maxF  uint

	anyDoneCtx context.Context
	anyDone    context.CancelFunc
	//anyDoneChan chan<- struct{}
	//anyWaitChan <-chan struct{}

	context context.Context
	cancel  context.CancelFunc
}

func New(opts ...GoroupOption) *Goroup {
	g := &Goroup{
		maxG: uint(runtime.NumCPU()),
		maxF: 100,
	}
	g.funcs = make(chan GoroupFunc, g.maxF)

	//g.anyDoneChan, g.anyWaitChan = goroutine.DoneWaitChan()
	g.anyDoneCtx, g.anyDone = context.WithCancel(context.Background())

	for _, o := range opts {
		o(g)
	}

	if g.context == nil {
		g.context = context.Background()
	}
	g.context, g.cancel = context.WithCancel(g.context)

	return g
}

func (g *Goroup) Add(f GoroupFunc) {
	select {
	case g.funcs <- f:
		g.wg.Add(1)
	default:
		panic("full")
	}
}

func (g *Goroup) Go() {
	sem := make(chan struct{}, g.maxG)

	go func() {
		for f := range g.funcs {
			sem <- struct{}{}
			go func(f GoroupFunc) {
				f(g.context)

				g.wg.Done()
				<-sem

				//select {
				//case g.anyDoneChan <- struct{}{}:
				//default:
				//}
				g.anyDone()
			}(f)
		}
	}()
}

func (g *Goroup) Wait() {
	g.wg.Wait()
	//close(g.funcs)
}

func (g *Goroup) WaitAny() {
	//<-g.anyWaitChan
	<-g.anyDoneCtx.Done()
}

func (g *Goroup) Cancel() {
	g.cancel()
}

////////////////////////////////////////////////////////////////////////////////

type GoroupOption func(*Goroup)

func MaxFuncs(n uint) GoroupOption {
	return func(g *Goroup) {
		g.maxF = n
	}
}

func MaxGoroutines(n uint) GoroupOption {
	return func(g *Goroup) {
		g.maxG = n
	}
}

func Context(c context.Context) GoroupOption {
	return func(g *Goroup) {
		g.context = c
	}
}
