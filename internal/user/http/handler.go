package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
	user "github.com/Vovarama1992/emelya-go/internal/user/ports"
)

type Handler struct {
	userService       user.UserServiceInterface
	notifier          notifier.NotifierInterface
	operationsService ports.OperationsService
}

func NewHandler(
	service user.UserServiceInterface,
	notifier notifier.NotifierInterface,
	operations ports.OperationsService,
) *Handler {
	return &Handler{
		userService:       service,
		notifier:          notifier,
		operationsService: operations,
	}
}

type UpdateProfileRequest struct {
	CardNumber *string  `json:"card_number,omitempty"`
	Balance    *float64 `json:"balance,omitempty"`
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	// обновляем поля если они пришли
	if req.CardNumber != nil {
		if err := h.userService.UpdateCardNumber(r.Context(), userID, *req.CardNumber); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось обновить номер карты"})
			return
		}
	}
	if req.Balance != nil {
		if err := h.userService.UpdateBalance(r.Context(), userID, *req.Balance); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось обновить баланс"})
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Профиль обновлён"})
}

// @Summary Получить всех пользователей
// @Tags user
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /api/admin/user/all [get]
func (h *Handler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не разрешён"})
		return
	}

	users, err := h.userService.GetAllUsers(r.Context())
	if err != nil {
		log.Printf("Ошибка GetAllUsers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось получить пользователей"})
		return
	}

	json.NewEncoder(w).Encode(users)
}

// @Summary Админ: все операции пользователя (депозиты, выводы, награды)
// @Tags admin-user
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Success 200 {object} money_usecase.Operations
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/user/operations [get]
func (h *Handler) GetUserOperations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не разрешён"})
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный user_id"})
		return
	}

	ops, err := h.operationsService.ListUserOperations(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось получить операции"})
		return
	}

	json.NewEncoder(w).Encode(ops)
}
