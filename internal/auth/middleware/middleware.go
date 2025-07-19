package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	models "github.com/Vovarama1992/emelya-go/internal/user/model"
	ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
)

type contextKey string

const UserContextKey contextKey = "user"

func AuthMiddleware(userService ports.UserServiceInterface, adminOnly bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Токен отсутствует", http.StatusUnauthorized)
				return
			}

			userID, err := jwtutil.ParseToken(authHeader[len("Bearer "):])
			if err != nil {
				http.Error(w, "Неверный токен", http.StatusUnauthorized)
				return
			}

			user, err := userService.FindUserByID(r.Context(), userID)
			if err != nil || user == nil {
				http.Error(w, "Пользователь не найден", http.StatusUnauthorized)
				return
			}

			if adminOnly && user.Role != models.RoleAdmin {
				http.Error(w, "Только для админов", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) *models.User {
	user, _ := ctx.Value(UserContextKey).(*models.User)
	return user
}
