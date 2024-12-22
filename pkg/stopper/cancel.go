package stopper

import (
	"context"
	"os/signal"
	"syscall"
)

func MakeContext(parent context.Context) (ctx context.Context, cancel context.CancelFunc) {
	return signal.NotifyContext(parent,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT)
}
