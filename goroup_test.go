package goroup_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/shu-go/goroup"
	"github.com/shu-go/gotwant"
)

func TestGoroup(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		var result int64 = 0

		gg := goroup.New()
		gg.Add(func(context.Context) {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&result, int64(1))
		})
		gg.Add(func(context.Context) {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&result, int64(2))
		})
		gg.Go()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		gg.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(3))
	})

	t.Run("WaitAny", func(t *testing.T) {
		var result int64 = 0

		f := func(context.Context) {
			atomic.AddInt64(&result, int64(1))
		}

		gg := goroup.New(goroup.MaxGoroutines(3))
		for i := 0; i < 10; i++ {
			gg.Add(f)
		}
		gg.Go()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		gg.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > int64(0))

		gg.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(10))
	})

	t.Run("WaitAnyCancel", func(t *testing.T) {
		var result int64 = 0

		f := func(context.Context) {
			atomic.AddInt64(&result, int64(1))
		}

		gg := goroup.New(goroup.MaxGoroutines(3))
		for i := 0; i < 10; i++ {
			gg.Add(f)
		}
		gg.Go()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		gg.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > int64(0))

		gg.Cancel()

		gg.Wait()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) > int64(0))
	})

    t.Run("GoAfterWait", func(t *testing.T)  {
		var result int64 = 0
		f := func(context.Context) {
			atomic.AddInt64(&result, int64(1))
		}

        gg := goroup.New()
        gg.Add(f)
        gg.Go()
        gg.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(1))

        gg.Add(f)
        gg.Go()
        gg.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
    })

    t.Run("GoAfterWaitAny", func(t *testing.T)  {
		var result int64 = 0
		f := func(context.Context) {
			atomic.AddInt64(&result, int64(1))
		}

        gg := goroup.New()
        gg.Add(f)
        gg.Go()
        gg.WaitAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(1))

        gg.Add(f)
        gg.Go()
        gg.WaitAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
    })
}
