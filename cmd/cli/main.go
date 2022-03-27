package main

import (
	"context"
	"log"
	"os"

	"github.com/mszostok/job-runner/internal/cli/config"
	"github.com/mszostok/job-runner/internal/xsignal"
)

func main() {
	rootCmd := NewRoot()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = xsignal.WithStopContext(ctx)

	if err := config.Initialize(); err != nil {
		log.Fatalf("while initialing config: %v", err)
	}

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
