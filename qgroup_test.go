package goroup_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"bitbucket.org/shu_go/goroup"
	"bitbucket.org/shu_go/gotwant"
)

func TestQGroup(t *testing.T) {
	t.Run("BasicUsage", func(t *testing.T) {
		var result int64 = 0

		qg := goroup.NewQueuedGroup(nil, 1)
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			atomic.AddInt64(&result, int64(gain))
		}

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		qg.Add(goroup.Ready(f, 2))
		qg.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))

		qg = goroup.NewQueuedGroup(nil, 1)
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(12))

		qg = goroup.NewQueuedGroup(nil, 2)
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(22))
	})

	t.Run("BasicUsage2", func(t *testing.T) {
		f := func(c context.Context, params ...interface{}) {
			gain := time.Duration(params[0].(int))
			time.Sleep(gain * time.Millisecond)
		}

		qg := goroup.NewQueuedGroup(nil, 1)
		stt := time.Now()
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Wait()
		gotwant.TestExpr(t, stt, time.Since(stt) > 500*time.Millisecond)

		qg = goroup.NewQueuedGroup(nil, 5)
		stt = time.Now()
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Add(goroup.Ready(f, 100))
		qg.Wait()
		gotwant.TestExpr(t, stt, time.Since(stt) < 500*time.Millisecond && time.Since(stt) > 100*time.Millisecond)
	})

	t.Run("WaitAny", func(t *testing.T) {
		var result int64 = 0

		qg := goroup.NewQueuedGroup(nil, 1)
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			atomic.AddInt64(&result, int64(gain))
			time.Sleep(1)
		}

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= 0 && atomic.LoadInt64(&result) <= 10)
		qg.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= 0 && atomic.LoadInt64(&result) <= 10)
		qg.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= 0 && atomic.LoadInt64(&result) <= 10)
		qg.Wait()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) == 10)
		qg.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) == 10)
	})

	t.Run("WaitAny2", func(t *testing.T) {
		qg := goroup.NewQueuedGroup(nil, 1)

		select {
		case <-goroup.Done(qg.WaitAny):
		case <-time.After(20 * time.Millisecond):
			t.Fail()
		}
	})

	t.Run("Wait2", func(t *testing.T) {
		qg := goroup.NewQueuedGroup(nil, 1)

		select {
		case <-goroup.Done(qg.Wait):
		case <-time.After(20 * time.Millisecond):
			t.Fail()
		}
	})

	t.Run("Cancel", func(t *testing.T) {
		var result int64 = 0

		qg := goroup.NewQueuedGroup(nil, 1)
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			atomic.AddInt64(&result, int64(gain))
		}

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Cancel()
		qg.Wait()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= 0 && atomic.LoadInt64(&result) <= 10)

		qg = goroup.NewQueuedGroup(nil, 5)
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Add(goroup.Ready(f, 2))
		qg.Cancel()
		qg.Wait()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= 0 && atomic.LoadInt64(&result) <= 20)
	})
}
