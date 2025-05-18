package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrlinqu/ltdav/internal/app/ltdav/config"
	proj_cfg "github.com/mrlinqu/ltdav/internal/config"
	dav_server "github.com/mrlinqu/ltdav/internal/dav-server"
	"github.com/rs/zerolog/log"
)

func main() {
	ctx := config.Init(context.Background())

	proj_cfg.LogBuildInfo()

	workingDir := config.GetValue(ctx, config.WorkDir)
	listenAddr := config.GetValue(ctx, config.Addr)

	if workingDir == "" {
		log.Panic().Msg("working dir is not defined")
	}

	if listenAddr == "" {
		log.Panic().Msg("listen addr is not defined")
	}

	tlsCertPath := config.GetValue(ctx, config.CertFile)
	tlsKeyPath := config.GetValue(ctx, config.KeyFile)
	authFile := config.GetValue(ctx, config.AuthFile)

	srv := dav_server.New(listenAddr, workingDir).
		WithTls(tlsCertPath, tlsKeyPath).
		WithAuth(authFile, "aaa")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := srv.ListenAndServe(ctx)
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server listen")
			c <- os.Signal(nil)
		}
	}()

	log.Info().Msg("server started")

	<-c

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer func() {
		cancel()
	}()

	t := make(chan struct{})

	go func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("server graceful shutdown")
		}

		log.Info().Msg("server graceful stopped")
		t <- struct{}{}
	}()

	select {
	case <-t:
		os.Exit(1)
	case <-c:
		os.Exit(1)
	case <-ctx.Done():
		os.Exit(1)
	}
}
