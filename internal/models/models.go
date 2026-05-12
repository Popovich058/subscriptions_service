package models

import (
	"context"
	"errors"
	"time"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Subscription — модель данных подписки в системе
type Subscription struct {
	ID          string    `json:"id"`                 // Уникальный идентификатор (UUID)
	ServiceName string    `json:"service_name"`       // Название сервиса подписки
	Price       int       `json:"price"`              // Стоимость подписки
	UserID      string    `json:"user_id"`            // ID пользователя
	StartDate   string    `json:"start_date"`         // Дата начала подписки
	EndDate     *string   `json:"end_date,omitempty"` // Дата окончания 
	CreatedAt   time.Time `json:"created_at"`         // Время создания записи
	UpdatedAt   time.Time `json:"updated_at"`         // Время последнего обновления
}

// CreateSubscriptionRequest — структура для создания новой подписки (данные из запроса)
type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required"`    // Название сервиса (обязательно)
	Price       int     `json:"price" binding:"required,min=0"`     // Стоимость (обязательна, ≥ 0)
	UserID      string  `json:"user_id" binding:"required,uuid"`    // ID пользователя (UUID, обязателен)
	StartDate   string  `json:"start_date" binding:"required"`      // Дата начала (обязательна)
	EndDate     *string `json:"end_date,omitempty"`                 // Дата окончания (опционально)
}

// UpdateSubscriptionRequest — структура для обновления подписки 
type UpdateSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`    // Новое название сервиса
	Price       int     `json:"price"`           // Новая стоимость
	StartDate   string  `json:"start_date"`      // Новая дата начала
	EndDate     *string `json:"end_date"`        // Новая дата окончания
}

// Выполняем миграции БД: создаём таблицу subscriptions и индексы
func RunMigrations(ctx context.Context, db *pgxpool.Pool) error {
	query := `
	CREATE TABLE IF NOT EXISTS subscriptions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		service_name VARCHAR(255) NOT NULL,
		price INTEGER NOT NULL CHECK (price >= 0),
		user_id UUID NOT NULL,
		start_date VARCHAR(7) NOT NULL,
		end_date VARCHAR(7),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
	CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);
	`

	_, err := db.Exec(ctx, query)
	if err != nil {
		return errors.New("failed to run migrations: " + err.Error())
	}
	return nil
}
