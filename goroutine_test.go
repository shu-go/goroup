package goroup_test

import (
	"testing"
	"time"

	"bitbucket.org/shu_go/goroup"
	"bitbucket.org/shu_go/gotwant"
)

func TestDone(t *testing.T) {
	doneChan := goroup.Done(func() { time.Sleep(100 * time.Millisecond) })
	<-doneChan
}

func TestSelect(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		_, _, ok := goroup.TrySelect()
		gotwant.Test(t, ok, false)

		_, _, ok = goroup.Select()
		gotwant.Test(t, ok, false)
	})

	t.Run("Closed", func(t *testing.T) {
		c1 := make(chan int, 1)
		close(c1)

		_, _, ok := goroup.TrySelect(c1)
		gotwant.Test(t, ok, false)

		_, _, ok = goroup.Select(c1)
		gotwant.Test(t, ok, false)
	})

	t.Run("SingleImmediate", func(t *testing.T) {
		c1 := make(chan int, 1)

		c1 <- 1

		c, v, ok := goroup.TrySelect(c1)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, c, c1)
		gotwant.Test(t, v, 1)

		c1 <- 1

		c, v, ok = goroup.Select(c1)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, c, c1)
		gotwant.Test(t, v, 1)
	})

	t.Run("Single", func(t *testing.T) {
		c1 := make(chan int, 1)

		go func() {
			time.Sleep(50 * time.Millisecond)
			c1 <- 1
		}()

		c, v, ok := goroup.TrySelect(c1)
		gotwant.Test(t, ok, false)
		gotwant.Test(t, c, nil)
		gotwant.Test(t, v, nil)

		c, v, ok = goroup.Select(c1)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, v, 1)
		gotwant.Test(t, c, c1)
	})

	t.Run("Multi", func(t *testing.T) {
		c1 := make(chan struct{}, 1)
		c2 := make(chan int, 1)
		c3 := make(chan string, 1)

		go func() {
			time.Sleep(50 * time.Millisecond)
			c1 <- struct{}{}
		}()

		c, v, ok := goroup.TrySelect(c1, c2, c3) // not yet
		gotwant.Test(t, ok, false)
		gotwant.Test(t, c, nil)
		gotwant.Test(t, v, nil)

		c, v, ok = goroup.Select(c1, c2, c3)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, c, c1)
		gotwant.Test(t, v, struct{}{})

		go func() {
			time.Sleep(50 * time.Millisecond)
			c2 <- 100
			time.Sleep(50 * time.Millisecond)
			c3 <- "ichi"
		}()

		c, v, ok = goroup.TrySelect(c1, c2, c3) // not yet
		gotwant.Test(t, ok, false)
		gotwant.Test(t, c, nil)
		gotwant.Test(t, v, nil)

		c, v, ok = goroup.Select(c1, c2, c3)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, c, c2)
		gotwant.Test(t, v, 100)

		c, v, ok = goroup.TrySelect(c1, c2, c3) // consumed, for now
		gotwant.Test(t, ok, false)
		gotwant.Test(t, c, nil)
		gotwant.Test(t, v, nil)

		time.Sleep(50 * time.Millisecond)
		c, v, ok = goroup.Select(c1, c2, c3)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, c, c3)
		gotwant.Test(t, v, "ichi")

		// DO NOT. causes "fatal error: all goroutines are asleep - deadlock!"
		//c, v, ok = goroup.Select(c1, c2, c3)
		//gotwant.Test(t, ok, false)
		c, v, ok = goroup.TrySelect(c1, c2, c3)
		gotwant.Test(t, ok, false)
		gotwant.Test(t, c, nil)
		gotwant.Test(t, v, nil)
	})

}
