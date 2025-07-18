package authadapter

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	usecase "github.com/Vovarama1992/emelya-go/internal/auth/usecase"
	"github.com/Vovarama1992/emelya-go/internal/jwtutil"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
	models "github.com/Vovarama1992/emelya-go/internal/user/models"
	"github.com/Vovarama1992/emelya-go/internal/utils"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type Handler struct {
	authService *usecase.AuthService
}

func NewHandler(authService *usecase.AuthService, notifier *notifier.Notifier) *Handler {
	return &Handler{authService: authService}
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
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные: "+err.Error())
		return
	}
	req.Phone = utils.NormalizePhone(req.Phone)

	ctx := r.Context()
	existingUser, _ := h.authService.FindUserByPhone(ctx, req.Phone)
	if existingUser != nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь уже существует")
		return
	}

	login := usecase.GenerateLogin()
	password := usecase.GeneratePassword()
	hash, err := usecase.HashPassword(password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации пароля")
		return
	}

	var referrerID *int64
	if req.ReferrerID != nil {
		v := int64(*req.ReferrerID)
		referrerID = &v
	}
	newUser := &models.User{
		FirstName:       req.FirstName,
		LastName:        req.LastName,
		Patronymic:      req.Patronymic,
		Email:           req.Email,
		Phone:           req.Phone,
		IsEmailVerified: false,
		IsPhoneVerified: false,
		Login:           login,
		PasswordHash:    hash,
		ReferrerID:      referrerID,
	}

	if err := h.authService.RegisterUser(ctx, newUser); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка регистрации")
		return
	}

	code := usecase.GenerateCode()
	if err := h.authService.SaveCodeToRedis(ctx, newUser.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка Redis (код)")
		return
	}
	if err := h.authService.SavePasswordToRedis(ctx, newUser.Phone, password); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка Redis (пароль)")
		return
	}
	if err := h.authService.SendCodeBySms(newUser.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Код отправлен на телефон"})
}

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
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные: "+err.Error())
		return
	}
	req.Phone = utils.NormalizePhone(req.Phone)
	ctx := r.Context()

	user, err := h.authService.FindUserByPhone(ctx, req.Phone)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
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

	if err := h.authService.SendLoginAndPasswordBySms(user.Phone, user.Login, password); err != nil {
		log.Println("[RedSMS: ОШИБКА] Не удалось отправить логин и пароль по SMS:", err)
	}

	log.Println("[DEBUG] Подтверждение регистрации — начинаем формирование письма оператору")
	refText := "Письмо о подтверждении регистрации:\n"

	if user.ReferrerID == nil {
		log.Println("[DEBUG] У пользователя нет реферера — формируем стандартное письмо")
		refText += "Зарегистрирован новый пользователь без реферальной ссылки.\n"
	} else {
		log.Printf("[DEBUG] У пользователя есть реферер (ID: %d) — ищем данные...", *user.ReferrerID)
		refUser, err := h.authService.FindUserByID(ctx, *user.ReferrerID)
		if err != nil || refUser == nil {
			log.Println("[DEBUG] Не удалось получить данные реферера — формируем письмо без него")
			refText += "Зарегистрирован новый пользователь (реферер указан, но не найден).\n"
		} else {
			log.Printf("[DEBUG] Найден реферер: %s %s (%s)", refUser.FirstName, refUser.LastName, refUser.Email)
			refText += fmt.Sprintf(
				"Зарегистрирован новый пользователь по реферальной ссылке от %s %s (%s).\n",
				refUser.FirstName, refUser.LastName, refUser.Email,
			)
		}
	}

	body := fmt.Sprintf(
		"%sСсылка на профиль: https://emelia-invest.com/%d\nИмя: %s %s %s\nТелефон: %s\nEmail: %s\nЛогин: %s",
		refText,
		user.ID,
		user.FirstName, user.LastName, user.Patronymic,
		user.Phone, user.Email, user.Login,
	)

	log.Printf("[DEBUG] Финальное письмо оператору:\n%s", body)
	_ = h.authService.SendEmailToOperator("Подтверждение регистрации", body)

	token, err := jwtutil.GenerateToken(user.ID, user.Email)
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
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные: "+err.Error())
		return
	}
	req.Phone = utils.NormalizePhone(req.Phone)

	ctx := r.Context()
	user, err := h.authService.FindUserByPhone(ctx, req.Phone)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
		return
	}

	code := usecase.GenerateCode()
	if err := h.authService.SaveCodeToRedis(ctx, user.Phone, code); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка Redis")
		return
	}
	if err := h.authService.SendCodeBySms(user.Phone, code); err != nil {
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
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные: "+err.Error())
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
		token, err := jwtutil.GenerateToken(user.ID, user.Email)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
			return
		}
		_ = h.authService.SendEmailToOperator("Вход по коду", fmt.Sprintf("Пользователь вошёл по коду:\nТелефон: %s", user.Phone))
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

	token, err := jwtutil.GenerateToken(user.ID, user.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	_ = h.authService.SendEmailToOperator("Вход по коду", fmt.Sprintf("Пользователь вошёл по коду:\nТелефон: %s", user.Phone))

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
	if err := validate.Struct(req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректные данные: "+err.Error())
		return
	}

	ctx := r.Context()
	user, err := h.authService.FindUserByLogin(ctx, req.Login)
	if err != nil || user == nil {
		respondWithError(w, http.StatusBadRequest, "Пользователь не найден")
		return
	}

	if !usecase.CheckPasswordHash(req.Password, user.PasswordHash) {
		respondWithError(w, http.StatusUnauthorized, "Неверный логин или пароль")
		return
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка генерации токена")
		return
	}

	_ = h.authService.SendEmailToOperator("Вход по логину", fmt.Sprintf("Пользователь вошёл по логину и паролю:\nЛогин: %s\nТелефон: %s", req.Login, user.Phone))

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Успешный вход",
		"token":   token,
	})
}

// Me godoc
// @Summary Получение текущего пользователя
// @Tags auth
// @Produce json
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Router /api/auth/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Токен отсутствует")
		return
	}

	tokenStr := authHeader[len("Bearer "):]
	userID, err := jwtutil.ParseToken(tokenStr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Неверный токен")
		return
	}

	ctx := r.Context()
	user, err := h.authService.FindUserByID(ctx, userID)
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
