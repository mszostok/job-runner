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
		os.Exit(1)
	}
}
