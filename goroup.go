package goroup

import (
	"fmt"
	"sync"
	"time"

	"bitbucket.org/shu/clise"
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

// Cancelled is a chan having checking function if a Routine is requested to cancel.
type Cancelled chan struct{}

func (c Cancelled) Cancelled() bool {
	select {
	case <-c:
		return true
	default:
	}
	return false
}

// Ready makes a Routine with f goroutinized.
//
// This function does not call f().
// Call (Routine).Go() as your needs.
// (Other methods like Done() and Wait() would not Go() internally)
//
// NOTE: If the function  f have loop, you should check if you requested to cancel, by calling c().
// Example:
// r := goroup.Ready(func(c Cancelled) {
//     for {
//	       if c.Cancelled() {
//	           return
//	       }
//
//	       // do something
//	   }
// })
//
// Example:
// r := goroup.Ready(func(c Cancelled) {
//	   // :
// })
// r.Go() // starts a goroutine
// r.Wait() // wait for the goroutine end
// r.Cancel() // cancel the goroutine
// <-r.Done() // wait for the goroutine end
func Ready(f func(c Cancelled)) routine {
	return routine{
		run: f,
		id:  fmt.Sprintf("%v", time.Now().String()),
	}
}

// routine holds a goroutinized function.
//
// It has 3 states:
//     Done : function is nil, function is end, function is cancelled
//     Not Going : routine is not Go(). This also be treated as done. (routine).Wait() returns immediately.
//                 To make sure going, call Go().
//     Going : routine is Go()ing. This is not done.
type routine struct {
	id string

	doneMut  sync.Mutex
	doneChan chan struct{}
	doneOnce sync.Once

	run func(Cancelled)
}

func (r *routine) markDone() {
	r.doneMut.Lock()
	r.doneOnce.Do(func() {
		if r.doneChan != nil {
			close(r.doneChan)
		}
	})
	r.doneMut.Unlock()
}

// Go starts the function as a goroutine.
//
// The function f passed in Ready(f) is called as a goroutine.
// Multiple call of Go() while its running is ignored.
// Multiple call of Go() after end or cancel is also ignored.
func (r *routine) Go() {
	if r.run == nil {
		return
	}

	r.doneMut.Lock()
	if r.doneChan != nil {
		r.doneMut.Unlock()
		return
	}
	r.doneChan = make(chan struct{})
	r.doneOnce = sync.Once{}
	r.doneMut.Unlock()

	c := Cancelled(r.doneChan)

	go func() {
		r.run(c)
		r.markDone()
	}()
}

// Wait waits the goroutine ends or is cancelled.
func (r *routine) Wait() {
	r.doneMut.Lock()
	if r.doneChan == nil {
		r.doneMut.Unlock()
		return
	}
	r.doneMut.Unlock()

	<-r.doneChan
}

// Cancel cancels the goroutine.
func (r *routine) Cancel() {
	r.markDone()
}

// Done returns receive-only chan that is sent when the goroutine is done.
func (r *routine) Done() <-chan struct{} {
	return Done(r.Wait)
}

// group is a group of Routines.
type group struct {
	m        sync.Mutex
	routines []*routine
}

// Group makes a group of Routines.
func Group(routines ...*routine) group {
	return group{
		routines: routines,
	}
}

// Add adds a routine in the group.
func (g *group) Add(r *routine) {
	g.m.Lock()
	g.routines = append(g.routines, r)
	g.m.Unlock()
}

// PurgeDone removes some Routines that are ended or cancelled.
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

// Go starts all Routines.
func (g *group) Go() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Go()
	}
}

// Cancel cancels all Routines.
func (g *group) Cancel() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Cancel()
	}
}

// Cancel waits for all Routines end or are cancelled.
func (g *group) Wait() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Wait()
	}
}

// Cancel waits for a Routine ends or is cancelled.
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

	for _, r := range routines {
		go func(rr *routine) {
			select {
			case <-rr.Done():
				anyDoneOnce.Do(func() {
					close(anyDoneChan)
				})
			case <-anyDoneChan:
			}
		}(r)
	}

	<-anyDoneChan
}

// Done returns receive-only chan that is sent when all goroutines are done.
func (g *group) Done() <-chan struct{} {
	return Done(g.Wait)
}

// Done returns receive-only chan that is sent when a goroutine is done.
func (g *group) DoneAny() <-chan struct{} {
	return Done(g.WaitAny)
}
