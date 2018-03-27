package goroup_test

import (
	"sync/atomic"
	"testing"
	"time"

	"bitbucket.org/shu/goroup"
	"bitbucket.org/shu/gotwant"
)

func TestDo(t *testing.T) {
	doneChan := goroup.Do(func() { time.Sleep(1 * time.Second) })
	<-doneChan
}

func TestRoutine(t *testing.T) {
	t.Run("Run", func(t *testing.T) {
		var result1 int64
		j1 := goroup.Routine(func(c goroup.Cancelled) {
			time.Sleep(500 * time.Millisecond)
			if !c() {
				atomic.AddInt64(&result1, 1)
			}
		})

		j1.Run()
		j1.Run()
		j1.Run()
		j1.Run()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))

		j1.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(1))
	})
	t.Run("Cancel", func(t *testing.T) {
		var result1 int64
		j1 := goroup.Routine(func(c goroup.Cancelled) {
			time.Sleep(500 * time.Millisecond)
			if !c() {
				atomic.AddInt64(&result1, 1)
			}
		})

		j1.Run()

		j1.Cancel()
		j1.Cancel()
		j1.Cancel()
		j1.Cancel()
		j1.Cancel()
		j1.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result1), int64(0))
	})
}

func TestGroup(t *testing.T) {
	var result int64
	j1 := goroup.Routine(func(c goroup.Cancelled) {
		time.Sleep(500 * time.Millisecond)
		if !c() {
			atomic.AddInt64(&result, 1)
		}
	})
	j2 := goroup.Routine(func(c goroup.Cancelled) {
		time.Sleep(200 * time.Millisecond)
		if !c() {
			atomic.AddInt64(&result, 2)
		}
	})
	j3 := goroup.Routine(func(c goroup.Cancelled) {
		time.Sleep(300 * time.Millisecond)
		if !c() {
			atomic.AddInt64(&result, 4)
		}
	})

	t.Run("Wait", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		jj1.Run()
		//g := goroup.Group(&jj1, &jj2, &jj3)
		g := goroup.Group()
		g.Add(&jj1)
		g.Add(&jj2)
		g.Add(&jj3)
		g.Run()
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(7))
	})

	t.Run("WaitAny", func(t *testing.T) {
		jj1 := j1
		jj2 := j2
		jj3 := j3

		atomic.StoreInt64(&result, 0)

		jj1.Run()
		g := goroup.Group(&jj1, &jj2, &jj3)
		g.Run()
		g.Run()
		g.Run()
		g.Run()
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
		g.Run()

		g.WaitAny() // jj2
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))

		g.Cancel()
		g.Wait()
		gotwant.Test(t, atomic.LoadInt64(&result), int64(2))
	})
}
