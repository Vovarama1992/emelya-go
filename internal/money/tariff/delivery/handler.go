package tariffhttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	ports "github.com/Vovarama1992/emelya-go/internal/money/ports"
	model "github.com/Vovarama1992/emelya-go/internal/money/tariff/model"
	"github.com/go-playground/validator/v10"
)

type Handler struct {
	service ports.TariffService
}

func NewHandler(service ports.TariffService) *Handler {
	return &Handler{service: service}
}

// @Summary Получить все тарифы
// @Tags tariff
// @Produce json
// @Success 200 {array} model.Tariff
// @Failure 500 {object} map[string]string
// @Router /api/admin/tariffs [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	tariffs, err := h.service.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Не удалось получить тарифы", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tariffs)
}

// @Summary Создать тариф
// @Tags tariff
// @Accept json
// @Produce json
// @Param data body CreateTariffRequest true "Данные тарифа"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/tariffs [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTariffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		return
	}

	var blockUntil *time.Time
	if req.BlockUntil != nil {
		t, err := time.Parse(time.RFC3339, *req.BlockUntil)
		if err != nil {
			http.Error(w, "Некорректный формат даты block_until", http.StatusBadRequest)
			return
		}
		blockUntil = &t
	}

	tariff := &model.Tariff{
		Name:        req.Name,
		BlockUntil:  blockUntil,
		DailyReward: req.DailyReward,
	}

	if err := h.service.Create(r.Context(), tariff); err != nil {
		http.Error(w, "Не удалось создать тариф", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Тариф создан"})
}

// @Summary Обновить тариф
// @Tags tariff
// @Accept json
// @Produce json
// @Param data body UpdateTariffRequest true "Обновляемые данные тарифа"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/tariffs [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var req UpdateTariffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Некорректный JSON", http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http.Error(w, "Ошибка валидации", http.StatusBadRequest)
		return
	}

	var blockUntil *time.Time
	if req.BlockUntil != nil {
		t, err := time.Parse(time.RFC3339, *req.BlockUntil)
		if err != nil {
			http.Error(w, "Некорректный формат даты block_until", http.StatusBadRequest)
			return
		}
		blockUntil = &t
	}

	tariff := &model.Tariff{
		ID:          req.ID,
		Name:        req.Name,
		BlockUntil:  blockUntil,
		DailyReward: req.DailyReward,
	}

	if err := h.service.Update(r.Context(), tariff); err != nil {
		http.Error(w, "Не удалось обновить тариф", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Тариф обновлён"})
}

// @Summary Удалить тариф
// @Tags tariff
// @Produce json
// @Param id query int true "ID тарифа"
// @Success 200 {object} map[string]string
// @Failure 400,500 {object} map[string]string
// @Router /api/admin/tariffs [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	queryID := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(queryID, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "Некорректный ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, "Не удалось удалить тариф", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Тариф удалён"})
}
