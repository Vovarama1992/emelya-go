package auth_usecase

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/notifier"
	model "github.com/Vovarama1992/emelya-go/internal/user/model"
	ports "github.com/Vovarama1992/emelya-go/internal/user/ports"
	"github.com/Vovarama1992/go-utils/ctxutil"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	UserService ports.UserServiceInterface
	redisClient *redis.Client
	smsApiKey   string
	smsSender   string
	notifier    *notifier.Notifier
}

func NewAuthService(userService ports.UserServiceInterface, redisClient *redis.Client, notifier *notifier.Notifier) *AuthService {
	return &AuthService{
		UserService: userService,
		redisClient: redisClient,
		smsApiKey:   os.Getenv("SMS_API_KEY"),
		smsSender:   os.Getenv("SMS_SENDER_NAME"),
		notifier:    notifier,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, newUser *model.User) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 2)
	defer cancel()
	return s.UserService.CreateUser(ctx, newUser)
}

func (s *AuthService) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return s.UserService.FindUserByID(ctx, userID)
}

func (s *AuthService) FindUserByPhone(ctx context.Context, phone string) (*model.User, error) {
	return s.UserService.FindUserByPhone(ctx, phone)
}

func (s *AuthService) FindUserByLogin(ctx context.Context, login string) (*model.User, error) {
	return s.UserService.FindUserByLogin(ctx, login)
}

func (s *AuthService) VerifyPhone(ctx context.Context, userID int64) error {
	return s.UserService.VerifyPhone(ctx, userID)
}

func (s *AuthService) SaveCodeToRedis(ctx context.Context, phone string, code string) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 1)
	defer cancel()
	key := fmt.Sprintf("auth_code:phone:%s", phone)
	return s.redisClient.Set(ctx, key, code, 5*time.Minute).Err()
}

func (s *AuthService) GetCodeFromRedis(ctx context.Context, phone string) (string, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 1)
	defer cancel()
	key := fmt.Sprintf("auth_code:phone:%s", phone)
	return s.redisClient.Get(ctx, key).Result()
}

func (s *AuthService) SavePasswordToRedis(ctx context.Context, phone string, password string) error {
	ctx, cancel := ctxutil.WithTimeout(ctx, 1)
	defer cancel()
	key := fmt.Sprintf("auth_password:phone:%s", phone)
	return s.redisClient.Set(ctx, key, password, 5*time.Minute).Err()
}

func (s *AuthService) GetPasswordFromRedis(ctx context.Context, phone string) (string, error) {
	ctx, cancel := ctxutil.WithTimeout(ctx, 1)
	defer cancel()
	key := fmt.Sprintf("auth_password:phone:%s", phone)
	return s.redisClient.Get(ctx, key).Result()
}

func (s *AuthService) SendCodeBySms(phone string, code string) error {
	return s.notifier.SendCodeBySms(phone, code)
}

func (s *AuthService) SendLoginAndPasswordBySms(phone string, login, password string) error {
	return s.notifier.SendLoginAndPasswordBySms(phone, login, password)
}

func (s *AuthService) SendEmailToOperator(subject, body string) error {
	return s.notifier.SendEmailToOperator(subject, body)
}

// =====================
// Генерация вспомогательных данных
// =====================

func GenerateLogin() string {
	const lettersOnly = "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ"
	return generateRandomString(8, lettersOnly)
}

func GeneratePassword() string {
	const digitsOnly = "123456789"
	return generateRandomString(6, digitsOnly)
}

func generateRandomString(length int, charset string) string {
	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func GenerateCode() string {
	num, _ := rand.Int(rand.Reader, big.NewInt(10000))
	return fmt.Sprintf("%04d", num.Int64())
}
