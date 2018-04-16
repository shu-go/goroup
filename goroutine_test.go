package goroup_test

import (
	"testing"
	"time"

	"bitbucket.org/shu/goroup"
	"bitbucket.org/shu/gotwant"
)

func TestDone(t *testing.T) {
	doneChan := goroup.Done(func() { time.Sleep(100 * time.Millisecond) })
	<-doneChan
}

func TestSelect(t *testing.T) {
	t.Run("None", func(t *testing.T) {
		_, _, ok := goroup.Select()
		gotwant.Test(t, ok, false)
	})

	t.Run("Closed", func(t *testing.T) {
		c1 := make(chan int, 1)
		close(c1)

		_, _, ok := goroup.Select(c1)
		gotwant.Test(t, ok, false)
	})

	t.Run("Single", func(t *testing.T) {
		c1 := make(chan int, 1)

		c1 <- 1

		c, v, ok := goroup.Select(c1)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, v, 1)
		gotwant.Test(t, c, c1)
	})

	t.Run("Multi", func(t *testing.T) {
		c1 := make(chan struct{}, 1)
		c2 := make(chan int, 1)
		c3 := make(chan string, 1)

		c1 <- struct{}{}

		c, v, ok := goroup.Select(c1, c2, c3)
		gotwant.Test(t, ok, true)
		gotwant.Test(t, v, struct{}{})
		gotwant.Test(t, c, c1)

		c2 <- 100
		c3 <- "ichi"
		c, v, ok = goroup.Select(c1, c2, c3)
		gotwant.Test(t, ok, true)
		gotwant.TestExpr(t, v, v == 100 || v == "ichi")
		gotwant.TestExpr(t, c, c == c2 || c == c3)

		c, v, ok = goroup.Select(c1, c2, c3)
		gotwant.Test(t, ok, true)
		gotwant.TestExpr(t, v, v == 100 || v == "ichi")
		gotwant.TestExpr(t, c, c == c2 || c == c3)

		// DO NOT. causes "fatal error: all goroutines are asleep - deadlock!"
		//c, v, ok = goroup.Select(c1, c2, c3)
		//gotwant.Test(t, ok, false)
	})

}
