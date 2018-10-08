package middleware

import (
	"github.com/inconshreveable/log15"
	"net/http"
)

type APIKeyBackend interface {
	Authenticate(apiKey string) error
}

type APIKeyMiddleWare struct {
	log     log15.Logger
	backend APIKeyBackend
}

func (m *APIKeyMiddleWare) Handle(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		// Authorization: xxxxxxxxxx
		authorization, ok := r.Header["Authorization"]
		if !ok || len(authorization) != 1 {
			m.log.Warn("No or invalid Authorization header")
			w.WriteHeader(401)
			w.Write([]byte("No or invalid Authorization header"))
			return
		}
		apiKey := authorization[0]
		if apiKey == "" {
			m.log.Warn("No API Key provided")
			w.WriteHeader(401)
			return
		}
		if err := m.backend.Authenticate(apiKey); err != nil {
			m.log.Warn("Invalid API Key")
			w.WriteHeader(401)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func CreateAPIKeyMiddleWare(backend APIKeyBackend) *APIKeyMiddleWare {
	return &APIKeyMiddleWare{
		log:     log15.New("module", "middleware.apikey"),
		backend: backend,
	}
}
