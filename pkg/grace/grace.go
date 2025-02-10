package grace

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Handle(process Process, logger *zap.Logger) {
	sigQuit := make(chan os.Signal, 2)
	signal.Notify(sigQuit, syscall.SIGINT, syscall.SIGTERM)
	eg := new(errgroup.Group)

	eg.Go(func() error {
		s := <-sigQuit
		return fmt.Errorf("%v", s)
	})

	go process.Run()

	if err := eg.Wait(); err != nil {
		logger.Info("graceful shut down", zap.NamedError("captured signal", err))
		process.Shutdown()
	}
}
