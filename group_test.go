package goroup_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"bitbucket.org/shu/goroup"
	"bitbucket.org/shu/gotwant"
)

func TestGroup(t *testing.T) {
	t.Run("One", func(t *testing.T) {
		var result int64 = 0
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			atomic.AddInt64(&result, int64(gain))
		}

		pg := goroup.NewGroup()
		pg.Add(goroup.Ready(f, 100))

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		g := pg.Go(nil)
		g.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(100))
	})

	t.Run("Multiple", func(t *testing.T) {
		var result int64 = 0
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			time.Sleep(time.Duration(gain) * time.Millisecond)
			atomic.AddInt64(&result, int64(gain))
		}

		pg := goroup.NewGroup()
		pg.Add(goroup.Ready(f, 100))
		pg.Add(goroup.Ready(f, 10))
		pg.Add(goroup.Ready(f, 1))

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		g := pg.Go(nil)

		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > 0)

		g.PurgeDone()
		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > 0)

		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(111))
	})

	t.Run("Cancel", func(t *testing.T) {
		var result int64 = 0
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			time.Sleep(time.Duration(gain) * time.Millisecond)
			atomic.AddInt64(&result, int64(gain))
		}

		pg := goroup.NewGroup()
		pg.Add(goroup.Ready(f, 100))
		pg.Add(goroup.Ready(f, 10))
		pg.Add(goroup.Ready(f, 1))

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		g := pg.Go(nil)

		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > 0)

		g.PurgeDone()
		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > 0)

		g.Cancel()
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(11))
	})
}
