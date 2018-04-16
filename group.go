package goroup

import (
	"sync"

	"bitbucket.org/shu/clise"
)

// Group is a Group of Routines.
type Group struct {
	*rawGroup
}

type rawGroup struct {
	m        sync.Mutex
	routines []Routine
}

// Group makes a Group of Routines.
func NewGroup(routines ...Routine) Group {
	return Group{
		rawGroup: &rawGroup{
			routines: routines,
		},
	}
}

// Add adds a routine in the Group.
func (g *Group) Add(r Routine) {
	g.m.Lock()
	g.routines = append(g.routines, r)
	g.m.Unlock()
}

// PurgeDone removes some Routines that are ended or cancelled.
func (g *Group) PurgeDone() {
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
func (g *Group) Go() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Go()
	}
}

// Cancel cancels all Routines.
func (g *Group) Cancel() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Cancel()
	}
}

// Cancel waits for all Routines end or are cancelled.
func (g *Group) Wait() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	for _, r := range routines {
		r.Wait()
	}
}

// Cancel waits for a Routine ends or is cancelled.
func (g *Group) WaitAny() {
	g.m.Lock()
	if len(g.routines) == 0 {
		g.m.Unlock()
		return
	}
	routines := make([]Routine, len(g.routines))
	copy(routines, g.routines)
	g.m.Unlock()

	anyDoneOnce := sync.Once{}
	anyDoneChan := make(chan struct{})

	for _, r := range routines {
		go func(rr Routine) {
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
func (g *Group) Done() <-chan struct{} {
	return Done(g.Wait)
}

// Done returns receive-only chan that is sent when a goroutine is done.
func (g *Group) DoneAny() <-chan struct{} {
	return Done(g.WaitAny)
}
