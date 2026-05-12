package repository

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"subscriptions_service/internal/models"
)

// SubscriptionRepo — репозиторий для работы с подписками в БД
type SubscriptionRepo struct {
	db *pgxpool.Pool // Пул соединений с PostgreSQL
}

// Создаём новый экземпляр репозитория
func NewSubscriptionRepo(db *pgxpool.Pool) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// Добавляем новую подписку в БД и возвращаем заполненную запись
func (r *SubscriptionRepo) Create(ctx context.Context, sub models.Subscription) (models.Subscription, error) {
	id := uuid.New()
	query := `
		INSERT INTO subscriptions (id, service_name, price, user_id, start_date, end_date)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	var created models.Subscription
	err := r.db.QueryRow(ctx, query,
		id, sub.ServiceName, sub.Price, sub.UserID, sub.StartDate, sub.EndDate,
	).Scan(&created.ID, &created.CreatedAt, &created.UpdatedAt)

	if err != nil {
		return models.Subscription{}, fmt.Errorf("failed to create subscription: %w", err)
	}

	created.ServiceName = sub.ServiceName
	created.Price = sub.Price
	created.UserID = sub.UserID
	created.StartDate = sub.StartDate
	created.EndDate = sub.EndDate
	return created, nil
}

// Получаем подписку по её ID из БД
func (r *SubscriptionRepo) GetByID(ctx context.Context, id string) (models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE id = $1
	`

	var sub models.Subscription
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
		&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
	)

	if err != nil {
		return models.Subscription{}, fmt.Errorf("subscription not found: %w", err)
	}

	return sub, nil
}

// Получаем список подписок с применением фильтров (user_id, service_name и др.)
func (r *SubscriptionRepo) List(ctx context.Context, filters map[string]string) ([]models.Subscription, error) {
	query := `
		SELECT id, service_name, price, user_id, start_date, end_date, created_at, updated_at
		FROM subscriptions WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	// Добавляем условия фильтрации по user_id
	if userID, ok := filters["user_id"]; ok && userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}

	// Добавляем фильтрацию по частичному совпадению названия сервиса
	if serviceName, ok := filters["service_name"]; ok && serviceName != "" {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argIdx)
		args = append(args, "%"+serviceName+"%")
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}
	defer rows.Close()

	var subs []models.Subscription
	for rows.Next() {
		var sub models.Subscription
		if err := rows.Scan(
			&sub.ID, &sub.ServiceName, &sub.Price, &sub.UserID,
			&sub.StartDate, &sub.EndDate, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, err
	}
		subs = append(subs, sub)
	}

	return subs, nil
}

// Обновляем данные подписки в БД и возвращаем обновлённую запись
func (r *SubscriptionRepo) Update(ctx context.Context, sub models.Subscription) (models.Subscription, error) {
	query := `
		UPDATE subscriptions
		SET service_name = $1, price = $2, start_date = $3, end_date = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING created_at, updated_at
	`

	var updated models.Subscription
	err := r.db.QueryRow(ctx, query,
		sub.ServiceName, sub.Price, sub.StartDate, sub.EndDate, sub.ID,
	).Scan(&updated.CreatedAt, &updated.UpdatedAt)

	if err != nil {
		return models.Subscription{}, fmt.Errorf("failed to update subscription: %w", err)
	}

	updated.ID = sub.ID
	updated.ServiceName = sub.ServiceName
	updated.Price = sub.Price
	updated.UserID = sub.UserID
	updated.StartDate = sub.StartDate
	updated.EndDate = sub.EndDate

	return updated, nil
}

// Удаляем подписку по ID; возвращаем ошибку, если запись не найдена
func (r *SubscriptionRepo) Delete(ctx context.Context, id string) error {
	result, err := r.db.Exec(ctx, "DELETE FROM subscriptions WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("subscription not found")
	}

	return nil
}

// Рассчитываем общую сумму подписок с учётом фильтров (user_id, даты и т.д.)
func (r *SubscriptionRepo) Sum(ctx context.Context, filters map[string]string) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	// Фильтр по пользователю
	if userID, ok := filters["user_id"]; ok && userID != "" {
		query += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, userID)
		argIdx++
	}

	// Фильтр по названию сервиса (частичное совпадение)
	if serviceName, ok := filters["service_name"]; ok && serviceName != "" {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", argIdx)
		args = append(args, "%"+serviceName+"%")
		argIdx++
	}

	// Фильтр по дате начала
	if startDate, ok := filters["start_date"]; ok && startDate != "" {
		query += fmt.Sprintf(" AND start_date >= $%d", argIdx)
		args = append(args, startDate)
		argIdx++
	}

	// Фильтр по дате окончания (с учётом NULL)
	if endDate, ok := filters["end_date"]; ok && endDate != "" {
		query += fmt.Sprintf(" AND (end_date IS NULL OR end_date >= $%d)", argIdx)
		args = append(args, endDate)
	}

	var total int
	err := r.db.QueryRow(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate sum: %w", err)
	}

	return total, nil
}
