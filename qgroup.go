package goroup

import (
	"context"
	"sync"
)

type QGroup struct {
	*rawQGroup
}

type rawQGroup struct {
	context context.Context
	cancel  context.CancelFunc

	waitm  sync.Mutex
	waited bool

	wg sync.WaitGroup

	queue chan PreRoutine
	sem   chan struct{}

	dispatcher Routine
}

func NewQueuedGroup(ctx context.Context, limit int64) QGroup {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	g := QGroup{
		rawQGroup: &rawQGroup{
			queue:   make(chan PreRoutine),
			sem:     make(chan struct{}, limit),
			context: ctx,
			cancel:  cancel,
		},
	}
	g.dispatcher = Go(g.dispatch, g.context)

	return g
}

func (g QGroup) Add(pr PreRoutine) {
	g.waitm.Lock()
	if g.waited {
		g.waitm.Unlock()
		return
	}
	g.waitm.Unlock()

	g.wg.Add(1)
	g.queue <- pr
}

func (g QGroup) Cancel() {
	g.cancel()
}

func (g QGroup) Wait() {
	g.waitm.Lock()
	g.waited = true
	g.waitm.Unlock()

	g.wg.Wait()
}

func (g QGroup) dispatch(ctx context.Context, params ...interface{}) {
	for {
		select {
		case <-ctx.Done():
			break
		case pr := <-g.queue:
			g.sem <- struct{}{}

			go func(pr PreRoutine) {
				r := pr.Go(ctx)
				r.Wait()

				<-g.sem
				g.wg.Done()
			}(pr)
		}
	}
}
