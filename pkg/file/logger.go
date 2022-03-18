package file

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"

	"github.com/mszostok/job-runner/internal/ctxutil"
	"github.com/mszostok/job-runner/internal/shutdown"
)

var _ shutdown.ShutdownableService = &Logger{}

const filePerm = 0666

type Logger struct {
	logsDir        string
	readBufferSize int
	filesystem     afero.Fs
	watcher        *Watcher

	activeSinks sync.Map
}

func NewLogger(opts ...Option) (*Logger, error) {
	watcher, err := NewWatcher()
	if err != nil {
		return nil, errors.Wrap(err, "while creating log watcher")
	}
	l := &Logger{
		// TODO: os.MkdirTemp is better as currently "directory is neither guaranteed to exist nor have accessible permissions".
		logsDir:        os.TempDir(),
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

func (l *Logger) NewSink(name string) (io.Writer, ReleaseSinkFn, error) {
	path := l.dst(name)
	fmt.Println(path)
	f, err := l.filesystem.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, filePerm)
	switch {
	case err == nil:
	case errors.Is(err, fs.ErrExist):
		return nil, nil, NewConflictError(name)
	default:
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

func (l *Logger) ReadAndFollow(ctx context.Context, name string) (<-chan []byte, <-chan error, error) {
	path := l.dst(name)
	file, err := l.filesystem.Open(path)
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
		defer cleanup()

		// 1. Drain file till EOF.
		if err := l.drainFileIgnoreEOF(ctx, file, sink); err != nil {
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
					if err := l.drainFileIgnoreEOF(ctx, file, sink); err != nil {
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

func (l *Logger) Shutdown() error {
	return l.watcher.Shutdown()
}

func (l *Logger) dst(name string) string {
	return filepath.Join(l.logsDir, name)
}

func (l *Logger) drainFileIgnoreEOF(ctx context.Context, file io.Reader, sink chan<- []byte) error {
	buff := make([]byte, l.readBufferSize)

	for {
		if ctxutil.ShouldExit(ctx) { // consumer gone
			return ctx.Err()
		}

		n, err := file.Read(buff)
		switch {
		case err == nil:
			sink <- buff[:n]
		case err == io.EOF:
			// This EOF is ignored, as later we want to watch this file for changes.
			return nil
		default:
			return errors.Wrap(err, "while reading log file")
		}
	}
}
