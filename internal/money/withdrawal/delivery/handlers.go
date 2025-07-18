package withdrawalhttp

import (
	"encoding/json"
	"net/http"

	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	service "github.com/Vovarama1992/emelya-go/internal/money/usecase"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct {
	withdrawalService *service.WithdrawalService
}

func NewHandler(withdrawalService *service.WithdrawalService) *Handler {
	return &Handler{
		withdrawalService: withdrawalService,
	}
}

// CreateWithdrawal godoc
// @Summary Юзер: создать заявку на вывод
// @Tags withdrawal
// @Accept json
// @Produce json
// @Param data body CreateWithdrawalRequest true "Данные заявки на вывод"
// @Success 200 {object} map[string]string
// @Failure 400,401,500 {object} map[string]string
// @Router /api/withdrawal/request [post]
func (h *Handler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
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

	var req CreateWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	if err := h.withdrawalService.CreateWithdrawal(r.Context(), int64(userID), req.RewardID, req.Amount); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось создать заявку")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Заявка создана"})
}

// GetMyWithdrawals godoc
// @Summary Юзер: получить свои заявки на вывод
// @Tags withdrawal
// @Produce json
// @Success 200 {array} interface{}
// @Failure 401,500 {object} map[string]string
// @Router /api/withdrawal/my [get]
func (h *Handler) GetMyWithdrawals(w http.ResponseWriter, r *http.Request) {
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

	withdrawals, err := h.withdrawalService.ListWithdrawalsByUser(r.Context(), int64(userID))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения заявок")
		return
	}

	json.NewEncoder(w).Encode(withdrawals)
}

// AdminGetAllWithdrawals godoc
// @Summary Админ: все заявки на вывод
// @Tags admin-withdrawal
// @Produce json
// @Success 200 {array} interface{}
// @Failure 500 {object} map[string]string
// @Router /api/admin/withdrawal/all [get]
func (h *Handler) AdminGetAllWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	withdrawals, err := h.withdrawalService.ListAllWithdrawals(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения заявок")
		return
	}

	json.NewEncoder(w).Encode(withdrawals)
}

// AdminGetPendingWithdrawals godoc
// @Summary Админ: заявки на вывод в статусе pending
// @Tags admin-withdrawal
// @Produce json
// @Success 200 {array} interface{}
// @Failure 500 {object} map[string]string
// @Router /api/admin/withdrawal/pending [get]
func (h *Handler) AdminGetPendingWithdrawals(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	withdrawals, err := h.withdrawalService.ListPendingWithdrawals(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения заявок")
		return
	}

	json.NewEncoder(w).Encode(withdrawals)
}

type AdminApproveWithdrawalRequest struct {
	WithdrawalID int64 `json:"withdrawal_id" validate:"required"`
}

// AdminApproveWithdrawal godoc
// @Summary Админ: подтвердить заявку на вывод
// @Tags admin-withdrawal
// @Accept json
// @Produce json
// @Param data body AdminApproveWithdrawalRequest true "ID заявки"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/withdrawal/approve [post]
func (h *Handler) AdminApproveWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req AdminApproveWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	if err := h.withdrawalService.ApproveWithdrawal(r.Context(), req.WithdrawalID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось подтвердить заявку")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Заявка подтверждена"})
}

// AdminRejectWithdrawal godoc
// @Summary Админ: отклонить заявку на вывод
// @Tags admin-withdrawal
// @Accept json
// @Produce json
// @Param data body AdminRejectWithdrawalRequest true "ID заявки и причина"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/withdrawal/reject [post]
func (h *Handler) AdminRejectWithdrawal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req AdminRejectWithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные запроса")
		return
	}

	if err := h.withdrawalService.RejectWithdrawal(r.Context(), req.WithdrawalID, req.Reason); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось отклонить заявку")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Заявка отклонена"})
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
