package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mrlinqu/ltdav/internal/config"
	x509_keypair_reloader "github.com/mrlinqu/ltdav/internal/x509-keypair-reloader"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/webdav"
)

func main() {
	ctx := config.Init(context.Background())

	workingDir := config.GetValue(ctx, config.WorkDir)
	listenAddr := config.GetValue(ctx, config.Addr)

	if workingDir == "" {
		log.Panic().Msg("working dir is not defined")
	}

	if listenAddr == "" {
		log.Panic().Msg("listen addr is not defined")
	}

	// if s.AuthBasicUserFile != "" {
	// 	sp, err := secret_provider.NewHtpasswordProvider(s.AuthBasicUserFile)
	// 	if err != nil {
	// 		return errors.Wrap(err, "htpassword error")
	// 	}

	// 	handler = http_auth.NewBasicAuthInterceptor(handler, sp, "")
	// }

	srv := &http.Server{
		Addr: listenAddr,
		Handler: &webdav.Handler{
			FileSystem: webdav.Dir(workingDir),
			LockSystem: webdav.NewMemLS(),
			Logger:     logger,
		},
	}

	tlsCertPath := config.GetValue(ctx, config.CertPath)
	tlsKeyPath := config.GetValue(ctx, config.KeyPath)

	tlsEnable := tlsCertPath != "" && tlsKeyPath != ""

	if tlsEnable {
		reloader, err := x509_keypair_reloader.New(ctx, tlsCertPath, tlsKeyPath)
		if err != nil {
			log.Panic().Err(err).Msg("create x509 keypair reloader")
		}

		srv.TLSConfig.GetCertificate = reloader.GetCertificateFunc()
	}

	go func() {
		var err error

		if tlsEnable {
			err = srv.ListenAndServeTLS("", "")
		} else {
			err = srv.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("server listen")
		}
	}()

	log.Info().Msg("server started")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
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

func logger(r *http.Request, err error) {
	if err != nil {
		log.Error().Err(err).
			Str("URL", r.URL.String()).
			Str("Method", r.Method).
			Msg("webdav error")
	} else {
		log.Debug().
			Str("URL", r.URL.String()).
			Str("Method", r.Method).
			Msg("webdav debug")
	}
}
