package dav_server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	x509_keypair_reloader "github.com/mrlinqu/ltdav/internal/x509-keypair-reloader"
	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
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

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello from secure server!\n")
}
func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	return mux
}
func configureTLS(certFile, keyFile string) *tls.Config {
	return &tls.Config{
		GetCertificate: getCertificateFunc(certFile, keyFile),
		//Certificates: []tls.Certificate{cert},
		//MinVersion:   tls.VersionTLS12,
		// CipherSuites: []uint16{
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// },
		//PreferServerCipherSuites: true,
	}
}
func getCertificateFunc(certFile, keyFile string) func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		log.Println("get cert")

		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			log.Fatalf("Failed to load certificate: %v", err)
		}

		return &cert, nil
	}

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

	handler := setupRouter()
	tlsCfg := configureTLS("cert.pem", "key.pem")

	srv := &http.Server{
		Addr:      "0.0.0.0:8443",
		Handler:   handler,
		TLSConfig: tlsCfg,
		//TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		//ReadTimeout:  5 * time.Second,
		//WriteTimeout: 10 * time.Second,
		//IdleTimeout:  120 * time.Second,
	}

	s.srv = srv
	return s.srv.ListenAndServeTLS("", "")

	// s.srv = &http.Server{
	// 	Addr: s.listenAddr,
	// 	Handler: &webdav.Handler{
	// 		FileSystem: webdav.Dir(s.workingDir),
	// 		LockSystem: webdav.NewMemLS(),
	// 		Logger:     s.logger,
	// 	},
	// 	ErrorLog: log.New(zlog.Logger, "", 0),
	// 	//ErrorLog: logger.New(&fwdToZapWriter{logger}, "", 0),
	// }

	// if s.passwdFilePath != "" {
	// 	secretProvider, err := secret_provider.NewHtpasswordProvider(s.passwdFilePath)
	// 	if err != nil {
	// 		return errors.Wrap(err, "create secret_provider")
	// 	}

	// 	s.srv.Handler = http_auth.NewBasicAuthInterceptor(s.srv.Handler, secretProvider, s.realm)
	// }

	// tlsConfig, err := s.initTLS(ctx)
	// if err != nil {
	// 	return errors.Wrap(err, "initTLS")
	// }

	// if tlsConfig != nil {
	// 	s.srv.TLSConfig = tlsConfig
	// 	return s.srv.ListenAndServeTLS("", "")
	// }

	// return s.srv.ListenAndServe()
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
		//MinVersion:               tls.VersionTLS13,
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
