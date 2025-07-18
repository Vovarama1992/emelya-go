package rewardhttp

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
	mux.Handle("/api/reward/my",
		withRecoverAndRateLimit(http.HandlerFunc(handler.GetMyRewards)),
	)

	// === ADMIN ===
	mux.Handle("/api/admin/reward/referral-income",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminCreateReferralReward))),
	)

	mux.Handle("/api/admin/reward/by-user",
		withRecover(withAdminAuth(http.HandlerFunc(handler.AdminGetRewardsByUser))),
	)
}
