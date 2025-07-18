package rewardhttp

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	model "github.com/Vovarama1992/emelya-go/internal/money/reward/model"
	service "github.com/Vovarama1992/emelya-go/internal/money/usecase"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct {
	rewardService *service.RewardService
}

func NewHandler(rewardService *service.RewardService) *Handler {
	return &Handler{
		rewardService: rewardService,
	}
}

// AdminCreateReferralReward godoc
// @Summary Админ: начислить доход от реферала
// @Tags admin-reward
// @Accept json
// @Produce json
// @Param data body AdminCreateReferralRewardRequest true "Данные о вознаграждении"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/reward/referral-income [post]
func (h *Handler) AdminCreateReferralReward(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req AdminCreateReferralRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	reward := &model.Reward{
		UserID: req.UserID,
		Type:   model.RewardTypeReferral,
		Amount: req.Amount,
	}

	if err := h.rewardService.Create(r.Context(), reward); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось создать вознаграждение")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Вознаграждение успешно создано"})
}

// AdminGetRewardsByUser godoc
// @Summary Админ: получить все награды пользователя
// @Tags admin-reward
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Success 200 {array} interface{}
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/reward/by-user [get]
func (h *Handler) AdminGetRewardsByUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный user_id")
		return
	}

	rewards, err := h.rewardService.FindByUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения вознаграждений")
		return
	}

	json.NewEncoder(w).Encode(rewards)
}

// GetMyRewards godoc
// @Summary Юзер: получить свои вознаграждения
// @Tags reward
// @Produce json
// @Success 200 {array} interface{}
// @Failure 401,500 {object} map[string]string
// @Router /api/reward/my [get]
func (h *Handler) GetMyRewards(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Токен отсутствует")
		return
	}
	userID, err := jwtutil.ParseToken(authHeader[len("Bearer "):])
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Неверный токен")
		return
	}

	rewards, err := h.rewardService.FindByUserID(r.Context(), int64(userID))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения вознаграждений")
		return
	}

	json.NewEncoder(w).Encode(rewards)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
