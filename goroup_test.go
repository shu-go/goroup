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
		g := goroup.Group(&r)
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
		r := goroup.Routine{}
		r.Wait()
		r.Go()
		r.Wait()
		r.Cancel()

		<-r.Done()
	})

	t.Run("NullGoroup", func(t *testing.T) {
		g := goroup.Goroup{}
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
		g := goroup.Group(&goroup.Routine{})
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
		j1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c() {
				atomic.AddInt64(&result1, 1)
			}
		})

		j1.Go()
		j1.Go()
		j1.Go()
		j1.Go()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))

		j1.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(1))
	})

	t.Run("Cancel", func(t *testing.T) {
		var result1 int64
		j1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c() {
				atomic.AddInt64(&result1, 1)
			}
		})

		j1.Go()

		j1.Cancel()
		j1.Cancel()
		j1.Cancel()
		j1.Cancel()
		j1.Cancel()
		j1.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))
	})

	t.Run("Done", func(t *testing.T) {
		var result1 int64
		j1 := goroup.Ready(func(c goroup.Cancelled) {
			time.Sleep(50 * time.Millisecond)
			if !c() {
				atomic.AddInt64(&result1, 1)
			}
		})

		j1.Go()
		ch := j1.Done()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))

		<-ch
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(1))
	})
}

func TestGoroup(t *testing.T) {
	var result int64
	j1 := goroup.Ready(func(c goroup.Cancelled) {
		time.Sleep(50 * time.Millisecond)
		if !c() {
			atomic.AddInt64(&result, 1)
		}
	})
	j2 := goroup.Ready(func(c goroup.Cancelled) {
		time.Sleep(20 * time.Millisecond)
		if !c() {
			atomic.AddInt64(&result, 2)
		}
	})
	j3 := goroup.Ready(func(c goroup.Cancelled) {
		time.Sleep(30 * time.Millisecond)
		if !c() {
			atomic.AddInt64(&result, 4)
		}
	})

	t.Run("Wait", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		jj1.Go()
		//g := goroup.Group(&jj1, &jj2, &jj3)
		g := goroup.Group()
		g.Add(&jj1)
		g.Add(&jj2)
		g.Add(&jj3)
		g.Go()
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("WaitAny", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		jj1.Go()
		g := goroup.Group(&jj1, &jj2, &jj3)
		g.Go()
		g.Go()
		g.Go()
		g.Go()
		g.WaitAny() // jj2
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
		g.WaitAny() // still jj2
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
		g.PurgeDone() // purge jj2

		g.WaitAny() // jj3
		gotwant.Test(t, atomic.LoadInt64(&result), int64(6))
		g.PurgeDone() // purge jj3

		g.WaitAny() // jj1
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("Cancel", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		g := goroup.Group(&jj1, &jj2, &jj3)
		g.Go()

		g.WaitAny() // jj2
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))

		g.Cancel()
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
	})

	t.Run("Done", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		jj1.Go()
		//g := goroup.Group(&jj1, &jj2, &jj3)
		g := goroup.Group()
		g.Add(&jj1)
		g.Add(&jj2)
		g.Add(&jj3)
		g.Go()

		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))
		<-g.Done()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("DoneAny", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		jj1.Go()
		g := goroup.Group(&jj1, &jj2, &jj3)
		g.Go()
		g.Go()
		g.Go()
		g.Go()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(0))

		<-g.DoneAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
		<-g.DoneAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
		g.PurgeDone() // purge jj2

		<-g.DoneAny()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(6))
		g.PurgeDone() // purge jj3

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
		g1 := goroup.Group(&r)
		g2 := goroup.Group(&r)

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
		g1 := goroup.Group(&r)

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
		if !c() {
			atomic.AddInt64(&result, 1)
		}
	})

	g := goroup.Group()
	g.Add(&r)

	r.Wait()
	g.Wait()
	// this test does end

	gotwant.Test(t, atomic.LoadInt64(&result), int64(0))
}
