package x509_keypair_reloader

import (
	"context"
	"crypto/tls"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/rs/zerolog/log"
)

type X509KeypairReloader struct {
	keyPath  string
	certPath string
	pair     *tls.Certificate

	mu sync.RWMutex
}

func New(ctx context.Context, certPath string, keyPath string) (*X509KeypairReloader, error) {
	reloader := &X509KeypairReloader{
		certPath: certPath,
		keyPath:  keyPath,
	}

	err := reloader.reload()
	if err != nil {
		return nil, err
	}

	reloader.startWatch(ctx)

	return reloader, nil
}

func (kp *X509KeypairReloader) startWatch(ctx context.Context) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP)

	go kp.watch(ctx, ch)
}

func (kp *X509KeypairReloader) watch(ctx context.Context, ch <-chan os.Signal) {
	for {
		select {
		case <-ch:
			err := kp.reload()
			if err != nil {
				log.Error().
					Err(err).
					Msg("Keeping old TLS certificate because the new one could not be loaded")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (kp *X509KeypairReloader) reload() error {
	log.Debug().
		Str("certPath", kp.certPath).
		Str("keyPath", kp.keyPath).
		Msg("Reloading TLS certificate and key")

	pair, err := tls.LoadX509KeyPair(kp.certPath, kp.keyPath)
	if err != nil {
		return err
	}

	kp.mu.Lock()
	kp.pair = &pair
	kp.mu.Unlock()

	return nil
}

func (kp *X509KeypairReloader) GetCertificateFunc() func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	return func(*tls.ClientHelloInfo) (*tls.Certificate, error) {
		kp.mu.RLock()
		defer kp.mu.RUnlock()

		return kp.pair, nil
	}
}
