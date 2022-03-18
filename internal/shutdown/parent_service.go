package shutdown

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
)

// ShutdownableService represents a service that supports graceful shutdown pattern.
type ShutdownableService interface {
	Shutdown() error
}

// ParentService aggregates services that are Shutdownable.
// Those services are registered in parent and shutdown is cascaded to them.
type ParentService struct {
	// children is a list of services which should be shut down together.
	// Each service is implementing ShutdownableService interface and may be shut-down in parallel.
	children []ShutdownableService
}

// Register is registering dependent service to be shutdown on parent service shutdown.
func (s *ParentService) Register(child ShutdownableService) {
	s.children = append(s.children, child)
}

// Shutdown is called to trigger shutdown of all associated children. It waits for all children.
//
// Child shutdown is considered successful also in cases when context.Cancelled error is returned.
func (s *ParentService) Shutdown() error {
	childShutdownFeedback := make(chan error, len(s.children))
	wg := &sync.WaitGroup{}

	// trigger shutdown
	for _, child := range s.children {
		wg.Add(1)
		go func(child ShutdownableService) {
			childShutdownFeedback <- child.Shutdown()
			wg.Done()
		}(child)
	}

	// Wait for all children to shut down.
	// TODO: we can think about some timeout but on the other hand it's scheduled on app shutdown level,
	//       so leaking goroutines is not a problem.
	wg.Wait()

	// At this point we are sure that all children responded.
	close(childShutdownFeedback)

	// produce single result
	var result *multierror.Error
	for err := range childShutdownFeedback {
		if err == nil || err == context.Canceled {
			continue
		}
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}
