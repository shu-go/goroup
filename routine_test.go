package goroup_test

import (
	"context"
	"testing"
	"time"

	"bitbucket.org/shu/goroup"
	"bitbucket.org/shu/gotwant"
)

func TestRoutine(t *testing.T) {
	t.Run("BasicUsage", func(t *testing.T) {
		result := 0
		f := func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			result += gain
		}

		gotwant.Test(t, result, 0)

		r := goroup.Go(f, nil, 2)

		gotwant.Test(t, result, 0)

		r.Wait()

		gotwant.Test(t, result, 2)
	})

	t.Run("PreRoutine", func(t *testing.T) {
		result := 0
		pr := goroup.Ready(func(c context.Context, params ...interface{}) {
			gain := params[0].(int)
			result += gain
		}, 1)

		gotwant.Test(t, result, 0)

		r1 := pr.Go(nil)
		r2 := pr.Go(nil)

		gotwant.Test(t, result, 0)

		r1.Wait()
		r2.Wait()

		gotwant.Test(t, result, 2)
	})

	t.Run("Cancel", func(t *testing.T) {
		result := 0
		f := func(c context.Context, params ...interface{}) {
			time.Sleep(50 * time.Millisecond)

			if goroup.ContextDone(c) {
				return
			}

			result++
		}

		gotwant.Test(t, result, 0)

		r := goroup.Go(f, nil)

		gotwant.Test(t, result, 0)

		r.Cancel()
		r.Wait()

		gotwant.Test(t, result, 0)
	})
}
