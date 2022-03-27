package file

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"

	"github.com/mszostok/job-runner/internal/shutdown"
)

// Ensure on compilation phase, that Logger implements shutdown.ShutdownableService.
var _ shutdown.ShutdownableService = &Logger{}

const filePerm = 0666

// Logger provides functionality to stream and fetch logs via file.
type Logger struct {
	logsDir        string
	readBufferSize int
	filesystem     afero.Fs
	watcher        *Watcher

	activeSinks sync.Map
}

// NewLogger returns a new Logger instance.
func NewLogger(opts ...Option) (*Logger, error) {
	watcher, err := NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "while creating log watcher")
	}
	l := &Logger{
		// TODO: os.MkdirTemp is better as currently "directory is neither guaranteed to exist nor have accessible permissions".
		//logsDir:        os.TempDir(),
		logsDir:        "/tmp",
		filesystem:     afero.NewOsFs(),
		watcher:        watcher,
		readBufferSize: 4096,
	}

	for _, option := range opts {
		option(l)
	}

	return l, nil
}

type ReleaseSinkFn func() error

// NewSink returns a new file sink.
// It's up to the caller to release returned Sink when it's not needed anymore.
func (l *Logger) NewSink(name string) (io.Writer, ReleaseSinkFn, error) {
	path := l.dst(name)
	f, err := l.filesystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return nil, nil, NewConflictError(name)
		}
		return nil, nil, errors.Wrap(err, "while opening file")
	}

	// fsnotify doesn't support close event: https://github.com/fsnotify/fsnotify/issues/22
	closeSink := make(chan struct{})
	l.activeSinks.Store(path, closeSink)
	release := func() error {
		close(closeSink)
		l.activeSinks.LoadAndDelete(path)
		return f.Close()
	}
	return f, release, nil
}

// ReadAndFollow reads Job logs and if log file is still in use, start watching it for a new entries.
func (l *Logger) ReadAndFollow(ctx context.Context, name string) (<-chan []byte, <-chan error, error) {
	path := l.dst(name)
	file, err := l.OpenWithReadDeadliner(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while opening log file")
	}

	var (
		sink   = make(chan []byte)
		issues = make(chan error)
	)

	cleanup := func() {
		if err := file.Close(); err != nil {
			issues <- err
		}

		close(issues)
		close(sink)
	}

	go func() {
		select {
		case <-ctx.Done():
			_ = file.SetReadDeadline(time.Now()) // release currently-blocked Read call
			// TODO: log error, we cannot push it to `issues` without proper synchronization and ensuring that
			// channel is still open.
		case <-sink:
			// nop, just release the goroutine
		}
	}()

	go func() {
		defer cleanup()

		// 1. Drain file till EOF.
		if err := l.drainFileIgnoreEOF(file, sink); err != nil {
			issues <- err
			return
		}

		closedSink, active := l.activeSinks.Load(path)
		if !active { // sink not active, no reason to observe it.
			issues <- io.EOF
			return
		}
		closedSinkNotify, ok := closedSink.(chan struct{})
		if !ok {
			issues <- errors.New("internal error: got incorrect sink notify type")
			return
		}
		// 2. Add observer for file changes.
		//    - WRITE - Sends new data. Assumption is that is always an appending action.
		//              Otherwise, we would need to place with file size and do a proper file seek.
		//    - DELETE - Sends EOF.
		//    - RENAME - Sends EOF.
		// TODO: We can think about closing this file on idle and open it on event to don't run into file descriptor limit.
		observer, err := l.watcher.AddObserver(path)
		if err != nil {
			issues <- err
			return
		}
		defer func() {
			if err := l.watcher.RemoveObserver(observer); err != nil {
				issues <- err
			}
		}()

		for {
			select {
			case <-closedSinkNotify:
				issues <- io.EOF
				return
			case <-ctx.Done():
				issues <- ctx.Err()
				return
			case event, ok := <-observer.Events:
				if !ok {
					return
				}
				switch event {
				case fsnotify.Write:
					if err := l.drainFileIgnoreEOF(file, sink); err != nil {
						issues <- err
						return
					}
				case fsnotify.Remove, fsnotify.Rename:
					issues <- io.EOF
					return
				}

			case err := <-observer.Errors:
				// don't need to check for closed chan, nil objs are also allowed
				issues <- err
				return
			}
		}
	}()

	return sink, issues, nil
}

// Shutdown removes all watches and closes the events channels.
func (l *Logger) Shutdown() error {
	return l.watcher.Shutdown()
}

func (l *Logger) dst(name string) string {
	return filepath.Join(l.logsDir, name)
}

// ReadCloseDeadliner represents a file in the filesystem.
type ReadCloseDeadliner interface {
	io.ReadCloser

	// SetReadDeadline sets the deadline for future Read calls and any
	// currently-blocked Read call.
	// A zero value for t means Read will not time out.
	SetReadDeadline(t time.Time) error
}

func (l *Logger) OpenWithReadDeadliner(path string) (ReadCloseDeadliner, error) {
	file, err := l.filesystem.Open(path)
	if err != nil {
		return nil, err
	}
	out, ok := file.(ReadCloseDeadliner)
	if !ok {
		_ = file.Close()
		return nil, errors.New("filesystem cannot open file with 'SetReadDeadline' support")
	}

	return out, nil
}

func (l *Logger) drainFileIgnoreEOF(file io.Reader, sink chan<- []byte) error {
	buff := make([]byte, l.readBufferSize)

	for {
		n, err := file.Read(buff)
		switch {
		case err == nil:
			out := make([]byte, n)
			copy(out, buff[:n])
			sink <- out
		case err == io.EOF:
			// This EOF is ignored, as later we want to watch this file for changes.
			return nil
		default:
			return errors.Wrap(err, "while reading log file")
		}
	}
}
