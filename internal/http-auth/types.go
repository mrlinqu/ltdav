package http_auth

type SecretProvider interface {
	Match(username string, password string) bool
}

const (
	HTTPHeaderAuthenticate = "WWW-Authenticate"
)
