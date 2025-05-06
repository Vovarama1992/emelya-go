package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/Vovarama1992/emelya-go/internal/user"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type AuthService struct {
	userRepo    user.Repository
	redisClient *redis.Client
	smsApiKey   string
	smsSender   string
}

func NewAuthService(userRepo user.Repository, redisClient *redis.Client) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		redisClient: redisClient,
		smsApiKey:   os.Getenv("SMS_API_KEY"),
		smsSender:   os.Getenv("SMS_SENDER_NAME"),
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, newUser *user.User) error {
	return s.userRepo.CreateUser(ctx, newUser)
}

func (s *AuthService) FindUserByPhone(ctx context.Context, phone string) (*user.User, error) {
	return s.userRepo.GetUserByPhone(ctx, phone)
}

func (s *AuthService) FindUserByLogin(ctx context.Context, login string) (*user.User, error) {
	return s.userRepo.GetUserByLogin(ctx, login)
}

func (s *AuthService) VerifyPhone(ctx context.Context, userID int) error {
	return s.userRepo.SetPhoneVerified(ctx, userID)
}

func (s *AuthService) SaveCodeToRedis(ctx context.Context, phone string, code string) error {
	key := fmt.Sprintf("auth_code:phone:%s", phone)
	return s.redisClient.Set(ctx, key, code, 5*time.Minute).Err()
}

func (s *AuthService) GetCodeFromRedis(ctx context.Context, phone string) (string, error) {
	key := fmt.Sprintf("auth_code:phone:%s", phone)
	return s.redisClient.Get(ctx, key).Result()
}

func (s *AuthService) SavePasswordToRedis(ctx context.Context, phone string, password string) error {
	key := fmt.Sprintf("auth_password:phone:%s", phone)
	return s.redisClient.Set(ctx, key, password, 5*time.Minute).Err()
}

func (s *AuthService) GetPasswordFromRedis(ctx context.Context, phone string) (string, error) {
	key := fmt.Sprintf("auth_password:phone:%s", phone)
	return s.redisClient.Get(ctx, key).Result()
}

func ParseToken(tokenStr string) (int, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неверный метод подписи")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, fmt.Errorf("некорректный токен")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return 0, fmt.Errorf("некорректная нагрузка токена")
	}

	return int(claims["user_id"].(float64)), nil
}

const safeCharset = "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ123456789"

// Логин: только буквы (без o, O), длина 8
func GenerateLogin() string {
	const lettersOnly = "abcdefghijklmnpqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ"
	return generateRandomString(8, lettersOnly)
}

// Пароль: только цифры (без 0), длина 6
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

func GenerateToken(userID int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
