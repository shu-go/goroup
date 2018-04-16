package goroup_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"bitbucket.org/shu/goroup"
	"bitbucket.org/shu/gotwant"
)

func TestDo(t *testing.T) {
	doneChan := goroup.Done(func() { time.Sleep(100 * time.Millisecond) })
	<-doneChan
}

func TestStates(t *testing.T) {
	t.Run("Routine", func(t *testing.T) {
		r := goroup.Ready(func(c goroup.Cancelled) { time.Sleep(100 * time.Millisecond) })
		<-r.Done() // does not block

		r.Go()
		select {
		case <-r.Done(): // blocks
			t.Error("!?")
		default:
		}

		time.Sleep(100 * time.Millisecond)
		<-r.Done() // does not block
	})

	t.Run("Group", func(t *testing.T) {
		r := goroup.Ready(func(c goroup.Cancelled) { time.Sleep(100 * time.Millisecond) })
		g := goroup.NewGroup(r)
		<-g.Done() // does not block

		g.Go()
		select {
		case <-g.Done(): // blocks
			t.Error("!?")
		default:
		}

		time.Sleep(100 * time.Millisecond)
		<-g.Done() // does not block
	})
}

func TestRoutine(t *testing.T) {
	t.Run("NullRoutine", func(t *testing.T) {
		r := goroup.Ready(nil)
		r.Wait()
		r.Go()
		r.Wait()
		r.Cancel()

		<-r.Done()
	})

	t.Run("NullGoroup", func(t *testing.T) {
		g := goroup.NewGroup()
		g.WaitAny()
		g.Wait()
		g.Go()
		g.WaitAny()
		g.Wait()
		g.Cancel()

		<-g.DoneAny()
		<-g.Done()

		g.PurgeDone()
	})

	t.Run("GoroupWithNullRoutine", func(t *testing.T) {
		r := goroup.Ready(nil)
		g := goroup.NewGroup(r)
		g.WaitAny()
		g.Wait()
		g.Go()
		g.WaitAny()
		g.Wait()
		g.Cancel()

		<-g.DoneAny()
		<-g.Done()

		g.PurgeDone()
	})

	t.Run("Go", func(t *testing.T) {
		var result1 int64
		r1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c.Cancelled() {
				atomic.AddInt64(&result1, 1)
			}
		})

		r1.Go()
		r1.Go()
		r1.Go()
		r1.Go()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))

		r1.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(1))
	})

	t.Run("GoCopied", func(t *testing.T) {
		var result1 int64
		r1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c.Cancelled() {
				atomic.AddInt64(&result1, 1)
			}
		})

		r2 := r1

		r1.Go()
		r2.Go()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))

		r1.Wait()
		r2.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(1))
	})

	t.Run("Cancel", func(t *testing.T) {
		var result1 int64
		r1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c.Cancelled() {
				atomic.AddInt64(&result1, 1)
			}
		})

		r1.Go()

		r1.Cancel()
		r1.Cancel()
		r1.Cancel()
		r1.Cancel()
		r1.Cancel()
		r1.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))
	})

	t.Run("Done", func(t *testing.T) {
		var result1 int64
		r1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c.Cancelled() {
				atomic.AddInt64(&result1, 1)
			}
		})

		r1.Go()
		ch := r1.Done()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))

		<-ch
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(1))
	})
}

func TestGoroup(t *testing.T) {
	var result int64
	f1 := func(c goroup.Cancelled) {
		time.Sleep(50 * time.Millisecond)
		if !c.Cancelled() {
			atomic.AddInt64(&result, 1)
		}
	}
	f2 := func(c goroup.Cancelled) {
		time.Sleep(20 * time.Millisecond)
		if !c.Cancelled() {
			atomic.AddInt64(&result, 2)
		}
	}
	f3 := func(c goroup.Cancelled) {
		time.Sleep(30 * time.Millisecond)
		if !c.Cancelled() {
			atomic.AddInt64(&result, 4)
		}
	}

	t.Run("Wait", func(t *testing.T) {

		r1 := goroup.Ready(f1)
		r2 := goroup.Ready(f2)
		r3 := goroup.Ready(f3)

		atomic.StoreInt64(&result, 0)

		r1.Go()
		//g := goroup.NewGroup(r1, r2, r3)
		g := goroup.NewGroup()
		g.Add(r1)
		g.Add(r2)
		g.Add(r3)
		g.Go()
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("WaitAny", func(t *testing.T) {
		r1 := goroup.Ready(f1)
		r2 := goroup.Ready(f2)
		r3 := goroup.Ready(f3)

		atomic.StoreInt64(&result, 0)

		r1.Go()
		g := goroup.NewGroup(r1, r2, r3)
		g.Go()
		g.Go()
		g.Go()
		g.Go()
		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(1))
		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(1))
		g.PurgeDone()

		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(3))
		g.PurgeDone()

		g.WaitAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("Cancel", func(t *testing.T) {
		r1 := goroup.Ready(f1)
		r2 := goroup.Ready(f2)
		r3 := goroup.Ready(f3)

		atomic.StoreInt64(&result, 0)

		g := goroup.NewGroup(r1, r2, r3)
		g.Go()

		g.WaitAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(1))

		g.Cancel()
		g.Wait()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(1))
	})

	t.Run("Done", func(t *testing.T) {
		r1 := goroup.Ready(f1)
		r2 := goroup.Ready(f2)
		r3 := goroup.Ready(f3)

		atomic.StoreInt64(&result, 0)

		r1.Go()
		//g := goroup.NewGroup(r1, r2, r3)
		g := goroup.NewGroup()
		g.Add(r1)
		g.Add(r2)
		g.Add(r3)
		g.Go()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))
		<-g.Done()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("DoneAny", func(t *testing.T) {
		r1 := goroup.Ready(f1)
		r2 := goroup.Ready(f2)
		r3 := goroup.Ready(f3)

		atomic.StoreInt64(&result, 0)

		r1.Go()
		g := goroup.NewGroup(r1, r2, r3)
		g.Go()
		g.Go()
		g.Go()
		g.Go()

		<-g.DoneAny() // may be r2, but not guaranteed
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(1))
		<-g.DoneAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(1))
		g.PurgeDone() // purge r2

		g.PurgeDone() // purge r2
		<-g.DoneAny()
		gotwant.TestExpr(t, atomic.LoadInt64(&result), atomic.LoadInt64(&result) >= int64(3))
		g.PurgeDone() // purge r3

		<-g.DoneAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})
}

func TestCollision(t *testing.T) {
	t.Run("MultiGo", func(t *testing.T) {
		var result int64
		r := goroup.Ready(func(c goroup.Cancelled) {
			atomic.AddInt64(&result, 1)
		})

		r.Go()
		r.Go()
		r.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(1))
	})

	t.Run("MultiGoAsync", func(t *testing.T) {
		var result int64
		r := goroup.Ready(func(c goroup.Cancelled) {
			atomic.AddInt64(&result, 1)
		})

		wg := sync.WaitGroup{}
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				r.Go()
				wg.Done()
			}()
		}
		wg.Wait()

		r.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(1))
	})

	t.Run("MultiGroup", func(t *testing.T) {
		var result int64
		r := goroup.Ready(func(c goroup.Cancelled) {
			atomic.AddInt64(&result, 1)
		})
		g1 := goroup.NewGroup(r)
		g2 := goroup.NewGroup(r)

		g1.Go()
		g2.Go()
		g1.Wait()
		g2.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(1))
	})

	t.Run("MultiGroupAsync", func(t *testing.T) {
		var result int64
		r := goroup.Ready(func(c goroup.Cancelled) {
			atomic.AddInt64(&result, 1)
		})
		g1 := goroup.NewGroup(r)

		wg := sync.WaitGroup{}
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				g1.Go()
				wg.Done()
			}()
		}
		wg.Wait()

		g1.Wait()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(1))
	})
}

func TestSequence(t *testing.T) {
	var result int64
	r := goroup.Ready(func(c goroup.Cancelled) {
		time.Sleep(50 * time.Millisecond)
		if !c.Cancelled() {
			atomic.AddInt64(&result, 1)
		}
	})

	g := goroup.NewGroup()
	g.Add(r)

	r.Wait()
	g.Wait()
	// this test does end

	gotwant.Test(t, atomic.LoadInt64(&result), int64(0))
}
