package main

import (
	"context"

	"github.com/AlmirSai/service/foundation/logger"
)

func main() {
	var log *logger.Logger

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "failed to run sales service", "error", err)
		panic("failed to run sales service: " + err.Error())
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	return nil
}
