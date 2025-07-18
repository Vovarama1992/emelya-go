package notifier

import (
	"net/http"
	"time"

	"github.com/Vovarama1992/go-utils/httputil"
)

func RegisterRoutes(mux *http.ServeMux, handler *NotifyHandler) {
	mux.Handle("/api/notify",
		httputil.RecoverMiddleware(
			httputil.NewRateLimiter(3, time.Minute)(
				http.HandlerFunc(handler.Notify),
			),
		),
	)
}
