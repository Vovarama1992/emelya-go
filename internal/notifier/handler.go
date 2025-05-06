package notifier

import (
	"encoding/json"
	"net/http"
)

type NotifyHandler struct {
	notifier *Notifier
}

func NewNotifyHandler(notifier *Notifier) *NotifyHandler {
	return &NotifyHandler{notifier: notifier}
}

type NotifyRequest struct {
	Text string `json:"text" example:"Имя: Иванов Иван\nТелефон: +7900...\nТариф: Премиум"`
}

// Notify godoc
// @Summary Отправка произвольного уведомления оператору
// @Tags notifier
// @Accept json
// @Produce json
// @Param data body NotifyRequest true "Текст уведомления"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/notify [post]
func (h *NotifyHandler) Notify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Метод не разрешён"})
		return
	}

	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Некорректный JSON"})
		return
	}

	if err := h.notifier.SendEmailToOperator("Уведомление от пользователя", req.Text); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Не удалось отправить письмо"})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Письмо отправлено"})
}
