package file

import "github.com/spf13/afero"

// Option provides an option to configure Service instance.
type Option func(cfg *Logger)

func WithFS(fs afero.Fs) Option {
	return func(cfg *Logger) {
		cfg.filesystem = fs
	}
}

func WithBufferSize(buffSize int) Option {
	return func(cfg *Logger) {
		cfg.readBufferSize = buffSize
	}
}

func WithLogsDir(baseDir string) Option {
	return func(cfg *Logger) {
		cfg.logsDir = baseDir
	}
}
