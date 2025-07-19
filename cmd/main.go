package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Vovarama1992/emelya-go/docs"
	"github.com/Vovarama1992/emelya-go/internal/scheduler"

	authadapter "github.com/Vovarama1992/emelya-go/internal/auth/delivery"
	authusecase "github.com/Vovarama1992/emelya-go/internal/auth/usecase"
	"github.com/Vovarama1992/emelya-go/internal/db"

	deposithttp "github.com/Vovarama1992/emelya-go/internal/money/deposit/delivery"
	depositinfra "github.com/Vovarama1992/emelya-go/internal/money/deposit/infra"
	usecase "github.com/Vovarama1992/emelya-go/internal/money/usecase"

	rewardhttp "github.com/Vovarama1992/emelya-go/internal/money/reward/delivery"
	rewardinfra "github.com/Vovarama1992/emelya-go/internal/money/reward/infra"

	withdrawalhttp "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/delivery"
	withdrawalinfra "github.com/Vovarama1992/emelya-go/internal/money/withdrawal/infra"

	"github.com/Vovarama1992/emelya-go/internal/notifier"
	notifieradapter "github.com/Vovarama1992/emelya-go/internal/notifier"

	useradapter "github.com/Vovarama1992/emelya-go/internal/user/http"
	userinfra "github.com/Vovarama1992/emelya-go/internal/user/infra"
	userusecase "github.com/Vovarama1992/emelya-go/internal/user/usecase"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	// Swagger docs setup
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

	// Пользователи
	userRepo := userinfra.NewUserRepository(dbConn)
	notifierService := notifier.NewNotifier()
	userService := userusecase.NewService(userRepo, notifierService)

	// Аутентификация
	authService := authusecase.NewAuthService(userService, redisClient, notifierService)
	authHandler := authadapter.NewHandler(authService, notifierService)

	// Деньги - депозиты, награды, выводы
	depositRepo := depositinfra.NewDepositRepository(dbConn)
	rewardRepo := rewardinfra.NewRewardRepository(dbConn)
	withdrawalRepo := withdrawalinfra.NewWithdrawalRepository(dbConn)

	rewardService := usecase.NewRewardService(rewardRepo)
	depositService := usecase.NewDepositService(depositRepo, rewardService, dbConn, notifierService)
	withdrawalService := usecase.NewWithdrawalService(withdrawalRepo, rewardService, dbConn, notifierService)
	operationService := usecase.NewOperationsService(depositService, rewardService, withdrawalService)

	// CRON для начисления наград по депозитам (каждый час)
	cronScheduler := scheduler.StartDepositRewardCron(depositService)
	defer cronScheduler.Stop()

	// HTTP Handlers
	userHandler := useradapter.NewHandler(userService, notifierService, operationService)
	depositHandler := deposithttp.NewHandler(depositService)
	rewardHandler := rewardhttp.NewHandler(rewardService)
	withdrawalHandler := withdrawalhttp.NewHandler(withdrawalService)
	notifyHandler := notifieradapter.NewNotifyHandler(notifierService)

	// HTTP Routes
	mux := http.NewServeMux()

	mux.HandleFunc("/api/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	authadapter.RegisterRoutes(mux, authHandler)
	useradapter.RegisterRoutes(mux, userHandler, userService)
	notifieradapter.RegisterRoutes(mux, notifyHandler)
	deposithttp.RegisterRoutes(mux, depositHandler, userService)
	rewardhttp.RegisterRoutes(mux, rewardHandler, userService)
	withdrawalhttp.RegisterRoutes(mux, withdrawalHandler, userService)

	// Swagger UI
	mux.Handle("/api/docs/", httpSwagger.Handler(
		httpSwagger.URL("/api/docs/doc.json"),
	))

	// CORS
	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://emelia-invest.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Сервер запущен на порту %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}
