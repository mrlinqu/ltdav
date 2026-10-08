package dav_server

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	http_auth "github.com/mrlinqu/ltdav/internal/http-auth"
	secret_provider "github.com/mrlinqu/ltdav/internal/http-auth/secret-provider"
	x509_keypair_reloader "github.com/mrlinqu/ltdav/internal/x509-keypair-reloader"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/net/webdav"
)

var (
	handlerInternalServerError = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	})

	handlerForbidden = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
	})
)

type DavServer struct {
	listenAddr string
	workingDir string

	certPath string
	keyPath  string

	passwdFilePath string
	realm          string

	srv *http.Server

	userHandlers   map[string]http.Handler
	userHandlersMu sync.RWMutex
}

func New(listenAddr string, workingDir string) *DavServer {
	return &DavServer{
		listenAddr: listenAddr,
		workingDir: workingDir,

		userHandlers: make(map[string]http.Handler),
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

	handler, err := s.getHandler()
	if err != nil {
		return errors.Wrap(err, "create http handler")
	}

	s.srv = &http.Server{
		Addr:     s.listenAddr,
		Handler:  handler,
		ErrorLog: log.New(zlog.Logger, "", 0),
	}

	tlsConfig, err := s.initTLS(ctx)
	if err != nil {
		return errors.Wrap(err, "initTLS")
	}

	if tlsConfig != nil {
		s.srv.TLSConfig = tlsConfig
		return s.srv.ListenAndServeTLS("", "")
	}

	return s.srv.ListenAndServe()
}

func (s *DavServer) initTLS(ctx context.Context) (*tls.Config, error) {
	if s.certPath == "" || s.keyPath == "" {
		return nil, nil
	}

	keyReloader, err := x509_keypair_reloader.New(ctx, s.certPath, s.keyPath)
	if err != nil {
		return nil, errors.Wrap(err, "create x509_keypair_reloader")
	}

	return &tls.Config{
		GetCertificate: keyReloader.GetCertificateFunc(),
	}, nil
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

func (s *DavServer) getHandler() (http.Handler, error) {
	if s.passwdFilePath == "" {
		return s.getDefaultHandler(), nil
	}

	secretProvider, err := secret_provider.NewHtpasswordProvider(s.passwdFilePath)
	if err != nil {
		return nil, errors.Wrap(err, "create secret_provider")
	}

	return http_auth.NewBasicAuthInterceptor(
		s.hanlerAuth,
		secretProvider,
		s.realm,
	), nil
}

func (s *DavServer) getDefaultHandler() http.Handler {
	return &webdav.Handler{
		FileSystem: webdav.Dir(s.workingDir),
		LockSystem: webdav.NewMemLS(),
		Logger:     s.logger,
	}
}

func (s *DavServer) hanlerAuth(username string) http.Handler {
	s.userHandlersMu.RLock()

	if handler, ok := s.userHandlers[username]; ok {
		return handler
	}

	s.userHandlersMu.RUnlock()

	if username == "." || username == ".." || strings.ContainsAny(username, `/\\`) || filepath.IsAbs(username) {
		return handlerForbidden
	}

	rootDir, err := filepath.Abs(s.workingDir)
	if err != nil {
		return handlerInternalServerError
	}

	userDir := filepath.Join(rootDir, username)
	if err := os.MkdirAll(userDir, 0750); err != nil {
		return handlerInternalServerError
	}

	resolvedRoot, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		return handlerInternalServerError
	}

	resolvedUserDir, err := filepath.EvalSymlinks(userDir)
	if err != nil {
		return handlerInternalServerError
	}

	rel, err := filepath.Rel(resolvedRoot, resolvedUserDir)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return handlerInternalServerError
	}

	handler := &webdav.Handler{
		FileSystem: webdav.Dir(resolvedUserDir),
		LockSystem: webdav.NewMemLS(),
		Logger:     s.logger,
	}

	s.userHandlersMu.Lock()
	defer s.userHandlersMu.Unlock()

	s.userHandlers[username] = handler

	return handler
}
