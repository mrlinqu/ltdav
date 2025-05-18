package http_auth

import (
	"net/http"
)

type BasicAuthInterceptor struct {
	realm          string
	secretProvider SecretProvider
	handler        http.Handler
}

func NewBasicAuthInterceptor(handler http.Handler, secretProvider SecretProvider, realm string) *BasicAuthInterceptor {
	return &BasicAuthInterceptor{
		realm:          realm,
		secretProvider: secretProvider,
		handler:        handler,
	}
}

func (a *BasicAuthInterceptor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok {
		a.unauth(w)
		return
	}

	if username == "" {
		a.unauth(w)
		return
	}

	if !a.secretProvider.Match(username, password) {
		a.unauth(w)
		return
	}

	a.handler.ServeHTTP(w, r)
}

func (a *BasicAuthInterceptor) unauth(w http.ResponseWriter) {
	w.Header().Set(HTTPHeaderAuthenticate, `Basic realm="`+a.realm+`", charset="UTF-8"`)
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}
