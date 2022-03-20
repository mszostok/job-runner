package file

import (
	"sync"

	"github.com/fsnotify/fsnotify"

	"github.com/mszostok/job-runner/internal/shutdown"
)

var _ shutdown.ShutdownableService = &Watcher{}

type Watcher struct {
	watcher        *fsnotify.Watcher
	mu             sync.RWMutex
	activeWatchers map[string][]*Observer
}

type Observer struct {
	Events chan fsnotify.Op
	Errors chan error
	path   string
}

func NewWatcher() (*Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		mu:             sync.RWMutex{},
		watcher:        watcher,
		activeWatchers: map[string][]*Observer{},
	}
	go w.notify()

	return w, nil
}

func (w *Watcher) AddObserver(path string) (*Observer, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.watcher.Add(path); err != nil {
		return nil, err
	}

	observer := &Observer{
		path:   path,
		Events: make(chan fsnotify.Op),
		Errors: make(chan error),
	}

	w.activeWatchers[path] = append(w.activeWatchers[path], observer)

	return observer, nil
}

func (w *Watcher) RemoveObserver(item *Observer) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	observers, found := w.activeWatchers[item.path]
	if !found {
		return nil
	}

	if len(observers) > 1 {
		// there are others
		w.activeWatchers[item.path] = w.filterOutObserver(observers, item)
		return nil
	}

	// we were last
	delete(w.activeWatchers, item.path)
	if err := w.watcher.Remove(item.path); err != nil {
		return err
	}

	return nil
}

func (w *Watcher) Shutdown() error {
	return w.watcher.Close()
}

// Based on: https://github.com/golang/go/wiki/SliceTricks#filter-in-place
func (w *Watcher) filterOutObserver(observers []*Observer, toRemove *Observer) []*Observer {
	n := 0
	for _, item := range observers {
		if item != toRemove {
			observers[n] = item
			n++
		}
	}
	observers = observers[:n]
	return observers
}

func (w *Watcher) notify() {
	for {
		select {
		case e, ok := <-w.watcher.Events:
			if !ok {
				return
			}

			w.mu.RLock()
			for _, observer := range w.activeWatchers[e.Name] {
				switch v := event(e); {
				case v.Is(fsnotify.Write):
					w.trySend(observer.Events, fsnotify.Write)
				case v.Is(fsnotify.Remove):
					w.trySend(observer.Events, fsnotify.Remove)
				case v.Is(fsnotify.Rename):
					w.trySend(observer.Events, fsnotify.Rename)
				}
			}
			w.mu.RUnlock()

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}

			w.mu.RLock()
			for _, observers := range w.activeWatchers {
				for _, observer := range observers {
					observer.Errors <- err
				}
			}
			w.mu.RUnlock()
		}
	}
}

func (w *Watcher) trySend(channel chan fsnotify.Op, op fsnotify.Op) {
	select {
	case channel <- op:
	default:
		// message dropped
	}
}

type event fsnotify.Event

func (e event) Is(op fsnotify.Op) bool {
	return e.Op&op == op
}
