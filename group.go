package goroup

import (
	"sync"

	"bitbucket.org/shu/clise"
)

// Group is a Group of Routines and Groups.
type Group struct {
	*rawGroup
}

type rawGroup struct {
	m     sync.Mutex
	goers []Goer
}

// Group makes a Group of Routines and Groups.
func NewGroup(goers ...Goer) Group {
	return Group{
		rawGroup: &rawGroup{
			goers: goers,
		},
	}
}

// Add adds a Goer in the Group.
func (g Group) Add(ger Goer) {
	g.m.Lock()
	g.goers = append(g.goers, ger)
	g.m.Unlock()
}

// PurgeDone removes some Goers that are ended or cancelled.
func (g Group) PurgeDone() {
	g.m.Lock()
	clise.Filter(&g.goers, func(i int) bool {
		return !g.goers[i].HasDone()
	})
	g.m.Unlock()
}

// Go starts all Goerss.
func (g Group) Go() {
	g.m.Lock()
	if len(g.goers) == 0 {
		g.m.Unlock()
		return
	}
	goers := make([]Goer, len(g.goers))
	copy(goers, g.goers)
	g.m.Unlock()

	for _, r := range goers {
		r.Go()
	}
}

// Cancel cancels all Goers
func (g Group) Cancel() {
	g.m.Lock()
	if len(g.goers) == 0 {
		g.m.Unlock()
		return
	}
	goers := make([]Goer, len(g.goers))
	copy(goers, g.goers)
	g.m.Unlock()

	for _, r := range goers {
		r.Cancel()
	}
}

// Cancel waits for all Goers end or are cancelled.
func (g Group) Wait() {
	g.m.Lock()
	if len(g.goers) == 0 {
		g.m.Unlock()
		return
	}
	goers := make([]Goer, len(g.goers))
	copy(goers, g.goers)
	g.m.Unlock()

	for _, r := range goers {
		r.Wait()
	}
}

// Cancel waits for a Goer ends or is cancelled.
func (g Group) WaitAny() {
	g.m.Lock()
	if len(g.goers) == 0 {
		g.m.Unlock()
		return
	}
	goers := make([]Goer, len(g.goers))
	copy(goers, g.goers)
	g.m.Unlock()

	anyDoneOnce := sync.Once{}
	anyDoneChan := make(chan struct{})

	for _, r := range goers {
		go func(rr Goer) {
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
	for _, r := range g.goers {
		if !r.HasDone() {
			return false
		}
	}
	return true
}
