package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"           // Веб-фреймворк Gin
	"github.com/jackc/pgx/v5/pgxpool"  // Драйвер PostgreSQL
	"subscriptions_service/internal/config"      // Работа с конфигурацией
	"subscriptions_service/internal/handlers"    // HTTP-обработчики
	"subscriptions_service/internal/middleware"  // Промежуточное ПО
	"subscriptions_service/internal/models"     // Модели данных
	"subscriptions_service/internal/repository" // Работа с БД
	"subscriptions_service/internal/service"    // Бизнес-логика
)

func main() {
	// Загружаем конфигурации приложения
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Настраиваем режима работы Gin (релиз/отладка)
	gin.SetMode(gin.ReleaseMode)
	if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	}

	ctx := context.Background()

	// Подключаемся к PostgreSQL
	dsn := cfg.Database.GetDSN()
	dbPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close() // Закрываем соединение при завершении

	// Проверяем подключения к БД
	if err := dbPool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("Connected to PostgreSQL")

	// Запускаем миграций БД, если включено в конфигурации
	if cfg.Database.Migrate {
		if err := models.RunMigrations(ctx, dbPool); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}
		log.Println("Migrations completed")
	}

	// Инициализируем слои приложения: репозиторий → сервис → обработчики
	repo := repository.NewSubscriptionRepo(dbPool)
	svc := service.NewSubscriptionService(repo, cfg)
	h := handlers.NewSubscriptionHandler(svc, cfg)

	// Создаём роутер Gin и добавление middleware
	r := gin.New()
	r.Use(middleware.Logger())  // Логирование запросов
	r.Use(gin.Recovery())      // Восстановление после паник

	// Группируем API-маршруты
	api := r.Group("/api/v1")
	{
		subs := api.Group("/subscriptions")
		{
			subs.GET("", h.List)       // Получаем список подписок
			subs.GET("/:id", h.Get)   // Получаем подписку по ID
			subs.POST("", h.Create)    // Создаём подписку
			subs.PUT("/:id", h.Update) // Обновляем подписку
			subs.DELETE("/:id", h.Delete) // Удаляем подписку
		}
		api.GET("/subscriptions/sum", h.Sum) // Получаем сумму подписок
	}

	// Настраиваем HTTP-сервер
	srv := &http.Server{
		Addr:    cfg.Server.Addr,
		Handler: r,
	}

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Server starting on %s", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	// Ожидаем сигналы завершения (SIGINT, SIGTERM)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Корректное завершение работы сервера
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("Server exited")
}
