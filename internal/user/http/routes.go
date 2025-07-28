package user

import (
	"net/http"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/auth/middleware"
	user_ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
	"github.com/Vovarama1992/go-utils/httputil"
)

func RegisterRoutes(mux *http.ServeMux, handler *Handler, userService user_ports.UserServiceInterface) {
	withRecover := func(h http.Handler) http.Handler {
		return httputil.RecoverMiddleware(h)
	}

	withRecoverAndRateLimit := func(h http.Handler) http.Handler {
		return httputil.RecoverMiddleware(httputil.NewRateLimiter(5, time.Minute)(h))
	}

	withUserAuth := func(h http.Handler) http.Handler {
		return middleware.AuthMiddleware(userService, false)(h)
	}

	withAdminAuth := func(h http.Handler) http.Handler {
		return middleware.AuthMiddleware(userService, true)(h)
	}

	// === USER ===
	mux.Handle("/api/user/update-profile",
		withRecoverAndRateLimit(http.HandlerFunc(handler.UpdateProfile)),
	)

	mux.Handle("/api/user/balance",
		withRecover(withUserAuth(http.HandlerFunc(handler.GetUserBalance))),
	)

	mux.Handle("/api/user/reward-balance",
		withRecover(withUserAuth(http.HandlerFunc(handler.GetUserRewardBalance))),
	)

	// === ADMIN ===
	mux.Handle("/api/admin/user/search-id",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminSearchByID))),
	)
	mux.Handle("/api/admin/user/all",
		withRecover(withAdminAuth(http.HandlerFunc(handler.GetAllUsers))),
	)

	mux.Handle("/api/admin/user/operations",
		withRecover(withAdminAuth(http.HandlerFunc(handler.GetUserOperations))),
	)

	mux.Handle("/api/admin/user/update-profile",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminUpdateProfile))),
	)

	mux.Handle("/api/admin/user/add-referal",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminAddReferal))),
	)
}
