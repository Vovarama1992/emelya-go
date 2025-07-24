package user

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	operation "github.com/Vovarama1992/emelya-go/internal/money/operation_model"
	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
	model "github.com/Vovarama1992/emelya-go/internal/user/model"
	user "github.com/Vovarama1992/emelya-go/internal/user/ports"

	"github.com/go-playground/validator/v10"
)

var _ = operation.Operations{}

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

// @Summary Обновить профиль (самостоятельно)
// @Tags user
// @Accept json
// @Produce json
// @Param data body UpdateProfileRequest true "Обновляемые поля"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/user/update-profile [post]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Отсутствует токен", http.StatusUnauthorized)
		return
	}
	tokenStr := authHeader[len("Bearer "):]
	userID, err := jwtutil.ParseToken(tokenStr)
	if err != nil {
		http.Error(w, "Неверный токен", http.StatusUnauthorized)
		return
	}

	userModel := &model.User{
		ID: userID,
	}
	if req.FirstName != nil {
		userModel.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		userModel.LastName = *req.LastName
	}
	if req.Patronymic != nil {
		userModel.Patronymic = *req.Patronymic
	}
	if req.Phone != nil {
		userModel.Phone = *req.Phone
	}
	if req.CardNumber != nil {
		userModel.CardNumber = req.CardNumber
	}

	if err := h.userService.UpdateProfile(r.Context(), userModel); err != nil {
		http.Error(w, "Не удалось обновить профиль", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Профиль обновлён"})
}

// @Summary Админ обновляет профиль пользователя
// @Tags admin-user
// @Accept json
// @Produce json
// @Param data body AdminUpdateProfileRequest true "Обновляемые поля пользователя"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/admin/user/update-profile [post]
func (h *Handler) AdminUpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req AdminUpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		return
	}

	userModel := &model.User{
		ID: req.UserID,
	}
	if req.FirstName != nil {
		userModel.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		userModel.LastName = *req.LastName
	}
	if req.Patronymic != nil {
		userModel.Patronymic = *req.Patronymic
	}
	if req.Phone != nil {
		userModel.Phone = *req.Phone
	}
	if req.CardNumber != nil {
		userModel.CardNumber = req.CardNumber
	}

	if err := h.userService.UpdateProfile(r.Context(), userModel); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Профиль обновлён"})
}

// @Summary Админ добавляет реферала пользователю
// @Tags admin-user
// @Accept json
// @Produce json
// @Param data body AddReferralRequest true "ID пользователя и ID реферала"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/admin/user/add-referal [post]
func (h *Handler) AdminAddReferal(w http.ResponseWriter, r *http.Request) {
	var req AddReferralRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		return
	}

	if err := h.userService.SetReferrer(r.Context(), req.UserID, req.ReferrerID); err != nil {
		http.Error(w, "Не удалось добавить реферала", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Реферал добавлен"})
}

// @Summary Найти пользователя по ID
// @Tags user
// @Produce json
// @Param id query int true "ID пользователя"
// @Success 200 {object} model.User
// @Failure 400,404,500 {object} map[string]string
// @Router /api/admin/user/search-id [get]
func (h *Handler) AdminSearchByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || userID <= 0 {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	user, err := h.userService.FindUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Пользователь не найден", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

// @Summary Получить всех пользователей
// @Tags user
// @Produce json
// @Success 200 {array} model.User
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
// @Success 200 {object} operation.Operations
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
