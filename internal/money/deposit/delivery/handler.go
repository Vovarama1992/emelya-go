package deposithttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	service "github.com/Vovarama1992/emelya-go/internal/money/usecase"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct {
	depositService *service.DepositService
}

func NewHandler(depositService *service.DepositService) *Handler {
	return &Handler{
		depositService: depositService,
	}
}

// CreateDeposit godoc
// @Summary Создать заявку на депозит
// @Tags deposit
// @Accept json
// @Produce json
// @Param data body DepositCreateRequest true "Сумма депозита"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/deposit/create [post]
func (h *Handler) CreateDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
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

	var req DepositCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	if err := h.depositService.CreateDeposit(r.Context(), int64(userID), req.Amount); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось создать депозит")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Заявка на депозит создана"})
}

// ApproveDeposit godoc
// @Summary Одобрить депозит
// @Tags deposit
// @Accept json
// @Produce json
// @Param id query int true "ID депозита"
// @Param approved_at query string true "Дата одобрения в формате RFC3339"
// @Param block_until query string true "Дата блокировки в формате RFC3339"
// @Param daily_reward query number true "Дневная награда"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/deposit/approve [post]
func (h *Handler) ApproveDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	idStr := r.URL.Query().Get("id")
	approvedAtStr := r.URL.Query().Get("approved_at")
	blockUntilStr := r.URL.Query().Get("block_until")
	dailyRewardStr := r.URL.Query().Get("daily_reward")
	tariffIDStr := r.URL.Query().Get("tariff_id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный id")
		return
	}

	approvedAt, err := time.Parse(time.RFC3339, approvedAtStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный approved_at")
		return
	}

	var blockUntil *time.Time
	if blockUntilStr != "" {
		t, err := time.Parse(time.RFC3339, blockUntilStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Некорректный block_until")
			return
		}
		blockUntil = &t
	}

	var dailyReward *float64
	if dailyRewardStr != "" {
		v, err := strconv.ParseFloat(dailyRewardStr, 64)
		if err != nil || v <= 0 {
			respondWithError(w, http.StatusBadRequest, "Некорректный daily_reward")
			return
		}
		dailyReward = &v
	}

	var tariffID *int64
	if tariffIDStr != "" {
		v, err := strconv.ParseInt(tariffIDStr, 10, 64)
		if err != nil || v <= 0 {
			respondWithError(w, http.StatusBadRequest, "Некорректный tariff_id")
			return
		}
		tariffID = &v
	}

	if err := h.depositService.ApproveDeposit(r.Context(), id, approvedAt, blockUntil, dailyReward, tariffID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось одобрить депозит")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Депозит одобрен"})
}

// GetDepositByID godoc
// @Summary Получить депозит по ID
// @Tags deposit
// @Produce json
// @Param id query int true "ID депозита"
// @Success 200 {object} interface{}
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/deposit/get [get]
func (h *Handler) GetDepositByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный id")
		return
	}

	deposit, err := h.depositService.GetDepositByID(r.Context(), id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения депозита")
		return
	}

	json.NewEncoder(w).Encode(deposit)
}

// GetUserDeposits godoc
// @Summary Получить все депозиты пользователя
// @Tags deposit
// @Produce json
// @Success 200 {array} interface{}
// @Failure 401,500 {object} map[string]string
// @Router /api/deposit/my [get]
func (h *Handler) GetUserDeposits(w http.ResponseWriter, r *http.Request) {
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

	deposits, err := h.depositService.GetDepositsByUserID(r.Context(), int64(userID))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения депозитов")
		return
	}

	json.NewEncoder(w).Encode(deposits)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// GetDepositsByUserID godoc
// @Summary Получить все депозиты по user_id (только для админа)
// @Tags deposit
// @Produce json
// @Param user_id query int true "ID пользователя"
// @Success 200 {array} interface{}
// @Failure 400,401,403,500 {object} map[string]string
// @Router /api/admin/deposit/by-user [get]
func (h *Handler) GetDepositsByUserID(w http.ResponseWriter, r *http.Request) {
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

	deposits, err := h.depositService.GetDepositsByUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения депозитов")
		return
	}

	json.NewEncoder(w).Encode(deposits)
}

// CloseDeposit godoc
// @Summary Закрыть депозит
// @Tags deposit
// @Accept json
// @Produce json
// @Param id query int true "ID депозита"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/deposit/close [post]
func (h *Handler) CloseDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный id")
		return
	}

	if err := h.depositService.CloseDeposit(r.Context(), id); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось закрыть депозит")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Депозит закрыт"})
}

// AdminCreateDeposit godoc
// @Summary Админ: создать депозит вручную
// @Tags admin-deposit
// @Accept json
// @Produce json
// @Param user_id query int true "ID инвестора"
// @Param data body AdminCreateDepositRequest true "Данные депозита"
// @Success 200 {object} map[string]interface{}
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/deposit/create [post]
func (h *Handler) AdminCreateDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный user_id")
		return
	}

	var req AdminCreateDepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	createdAt, err := time.Parse(time.RFC3339, req.CreatedAt)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный created_at")
		return
	}

	var approvedAt *time.Time
	if req.ApprovedAt != "" {
		t, err := time.Parse(time.RFC3339, req.ApprovedAt)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Некорректный approved_at")
			return
		}
		approvedAt = &t
	}

	var blockUntil *time.Time
	if req.BlockUntil != "" {
		t, err := time.Parse(time.RFC3339, req.BlockUntil)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Некорректный block_until")
			return
		}
		blockUntil = &t
	}

	var dailyReward *float64
	if req.DailyReward != nil {
		dailyReward = req.DailyReward
	}

	var tariffID *int64
	if req.TariffID != nil {
		tariffID = req.TariffID
	}

	id, err := h.depositService.CreateDepositByAdmin(
		r.Context(),
		userID,
		req.Amount,
		createdAt,
		approvedAt,
		blockUntil,
		dailyReward,
		tariffID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось создать депозит")
		return
	}

	json.NewEncoder(w).Encode(map[string]any{"message": "Депозит создан", "deposit_id": id})
}

// AdminDeleteDeposit godoc
// @Summary Админ: удалить депозит (ручное удаление)
// @Tags admin-deposit
// @Accept json
// @Produce json
// @Param deposit_id query int true "ID депозита"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/deposit/delete [delete]
func (h *Handler) AdminDeleteDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	depositIDStr := r.URL.Query().Get("deposit_id")
	depositID, err := strconv.ParseInt(depositIDStr, 10, 64)
	if err != nil || depositID <= 0 {
		respondWithError(w, http.StatusBadRequest, "Некорректный deposit_id")
		return
	}

	if err := h.depositService.DeleteDepositByAdmin(r.Context(), depositID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось удалить депозит")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Депозит удалён"})
}

// ListPendingDeposits godoc
// @Summary Админ: получить все депозиты в статусе pending
// @Tags admin-deposit
// @Produce json
// @Success 200 {array} interface{}
// @Failure 500 {object} map[string]string
// @Router /api/admin/deposit/pending [get]
func (h *Handler) ListPendingDeposits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	deposits, err := h.depositService.ListPendingDeposits(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения депозитов")
		return
	}

	json.NewEncoder(w).Encode(deposits)
}

// GetTotalApprovedAmount godoc
// @Summary Получить общую сумму одобренных депозитов
// @Tags admin-deposit
// @Produce json
// @Success 200 {object} map[string]float64
// @Failure 500 {object} map[string]string
// @Router /api/admin/deposit/total-approved-amount [get]
func (h *Handler) GetTotalApprovedAmount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	total, err := h.depositService.GetTotalApprovedAmount(r.Context())
	if err != nil {
		http.Error(w, "Failed to get total approved amount", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]float64{"total_approved_amount": total})
}
