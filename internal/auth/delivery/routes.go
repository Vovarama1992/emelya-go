package authadapter

import (
	"net/http"
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler) {
	// Registration
	mux.Handle("/api/auth/request-register",
		httputil.RecoverMiddleware(
			httputil.NewRateLimiter(10, time.Minute)(
				http.HandlerFunc(handler.RequestRegister),
			),
		),
	)
	mux.Handle("/api/auth/confirm-register",
		httputil.RecoverMiddleware(
			http.HandlerFunc(handler.ConfirmRegister),
		),
	)

	// Login
	mux.Handle("/api/auth/request-login",
		httputil.RecoverMiddleware(
			http.HandlerFunc(handler.RequestLogin),
		),
	)
	mux.Handle("/api/auth/confirm-login",
		httputil.RecoverMiddleware(
			http.HandlerFunc(handler.ConfirmLogin),
		),
	)
	mux.Handle("/api/auth/login-by-creds",
		httputil.RecoverMiddleware(
			http.HandlerFunc(handler.LoginByCredentials),
		),
	)

	// Current user
	mux.Handle("/api/auth/me",
		httputil.RecoverMiddleware(
			http.HandlerFunc(handler.Me),
		),
	)
}
