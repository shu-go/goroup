package goroup

import (
	"context"
	"sync"

	"bitbucket.org/shu_go/clise"
)

// Group is a Group of Routines and Groups.
type Group struct {
	*rawGroup
}

type rawGroup struct {
	m        sync.Mutex
	routines []Routine

	ctx    context.Context
	cancel context.CancelFunc
}

// PurgeDone removes some Routines that are ended or cancelled.
func (g Group) PurgeDone() {
	g.m.Lock()
	clise.Filter(&g.routines, func(i int) bool {
		return !g.routines[i].HasDone()
	})
	g.m.Unlock()
}

// Cancel cancels all Routines
func (g Group) Cancel() {
	g.cancel()
}

// Cancel waits for all Routines end or are cancelled.
func (g Group) Wait() {
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
func (g Group) WaitAny() {
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
func (g Group) Done() <-chan struct{} {
	return Done(g.Wait)
}

// Done returns receive-only chan that is sent when a goroutine is done.
func (g Group) DoneAny() <-chan struct{} {
	return Done(g.WaitAny)
}

// HasDone returns immediate state
func (g Group) HasDone() bool {
	g.m.Lock()
	defer g.m.Unlock()
	for _, r := range g.routines {
		if !r.HasDone() {
			return false
		}
	}
	return true
}
