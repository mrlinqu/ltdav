package dav_server

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"

	http_auth "github.com/mrlinqu/ltdav/internal/http-auth"
	secret_provider "github.com/mrlinqu/ltdav/internal/http-auth/secret-provider"
	x509_keypair_reloader "github.com/mrlinqu/ltdav/internal/x509-keypair-reloader"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/net/webdav"
)

type DavServer struct {
	listenAddr string
	workingDir string

	certPath string
	keyPath  string

	passwdFilePath string
	realm          string

	srv *http.Server
}

func New(listenAddr string, workingDir string) *DavServer {
	return &DavServer{
		listenAddr: listenAddr,
		workingDir: workingDir,
	}
}

func (s *DavServer) WithTls(certPath string, keyPath string) *DavServer {
	s.certPath = certPath
	s.keyPath = keyPath

	return s
}

func (s *DavServer) WithAuth(passwdFilePath string, realm string) *DavServer {
	s.passwdFilePath = passwdFilePath
	s.realm = realm

	return s
}

func (s *DavServer) ListenAndServe(ctx context.Context) error {
	zlog.Debug().
		Str("listenAddr", s.listenAddr).
		Str("workingDir", s.workingDir).
		Str("certPath", s.certPath).
		Str("keyPath", s.keyPath).
		Str("passwdFilePath", s.passwdFilePath).
		Str("realm", s.realm).
		Msg("starting dav server")

	s.srv = &http.Server{
		Addr: s.listenAddr,
		Handler: &webdav.Handler{
			FileSystem: webdav.Dir(s.workingDir),
			LockSystem: webdav.NewMemLS(),
			Logger:     s.logger,
		},
		ErrorLog: log.New(zlog.Logger, "", 0),
		//ErrorLog: logger.New(&fwdToZapWriter{logger}, "", 0),
	}

	if s.passwdFilePath != "" {
		secretProvider, err := secret_provider.NewHtpasswordProvider(s.passwdFilePath)
		if err != nil {
			return errors.Wrap(err, "create secret_provider")
		}

		s.srv.Handler = http_auth.NewBasicAuthInterceptor(s.srv.Handler, secretProvider, s.realm)
	}

	if s.certPath != "" || s.keyPath != "" {
		keyReloader, err := x509_keypair_reloader.New(ctx, s.certPath, s.keyPath)
		if err != nil {
			return errors.Wrap(err, "create x509_keypair_reloader")
		}

		s.srv.TLSConfig = &tls.Config{
			//MinVersion:               tls.VersionTLS13,
			GetCertificate: keyReloader.GetCertificateFunc(),
		}

		//s.srv.TLSConfig.GetCertificate = keyReloader.GetCertificateFunc()

		return s.srv.ListenAndServeTLS("", "")
	}

	return s.srv.ListenAndServe()
}

func (s *DavServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *DavServer) logger(r *http.Request, err error) {
	if err != nil {
		zlog.Error().Err(err).
			Str("URL", r.URL.String()).
			Str("Method", r.Method).
			Msg("webdav error")
	} else {
		zlog.Debug().
			Str("URL", r.URL.String()).
			Str("Method", r.Method).
			Msg("webdav debug")
	}
}
