package goroup

import (
	"context"
	"sync"
)

type PreGroup struct {
	*rawPreGroup
}

type rawPreGroup struct {
	m    sync.Mutex
	pres []PreRoutine
}

func NewGroup() PreGroup {
	return PreGroup{
		rawPreGroup: &rawPreGroup{},
	}
}

func (p PreGroup) Add(pr PreRoutine) {
	p.m.Lock()
	defer p.m.Unlock()

	p.pres = append(p.pres, pr)
}

func (p PreGroup) Go(ctx context.Context) Group {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	p.m.Lock()
	var pres []PreRoutine = make([]PreRoutine, len(p.pres))
	copy(pres, p.pres)
	p.m.Unlock()

	var routines []Routine
	for _, pr := range pres {
		routines = append(routines, pr.Go(ctx))
	}

	return Group{
		rawGroup: &rawGroup{
			routines: routines,
			ctx:      ctx,
			cancel:   cancel,
		},
	}
}
