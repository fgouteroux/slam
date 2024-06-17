package webserver

import (
	"context"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var testQuit chan os.Signal

func NewServerWithTimeout(listenAddress string, readTimeout, writeTimeout, idleTimeout int, t time.Duration) (*http.Server, chan struct{}) {

	idleConnectionsClosed := make(chan struct{})

	srv := &http.Server{
		Addr:         listenAddress,
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}

	sigChan := make(chan os.Signal)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		<-sigChan

		log.Info("shutting down server...")

		// The context is used to inform the server it has n seconds to finish
		// the request it is currently handling
		ctx, cancel := context.WithTimeout(context.Background(), t)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Errorf("Error while shutting down server: %s. Initiating force shutdown...", err.Error())
		}
		close(idleConnectionsClosed)
	}()

	return srv, idleConnectionsClosed
}
