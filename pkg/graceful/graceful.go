package graceful

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jordanp/goapp/pkg/log"
)

func MakeListenAndServe(log log.Logger, timeout time.Duration) func(addr string, handler http.Handler) error {
	return func(addr string, handler http.Handler) error {
		if host, port, err := net.SplitHostPort(addr); err == nil {
			if host == "" {
				host = net.IPv4zero.String()
			}
			log.Infof("start listening on http://%s", net.JoinHostPort(host, port))
		}
		httpServer := http.Server{Addr: addr, Handler: handler}
		errorC := make(chan error, 1)
		go func() {
			if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
				errorC <- err
			}
		}()
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		select {
		case err := <-errorC:
			return err
		case sig := <-signals:
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			log.Infof("starting server shutdown (signal=%s timeout=%s)", sig.String(), timeout)
			if err := httpServer.Shutdown(ctx); err != nil {
				return fmt.Errorf("unclean shutdown of HTTP server: %s", err)
			}

			log.Debug("finished all in-flight HTTP requests")
			if deadline, ok := ctx.Deadline(); ok {
				secs := (time.Until(deadline) + time.Second/2) / time.Second
				log.Debugf("shutdown finished %ds before deadline", secs)
			}
			return nil

		}
	}
}
