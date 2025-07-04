package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Vovarama1992/emelya-go/docs"
	"github.com/Vovarama1992/emelya-go/internal/auth"
	"github.com/Vovarama1992/emelya-go/internal/db"
	"github.com/Vovarama1992/emelya-go/internal/notifier"
	"github.com/Vovarama1992/emelya-go/internal/user"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Настраиваем swagger
	docs.SwaggerInfo.Title = "Emelya API"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Description = "API для регистрации, логина и управления пользователями"
	docs.SwaggerInfo.Host = "emelia-invest.com"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"https"}

	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	dbConn, err := db.New()
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer dbConn.Pool.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_ADDR"),
	})
	defer redisClient.Close()

	userRepo := user.NewPostgresRepository(dbConn)
	notifierService := notifier.NewNotifier()
	authService := auth.NewAuthService(userRepo, redisClient)
	authHandler := auth.NewHandler(authService, notifierService)
	userService := user.NewService(userRepo, notifierService)
	userHandler := user.NewHandler(userService)
	notifyHandler := notifier.NewNotifyHandler(notifierService)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	// Auth endpoints
	mux.HandleFunc("/api/auth/request-register", authHandler.RequestRegister)
	mux.HandleFunc("/api/auth/confirm-register", authHandler.ConfirmRegister)
	mux.HandleFunc("/api/auth/request-login", authHandler.RequestLogin)
	mux.HandleFunc("/api/auth/confirm-login", authHandler.ConfirmLogin)
	mux.HandleFunc("/api/auth/login-by-creds", authHandler.LoginByCredentials)
	mux.HandleFunc("/api/auth/me", authHandler.Me)

	// Notifier endpoint
	mux.HandleFunc("/api/notify", notifyHandler.Notify)

	// User endpoints
	mux.HandleFunc("/api/user/update-profile", userHandler.UpdateProfile)
	mux.HandleFunc("/api/user/request-withdraw", userHandler.RequestWithdraw)
	mux.HandleFunc("/api/user/all", userHandler.GetAllUsers) // добавлен маршрут получения всех пользователей

	// Swagger UI
	mux.Handle("/api/docs/", httpSwagger.Handler(
		httpSwagger.URL("/api/docs/doc.json"),
	))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://emelia-invest.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := c.Handler(mux)

	fmt.Printf("Сервер запущен на порту %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
