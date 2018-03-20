package gorou

import (
	"fmt"
	"sync"
	"time"

	"bitbucket.org/shu/clise"
)

func Do(f func()) <-chan struct{} {
	doneChan := make(chan struct{})
	go func() {
		f()
		doneChan <- struct{}{}
	}()
	return doneChan
}

type Cancelled func() bool

func Routine(f func(Cancelled)) routine {
	return routine{
		run: f,
		id:  fmt.Sprintf("%v", time.Now().String()),
	}
}

type routine struct {
	id string

	doneMut  sync.Mutex
	doneChan chan struct{}
	doneOnce sync.Once

	run func(Cancelled)
}

func (j *routine) markDone() {
	j.doneMut.Lock()
	j.doneOnce.Do(func() {
		if j.doneChan != nil {
			close(j.doneChan)
		}
	})
	j.doneMut.Unlock()
}

func (j *routine) Run() {
	j.doneMut.Lock()
	if j.doneChan != nil {
		select {
		case <-j.doneChan:
			// can restart
		default:
			// running
			j.doneMut.Unlock()
			return
		}
	}
	j.doneChan = make(chan struct{})
	j.doneOnce = sync.Once{}
	j.doneMut.Unlock()

	c := func() bool {
		select {
		case <-j.doneChan:
			return true
		default:
		}
		return false
	}

	go func() {
		j.run(c)
		j.markDone()
	}()
}

func (j *routine) Wait() {
	j.doneMut.Lock()
	if j.doneChan == nil {
		j.doneMut.Unlock()
		return
	}
	j.doneMut.Unlock()

	<-j.doneChan
}

func (j *routine) Cancel() {
	j.markDone()
}

type group struct {
	m        sync.Mutex
	routines []*routine
}

func Group(routines ...*routine) group {
	return group{
		routines: routines,
	}
}

func (g *group) Add(j *routine) {
	g.m.Lock()
	g.routines = append(g.routines, j)
	g.m.Unlock()
}

func (g *group) PurgeDone() {
	g.m.Lock()
	clise.Filter(&g.routines, func(i int) bool {
		select {
		case <-g.routines[i].doneChan:
			return false
		default:
		}
		return true
	})
	g.m.Unlock()
}

func (g *group) Run() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, j := range routines {
		j.Run()
	}
}

func (g *group) Cancel() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, j := range routines {
		j.Cancel()
	}
}

func (g *group) Wait() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, j := range routines {
		j.Wait()
	}
}

func (g *group) WaitAny() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	anyDoneOnce := sync.Once{}
	anyDoneChan := make(chan struct{})

	for _, j := range routines {
		go func(j *routine) {
			select {
			case <-j.doneChan:
				anyDoneOnce.Do(func() {
					close(anyDoneChan)
				})
			case <-anyDoneChan:
			}
		}(j)
	}

	<-anyDoneChan
}
