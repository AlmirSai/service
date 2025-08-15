package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/AlmirSai/service/foundation/logger"
)

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "********** SEND ALERT **********")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		// TODO: Implement a function to extract trace ID from context
		// TODO: web.GetTraceID(ctx)
		return ""
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "SALES", traceIDFn, events)

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "failed to run sales service", "error", err)
		panic("failed to run sales service: " + err.Error())
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	sig := <-shutdown

	log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
	defer log.Info(ctx, "shutdown", "status", "shutdown completed", "signal", sig)

	return nil
}
