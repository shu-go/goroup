package goroup_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shu-go/goroup"
	"github.com/shu-go/gotwant"
)

func TestRoutine(t *testing.T) {
	t.Run("BasicUsage", func(t *testing.T) {
		var result int64 = 0
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			atomic.AddInt64(&result, int64(gain))
		}

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		r := goroup.Go(f, nil, 2)

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		r.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
	})

	t.Run("PreRoutine", func(t *testing.T) {
		var result int64 = 0
		pr := goroup.Ready(func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			atomic.AddInt64(&result, int64(gain))
		}, 1)

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		r1 := pr.Go(context.TODO())
		r2 := pr.Go(context.TODO())

		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(0))

		r1.Wait()
		r2.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
	})

	t.Run("Return", func(t *testing.T) {
		var result int64 = 0
		var anotherResult int64 = 0
		pr := goroup.Ready(func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			another := params[1].(*int64)
			atomic.AddInt64(&result, int64(gain))
			atomic.AddInt64(another, 1)
		}, 1, &anotherResult)

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		r1 := pr.Go(context.TODO())
		r2 := pr.Go(context.TODO())

		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(0))

		r1.Wait()
		r2.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
		gotwant.Test(t, atomic.LoadInt64(&anotherResult), int64(2))
	})

	t.Run("Cancel", func(t *testing.T) {
		var result int64 = 0
		f := func(c context.Context, params ...interface{}) {
			time.Sleep(50 * time.Millisecond)

			if goroup.ContextDone(c) {
				return
			}

			atomic.AddInt64(&result, int64(1))
		}

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		r := goroup.Go(f, nil)

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		r.Cancel()
		r.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))
	})
}
