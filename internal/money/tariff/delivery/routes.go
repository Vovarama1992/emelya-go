package tariffhttp

import (
	"net/http"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/auth/middleware"
	user_ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
	"github.com/Vovarama1992/go-utils/httputil"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, userService user_ports.UserServiceInterface) {
	withRecoverAndRateLimit := func(h http.Handler) http.Handler {
		return httputil.RecoverMiddleware(httputil.NewRateLimiter(5, time.Minute)(h))
	}

	withAdminAuth := func(h http.Handler) http.Handler {
		return middleware.AuthMiddleware(userService, true)(h)
	}

	mux.Handle("/api/admin/tariffs", withRecoverAndRateLimit(
		withAdminAuth(http.HandlerFunc(handler.HandleTariffs)),
	))
}
