package deposithttp

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

	// === PUBLIC ===
	mux.Handle("/api/deposit/create",
		withRecoverAndRateLimit(http.HandlerFunc(handler.CreateDeposit)),
	)

	mux.Handle("/api/deposit/my",
		withRecover(http.HandlerFunc(handler.GetUserDeposits)),
	)

	// === ADMIN ===
	mux.Handle("/api/admin/deposit/approve",
		withRecover(withAdminAuth(http.HandlerFunc(handler.ApproveDeposit))),
	)

	mux.Handle("/api/admin/deposit/get",
		withRecover(withAdminAuth(http.HandlerFunc(handler.GetDepositByID))),
	)

	mux.Handle("/api/admin/deposit/by-user",
		withRecover(withAdminAuth(http.HandlerFunc(handler.GetDepositsByUserID))),
	)

	mux.Handle("/api/admin/deposit/close",
		withRecover(withAdminAuth(http.HandlerFunc(handler.CloseDeposit))),
	)

	mux.Handle("/api/admin/deposit/create",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminCreateDeposit))),
	)

	mux.Handle("/api/admin/deposit/delete",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminDeleteDeposit))),
	)

	mux.Handle("/api/admin/deposit/pending",
		withRecover(withAdminAuth(http.HandlerFunc(handler.ListPendingDeposits))),
	)

	mux.Handle("/api/admin/deposit/total-approved-amount", http.HandlerFunc(handler.GetTotalApprovedAmount))
}
