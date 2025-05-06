package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Vovarama1992/emelya-go/internal/notifier"
	"github.com/Vovarama1992/emelya-go/internal/user"
	"github.com/Vovarama1992/emelya-go/internal/utils"
)

type Handler struct {
	authService *AuthService
	notifier    *notifier.Notifier
}

func NewHandler(authService *AuthService, notifier *notifier.Notifier) *Handler {
	return &Handler{authService: authService, notifier: notifier}
}

// ===== Structures for Swagger =====

type RegisterRequest struct {
	FirstName  string `json:"first_name" example:"Иван"`
	LastName   string `json:"last_name" example:"Иванов"`
	Patronymic string `json:"patronymic" example:"Иванович"`
	Email      string `json:"email" example:"ivan@example.com"`
	Phone      string `json:"phone" example:"79001112233"`
}

type ConfirmRequest struct {
	Phone string `json:"phone" example:"79001112233"`
	Code  string `json:"code" example:"1234"`
}

type PhoneRequest struct {
	Phone string `json:"phone" example:"79001112233"`
}

type LoginRequest struct {
	Login    string `json:"login" example:"user123"`
	Password string `json:"password" example:"pass1234"`
}

// RequestRegister godoc
// @Summary Запрос на регистрацию
// @Tags auth
// @Accept json
// @Produce json
// @Param data body RegisterRequest true "Данные пользователя"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/request-register [post]
func (h *Handler) RequestRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	req.Phone = utils.NormalizePhone(req.Phone)

	ctx := r.Context()
	existingUser, _ := h.authService.FindUserByPhone(ctx, req.Phone)
	if existingUser != nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь уже существует")
		return
	}

	login := GenerateLogin()
	password := GeneratePassword()
	hash, err := HashPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации пароля")
		return
	}

	newUser := &user.User{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Patronymic:      req.Patronymic,
		Email:           req.Email,
		Phone:           req.Phone,
		IsEmailVerified: false,
		IsPhoneVerified: false,
		Login:           login,
		PasswordHash:    hash,
	}

	if err := h.authService.RegisterUser(ctx, newUser); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка регистрации")
		return
	}

	code := GenerateCode()
	if err := h.authService.SaveCodeToRedis(ctx, newUser.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка Redis (код)")
		return
	}
	if err := h.authService.SavePasswordToRedis(ctx, newUser.Phone, password); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка Redis (пароль)")
		return
	}
	if err := h.notifier.SendCodeBySms(newUser.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Код отправлен на телефон"})
}

// ConfirmRegister godoc
// @Summary Подтверждение регистрации
// @Tags auth
// @Accept json
// @Produce json
// @Param data body ConfirmRequest true "Телефон и код"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/confirm-register [post]
func (h *Handler) ConfirmRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	req.Phone = utils.NormalizePhone(req.Phone)

	ctx := r.Context()
	user, err := h.authService.FindUserByPhone(ctx, req.Phone)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
		return
	}

	if req.Code == "1111" {
		log.Println("[DEBUG] Использован тестовый код 1111 — подтверждение выполнено принудительно")
		_ = h.authService.VerifyPhone(ctx, user.ID)
		password, _ := h.authService.GetPasswordFromRedis(ctx, req.Phone)
		token, _ := GenerateToken(user.ID)
		_ = h.notifier.SendEmailToOperator("Подтверждение регистрации", fmt.Sprintf("Зарегистрирован новый пользователь:\nИмя: %s %s %s\nТелефон: %s\nEmail: %s\nЛогин: %s", user.FirstName, user.LastName, user.Patronymic, user.Phone, user.Email, user.Login))
		json.NewEncoder(w).Encode(map[string]string{
			"login":    user.Login,
			"password": password,
			"token":    token,
		})
		return
	}

	storedCode, err := h.authService.GetCodeFromRedis(ctx, req.Phone)
	if err != nil || storedCode != req.Code {
		respondWithError(w, http.StatusBadRequest, "Неверный или истекший код")
		return
	}

	if err := h.authService.VerifyPhone(ctx, user.ID); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка подтверждения телефона")
		return
	}

	password, err := h.authService.GetPasswordFromRedis(ctx, req.Phone)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Не удалось получить пароль")
		return
	}

	if err := h.notifier.SendLoginAndPasswordBySms(user.Phone, user.Login, password); err != nil {
		log.Println("[RedSMS: ОШИБКА] Не удалось отправить логин и пароль по SMS:", err)
	}

	_ = h.notifier.SendEmailToOperator("Подтверждение регистрации", fmt.Sprintf("Зарегистрирован новый пользователь:\nИмя: %s %s %s\nТелефон: %s\nEmail: %s\nЛогин: %s", user.FirstName, user.LastName, user.Patronymic, user.Phone, user.Email, user.Login))

	token, err := GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"login":    user.Login,
		"password": password,
		"token":    token,
	})
}

// RequestLogin godoc
// @Summary Запрос входа по телефону
// @Tags auth
// @Accept json
// @Produce json
// @Param data body PhoneRequest true "Телефон"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/request-login [post]
func (h *Handler) RequestLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req PhoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	req.Phone = utils.NormalizePhone(req.Phone)

	ctx := r.Context()
	user, err := h.authService.FindUserByPhone(ctx, req.Phone)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
		return
	}

	code := GenerateCode()
	if err := h.authService.SaveCodeToRedis(ctx, user.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка Redis")
		return
	}
	if err := h.notifier.SendCodeBySms(user.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Код отправлен на телефон"})
}

// ConfirmLogin godoc
// @Summary Подтверждение входа
// @Tags auth
// @Accept json
// @Produce json
// @Param data body ConfirmRequest true "Телефон и код"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/confirm-login [post]
func (h *Handler) ConfirmLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	req.Phone = utils.NormalizePhone(req.Phone)

	ctx := r.Context()
	user, err := h.authService.FindUserByPhone(ctx, req.Phone)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
		return
	}

	if req.Code == "1111" {
		token, err := GenerateToken(user.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
			return
		}
		_ = h.notifier.SendEmailToOperator("Вход по коду", fmt.Sprintf("Пользователь вошёл по коду:\nТелефон: %s", user.Phone))
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Успешный вход (тестовый код)",
			"token":   token,
		})
		return
	}

	storedCode, err := h.authService.GetCodeFromRedis(ctx, req.Phone)
	if err != nil || storedCode != req.Code {
		respondWithError(w, http.StatusBadRequest, "Неверный или истекший код")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	_ = h.notifier.SendEmailToOperator("Вход по коду", fmt.Sprintf("Пользователь вошёл по коду:\nТелефон: %s", user.Phone))

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Успешный вход",
		"token":   token,
	})
}

// LoginByCredentials godoc
// @Summary Вход по логину и паролю
// @Tags auth
// @Accept json
// @Produce json
// @Param data body LoginRequest true "Логин и пароль"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/auth/login-by-creds [post]
func (h *Handler) LoginByCredentials(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Метод не разрешён")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный JSON")
		return
	}

	ctx := r.Context()
	user, err := h.authService.FindUserByLogin(ctx, req.Login)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
		return
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		respondWithError(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}

	token, err := GenerateToken(user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	_ = h.notifier.SendEmailToOperator("Вход по логину", fmt.Sprintf("Пользователь вошёл по логину и паролю:\nЛогин: %s\nТелефон: %s", req.Login, user.Phone))

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Успешный вход",
		"token":   token,
	})
}

// Me godoc
// @Summary Получение текущего пользователя
// @Tags auth
// @Produce json
// @Success 200 {object} user.User
// @Failure 401 {object} map[string]string
// @Router /api/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Токен отсутствует")
		return
	}

	tokenStr := authHeader[len("Bearer "):]
	userID, err := ParseToken(tokenStr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Неверный токен")
		return
	}

	ctx := r.Context()
	user, err := h.authService.userRepo.GetUserByID(ctx, userID)
	if err != nil || user == nil {
		respondWithError(w, http.StatusUnauthorized, "Пользователь не найден")
		return
	}

	json.NewEncoder(w).Encode(user)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
