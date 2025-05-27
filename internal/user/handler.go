package user

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type UpdateProfileRequest struct {
	CardNumber string `json:"card_number" example:"1234567812345678"`
}

type RequestWithdrawRequest struct {
	Amount float64 `json:"amount" example:"1500"`
}

// @Summary Обновление профиля
// @Tags user
// @Accept json
// @Produce json
// @Param data body UpdateProfileRequest true "Обновляемые поля"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/user/update-profile [post]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не разрешён"})
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.CardNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный JSON"})
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Токен отсутствует"})
		return
	}

	tokenStr := authHeader[len("Bearer "):]
	userID, err := jwtutil.ParseToken(tokenStr)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный токен"})
		return
	}

	if err := h.service.UpdateProfile(r.Context(), userID, req.CardNumber); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось обновить профиль"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Профиль обновлён"})
}

// @Summary Запрос на вывод средств
// @Tags user
// @Accept json
// @Produce json
// @Param data body RequestWithdrawRequest true "Сумма вывода"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/user/request-withdraw [post]
func (h *Handler) RequestWithdraw(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не разрешён"})
		return
	}

	var req RequestWithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Amount <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректная сумма"})
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Токен отсутствует"})
		return
	}

	userID, err := jwtutil.ParseToken(authHeader[len("Bearer "):])
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Неверный токен"})
		return
	}

	user, err := h.service.repo.GetUserByID(r.Context(), userID)
	if err != nil || user == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Пользователь не найден"})
		return
	}

	body := fmt.Sprintf("Пользователь %s %s (%s, %s) запросил вывод: %.2f ₽",
		user.FirstName, user.LastName, user.Email, user.Phone, req.Amount,
	)
	_ = h.service.notifier.SendEmailToOperator("Запрос на вывод", body)

	json.NewEncoder(w).Encode(map[string]string{"message": "Запрос на вывод отправлен"})
}
