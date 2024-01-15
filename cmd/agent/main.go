package main

import (
	"context"
	"os"

	"github.com/mszostok/job-runner/internal/xsignal"
)

func main() {
	rootCmd := NewRoot()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = xsignal.WithStopContext(ctx)

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// error is already handled by `cobra`, we don't want to log it here as we will duplicate the message.
		// If needed, based on error type we can exit with different codes.
		os.Exit(1)
	}
}
