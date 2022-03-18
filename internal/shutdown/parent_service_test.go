package shutdown_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mszostok/job-runner/internal/cli/heredoc"
	"github.com/mszostok/job-runner/internal/shutdown"
)

func TestParentService(t *testing.T) {
	const finishImmediately = 50 * time.Millisecond

	t.Run("shutdown immediately if no children", func(t *testing.T) {
		t.Parallel()
		// given
		var (
			svc = &shutdown.ParentService{}
		)

		// when, then
		err := AssertFuncResultAtMost(t, svc.Shutdown, finishImmediately)
		assert.NoError(t, err)
	})

	t.Run("shutdown immediately if all register children finished", func(t *testing.T) {
		t.Parallel()
		// given
		var (
			childrenWatcher = &childrenServiceWatcher{}
			svc             = &shutdown.ParentService{}

			childrenNo     = 30
			unlockShutdown = make(chan struct{})
		)

		childrenWatcher.ExpectedChildren(childrenNo)
		for i := 0; i < childrenNo; i++ {
			svc.Register(&delayedServiceMock{unlockShutdown, childrenWatcher.ChildDone})
		}

		// when
		close(unlockShutdown)
		res := AssertFuncResultAtMost(t, svc.Shutdown, finishImmediately)

		// then
		assert.NoError(t, res)
		childrenWatcher.RequireResourcesRelease(t)
	})

	t.Run("shutdown immediately without error even if one child returned context.Cancelled", func(t *testing.T) {
		t.Parallel()
		// given
		var (
			childrenWatcher = childrenServiceWatcher{}
			svc             = &shutdown.ParentService{}
			unlockShutdown  = make(chan struct{})
		)
		childrenWatcher.ExpectedChildren(3)

		svc.Register(&delayedServiceMock{unlockShutdown, childrenWatcher.ChildDone})
		svc.Register(&erroneousServiceMock{unlockShutdown, childrenWatcher.ChildDone, context.Canceled})
		svc.Register(&delayedServiceMock{unlockShutdown, childrenWatcher.ChildDone})

		// when
		close(unlockShutdown)
		res := AssertFuncResultAtMost(t, svc.Shutdown, finishImmediately)

		// then
		assert.NoError(t, res)
		childrenWatcher.RequireResourcesRelease(t)
	})

	t.Run("shutdown with error if one child returned generic error", func(t *testing.T) {
		t.Parallel()
		// given
		var (
			childrenWatcher = &childrenServiceWatcher{}
			svc             = &shutdown.ParentService{}
			genericErr      = errors.New("oops! internal error occurred :)")
			unlockShutdown  = make(chan struct{})
		)

		expErr := heredoc.Doc(`
			1 error occurred:
				* oops! internal error occurred :)

			`)
		childrenWatcher.ExpectedChildren(3)
		svc.Register(&delayedServiceMock{unlockShutdown, childrenWatcher.ChildDone})
		svc.Register(&erroneousServiceMock{unlockShutdown, childrenWatcher.ChildDone, context.Canceled})
		svc.Register(&erroneousServiceMock{unlockShutdown, childrenWatcher.ChildDone, genericErr})

		// when
		close(unlockShutdown)
		res := AssertFuncResultAtMost(t, svc.Shutdown, finishImmediately)

		// then
		assert.EqualError(t, res, expErr)
		childrenWatcher.RequireResourcesRelease(t)
	})

	t.Run("shutdown with error if all children returned generic error", func(t *testing.T) {
		t.Parallel()
		// given
		var (
			childrenWatcher = &childrenServiceWatcher{}
			svc             = &shutdown.ParentService{}
			unlockShutdown  = make(chan struct{})
		)

		childrenWatcher.ExpectedChildren(3)
		svc.Register(&erroneousServiceMock{unlockShutdown, childrenWatcher.ChildDone, errors.New("oops! svc1 error :)")})
		svc.Register(&erroneousServiceMock{unlockShutdown, childrenWatcher.ChildDone, errors.New("oops! svc2 error :)")})
		svc.Register(&erroneousServiceMock{unlockShutdown, childrenWatcher.ChildDone, errors.New("oops! svc3 error :)")})

		// when
		close(unlockShutdown)
		res := AssertFuncResultAtMost(t, svc.Shutdown, finishImmediately)

		// then
		require.Error(t, res)
		assert.Contains(t, res.Error(), "3 errors occurred:")
		assert.Contains(t, res.Error(), "	* oops! svc1 error :)")
		assert.Contains(t, res.Error(), "	* oops! svc2 error :)")
		assert.Contains(t, res.Error(), "	* oops! svc3 error :)")

		childrenWatcher.RequireResourcesRelease(t)
	})

	t.Run("shutdown wait for stalled child", func(t *testing.T) {
		t.Parallel()
		// given
		var (
			childrenWatcher = &childrenServiceWatcher{}
			svc             = &shutdown.ParentService{}
		)

		childrenWatcher.ExpectedChildren(3)
		unlockShutdown := make(chan struct{})
		cleanupStalled := make(chan struct{})

		svc.Register(&delayedServiceMock{unlockShutdown, childrenWatcher.ChildDone})
		svc.Register(&stalledServiceMock{cleanupStalled, childrenWatcher.ChildDone})
		svc.Register(&delayedServiceMock{unlockShutdown, childrenWatcher.ChildDone})

		// when, then
		close(unlockShutdown)
		AssertFuncDoesntResultAtMost(t, svc.Shutdown, time.Second)

		// then
		close(cleanupStalled)
		childrenWatcher.RequireResourcesRelease(t)
	})
}

type childrenServiceWatcher struct {
	wg sync.WaitGroup
}

func (ts *childrenServiceWatcher) ExpectedChildren(n int) { ts.wg.Add(n) }
func (ts *childrenServiceWatcher) ChildDone()             { ts.wg.Done() }
func (ts *childrenServiceWatcher) RequireResourcesRelease(t *testing.T) {
	done := make(chan struct{})
	go func() {
		ts.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second * 5):
		t.Fatal("Resource release timed out")
	}
}

// AssertFuncResultAtMost asserts that the given function returned result before timeout
func AssertFuncResultAtMost(t *testing.T, fn func() error, timeout time.Duration) error {
	t.Helper()

	resChanReturned := make(chan struct{})
	var res error
	go func() {
		res = fn()
		close(resChanReturned)
	}()
	select {
	case <-resChanReturned:
	case <-time.After(timeout):
		t.Fatalf("given function didn't return after %v", timeout)
	}

	return res
}

// AssertFuncDoesntResultAtMost asserts that the given function doesn't return before timeout
func AssertFuncDoesntResultAtMost(t *testing.T, fn func() error, timeout time.Duration) {
	t.Helper()

	fnReturned := make(chan struct{})
	go func() {
		_ = fn()
		close(fnReturned)
	}()
	select {
	case <-fnReturned:
		t.Fatalf("given function returned before %v", timeout)
	case <-time.After(timeout):
	}
}

type delayedServiceMock struct {
	ShutdownUnlocked chan struct{}
	OnBackgroundDone func()
}

func (s *delayedServiceMock) Shutdown() error {
	<-s.ShutdownUnlocked
	s.OnBackgroundDone()
	return nil
}

type erroneousServiceMock struct {
	ShutdownUnlocked chan struct{}
	OnBackgroundDone func()
	Error            error
}

func (s *erroneousServiceMock) Shutdown() error {
	<-s.ShutdownUnlocked
	s.OnBackgroundDone()

	return s.Error
}

type stalledServiceMock struct {
	Cleanup          chan struct{}
	OnBackgroundDone func()
}

func (s *stalledServiceMock) Shutdown() error {
	<-s.Cleanup
	s.OnBackgroundDone()
	return nil
}
