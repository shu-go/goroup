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

// Cancelled is a check function if a Routine is requested to cancel.
type Cancelled func() bool

// Ready makes a Routine with f goroutinized.
//
// This function does not call f().
// Call (Routine).Go() as your needs.
// (Other methods like Done() and Wait() would not Go() internally)
//
// NOTE: If the function  f have loop, you should check if you requested to cancel, by calling c().
// Example:
// r := goroup.Routine(func(c Cancelled) {
//     for {
//	       if c() {
//	           return
//	       }
//
//	       // do something
//	   }
// })
//
// Example:
// r := goroup.Routine(func(c Cancelled) {
//	   // :
// })
// r.Go() // starts a goroutine
// r.Wait() // wait for the goroutine end
// r.Cancel() // cancel the goroutine
// <-r.Done() // wait for the goroutine end
func Ready(f func(c Cancelled)) Routine {
	return Routine{
		run: f,
		id:  fmt.Sprintf("%v", time.Now().String()),
	}
}

// Routine holds a goroutinized function.
type Routine struct {
	id string

	doneMut  sync.Mutex
	doneChan chan struct{}
	doneOnce sync.Once

	run func(Cancelled)
}

func (r *Routine) markDone() {
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
func (r *Routine) Go() {
	if r.run == nil {
		return
	}

	r.doneMut.Lock()
	if r.doneChan != nil {
		select {
		case <-r.doneChan:
			// can restart
		default:
			// running
			r.doneMut.Unlock()
			return
		}
	}
	r.doneChan = make(chan struct{})
	r.doneOnce = sync.Once{}
	r.doneMut.Unlock()

	c := func() bool {
		select {
		case <-r.doneChan:
			return true
		default:
		}
		return false
	}

	go func() {
		r.run(c)
		r.markDone()
	}()
}

// Wait waits the goroutine ends or is cancelled.
func (r *Routine) Wait() {
	r.doneMut.Lock()
	if r.doneChan == nil {
		r.doneMut.Unlock()
		return
	}
	r.doneMut.Unlock()

	<-r.doneChan
}

// Cancel cancels the goroutine.
func (r *Routine) Cancel() {
	r.markDone()
}

// Done returns receive-only chan that is sent when the goroutine is done.
func (r *Routine) Done() <-chan struct{} {
	return Done(r.Wait)
}

// Goroup is a group of Routines.
type Goroup struct {
	m        sync.Mutex
	routines []*Routine
}

// Group makes a group of Routines.
func Group(routines ...*Routine) Goroup {
	return Goroup{
		routines: routines,
	}
}

// Add adds a routine in the group.
func (g *Goroup) Add(r *Routine) {
	g.m.Lock()
	g.routines = append(g.routines, r)
	g.m.Unlock()
}

// PurgeDone removes some Routines that are ended or cancelled.
func (g *Goroup) PurgeDone() {
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
func (g *Goroup) Go() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Go()
	}
}

// Cancel cancels all Routines.
func (g *Goroup) Cancel() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Cancel()
	}
}

// Cancel waits for all Routines end or are cancelled.
func (g *Goroup) Wait() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Wait()
	}
}

// Cancel waits for a Routine ends or is cancelled.
func (g *Goroup) WaitAny() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]*Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	anyDoneOnce := sync.Once{}
	anyDoneChan := make(chan struct{})

	for _, r := range routines {
		go func(rr *Routine) {
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
func (g *Goroup) Done() <-chan struct{} {
	return Done(g.Wait)
}

// Done returns receive-only chan that is sent when a goroutine is done.
func (g *Goroup) DoneAny() <-chan struct{} {
	return Done(g.WaitAny)
}
