package config

import (
	"context"
	"flag"
	"os"
	"strings"
)

const (
	WorkDir  = "work_dir"
	Addr     = "addr"
	CertFile = "cert_file"
	KeyFile  = "key_file"
	AuthFile = "auth_file"
	Realm    = "auth_realm"
)

var defaultValues = map[string]string{
	WorkDir: "./",
	Addr:    "0.0.0.0:9800",
}

const (
	envPrefix = "LTDAV_"
)

type configStorage map[string]string

func (c configStorage) Value(key string) string {
	return c[key]
}

var configContextKey = struct{}{}

func Init(ctx context.Context) context.Context {
	dir := flag.String("d", "", "Directory to serve from. Default is CWD")
	addr := flag.String("l", "", "address to listen. Default 0.0.0.0:9800")
	cert := flag.String("c", "", "Path to TLS cert file")
	key := flag.String("k", "", "Path to TLS key file")
	auth := flag.String("a", "", "Path to auth file")
	realm := flag.String("r", "", "Auth realm text")

	flag.Parse()

	cfg := configStorage{}

	cfg[WorkDir] = *dir
	cfg[Addr] = *addr
	cfg[CertFile] = *cert
	cfg[KeyFile] = *key
	cfg[AuthFile] = *auth
	cfg[Realm] = *realm

	return context.WithValue(ctx, configContextKey, cfg)
}

func GetValue(ctx context.Context, key string) string {
	stor, ok := ctx.Value(configContextKey).(configStorage)
	if !ok {
		return ""
	}

	if val := stor.Value(key); val != "" {
		return val
	}

	if val := getFromEnv(key); val != "" {
		return val
	}

	return defaultValues[key]
}

func getFromEnv(key string) string {
	k := strings.ToUpper(envPrefix + key)

	return os.Getenv(k)
}
