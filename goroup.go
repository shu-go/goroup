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

	anyM       sync.Mutex
	anyDoneCtx context.Context
	anyDone    context.CancelFunc

	context context.Context
	cancel  context.CancelFunc
}

func New(opts ...GoroupOption) *Goroup {
	g := &Goroup{
		maxG: uint(runtime.NumCPU()),
		maxF: 100,
	}
	g.funcs = make(chan GoroupFunc, g.maxF)

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

	g.anyM.Lock()
	g.anyDoneCtx, g.anyDone = context.WithCancel(context.Background())
	g.anyM.Unlock()

	go func() {
		for f := range g.funcs {
			sem <- struct{}{}
			go func(f GoroupFunc) {
				f(g.context)

				g.wg.Done()
				<-sem

				g.anyM.Lock()
				g.anyDone()
				g.anyM.Unlock()
			}(f)
		}
	}()
}

func (g *Goroup) Wait() {
	g.wg.Wait()
}

func (g *Goroup) WaitAny() {
	g.anyM.Lock()
    done := g.anyDoneCtx.Done()
	g.anyM.Unlock()

	<-done
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
