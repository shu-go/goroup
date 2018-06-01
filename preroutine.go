package goroup

import "context"

type PreRoutine struct {
	run    GoFunc
	params []interface{}
}

func Ready(run GoFunc, params ...interface{}) PreRoutine {
	return PreRoutine{
		run:    run,
		params: params,
	}
}

func (p PreRoutine) Go(parent context.Context) Routine {
	return Go(p.run, parent, p.params...)
}
