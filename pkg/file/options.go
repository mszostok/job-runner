package file

import "github.com/spf13/afero"

// Option provides an option to configure Service instance.
type Option func(cfg *Logger)

// WithFS changes used file system
func WithFS(fs afero.Fs) Option {
	return func(cfg *Logger) {
		cfg.filesystem = fs
	}
}

// WithBufferSize changes the maximum buffer size (chunk) read from the underlying log file.
func WithBufferSize(buffSize int) Option {
	return func(cfg *Logger) {
		cfg.readBufferSize = buffSize
	}
}

// WithLogsDir changes base dir log. It needs to exist.
func WithLogsDir(baseDir string) Option {
	return func(cfg *Logger) {
		cfg.logsDir = baseDir
	}
}
