// Package shutdown provides small, dependency-free helpers for graceful
// process shutdown: waiting for a termination signal and running a
// bounded-timeout shutdown function against it.
package shutdown

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// WaitForSignals returns a channel that receives the first configured OS signal.
// If no signals are provided, it listens for os.Interrupt and syscall.SIGTERM.
func WaitForSignals(signals ...os.Signal) <-chan os.Signal {
	stop := make(chan os.Signal, 1)
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	signal.Notify(stop, signals...)
	return stop
}

// GracefulShutdown calls shutdownFn with a timeout context and returns any error.
func GracefulShutdown(parent context.Context, timeout time.Duration, shutdownFn func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	return shutdownFn(ctx)
}
