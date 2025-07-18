package withdrawalhttp

import (
	"net/http"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/auth/middleware"
	ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
	"github.com/Vovarama1992/go-utils/httputil"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, userService ports.UserServiceInterface) {
	withRecover := func(h http.Handler) http.Handler {
		return httputil.RecoverMiddleware(h)
	}

	withRecoverAndRateLimit := func(h http.Handler) http.Handler {
		return httputil.RecoverMiddleware(httputil.NewRateLimiter(3, time.Minute)(h))
	}

	withAdminAuth := func(h http.Handler) http.Handler {
		return middleware.AuthMiddleware(userService, true)(h)
	}

	// === USER ===
	mux.Handle("/api/withdrawal/request",
		withRecoverAndRateLimit(http.HandlerFunc(handler.CreateWithdrawal)),
	)

	mux.Handle("/api/withdrawal/my",
		withRecoverAndRateLimit(http.HandlerFunc(handler.GetMyWithdrawals)),
	)

	// === ADMIN ===
	mux.Handle("/api/admin/withdrawal/all",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminGetAllWithdrawals))),
	)

	mux.Handle("/api/admin/withdrawal/pending",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminGetPendingWithdrawals))),
	)

	mux.Handle("/api/admin/withdrawal/approve",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminApproveWithdrawal))),
	)

	mux.Handle("/api/admin/withdrawal/reject",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminRejectWithdrawal))),
	)
}
