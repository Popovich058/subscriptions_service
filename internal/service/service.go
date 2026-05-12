package service

import (
	"context"
	"errors"
	"subscriptions_service/internal/config"
	"subscriptions_service/internal/models"
	"subscriptions_service/internal/repository"
)

// Ошибки 
var (
	ErrInvalidInput = errors.New("invalid input")          // Некорректные входные данные
	ErrNotFound     = errors.New("subscription not found") // Подписка не найдена
)

// SubscriptionService — сервис бизнес-логики для работы с подписками
type SubscriptionService struct {
	repo *repository.SubscriptionRepo // Репозиторий для работы с БД
	cfg  *config.Config               // Конфигурация приложения
}

// Создаём новый экземпляр сервиса подписок
func NewSubscriptionService(repo *repository.SubscriptionRepo, cfg *config.Config) *SubscriptionService {
	return &SubscriptionService{repo: repo, cfg: cfg}
}

// Создаём новую подписку: проверяем входные данные и передаёт в репозиторий
func (s *SubscriptionService) Create(ctx context.Context, req models.CreateSubscriptionRequest) (models.Subscription, error) {
	if req.ServiceName == "" {
		return models.Subscription{}, ErrInvalidInput
	}
	if req.Price < 0 {
		return models.Subscription{}, ErrInvalidInput
	}
	if req.UserID == "" {
		return models.Subscription{}, ErrInvalidInput
	}
	if req.StartDate == "" {
		return models.Subscription{}, ErrInvalidInput
	}

	sub := models.Subscription{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	return s.repo.Create(ctx, sub)
}

// Получаем подписку по ID: проверяем ID и запрашиваем из репозитория
func (s *SubscriptionService) GetByID(ctx context.Context, id string) (models.Subscription, error) {
	if id == "" {
		return models.Subscription{}, ErrInvalidInput
	}
	return s.repo.GetByID(ctx, id)
}

// Получаем список подписок с фильтрами через репозиторий
func (s *SubscriptionService) List(ctx context.Context, filters map[string]string) ([]models.Subscription, error) {
	return s.repo.List(ctx, filters)
}

// Обновляем подписку: проверяем ID, получаем текущую запись, применяем изменения и сохраняем
func (s *SubscriptionService) Update(ctx context.Context, id string, req models.UpdateSubscriptionRequest) (models.Subscription, error) {
	if id == "" {
		return models.Subscription{}, ErrInvalidInput
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return models.Subscription{}, ErrNotFound
	}

	if req.ServiceName != "" {
		existing.ServiceName = req.ServiceName
	}
	if req.Price > 0 {
		existing.Price = req.Price
	}
	if req.StartDate != "" {
		existing.StartDate = req.StartDate
	}
	if req.EndDate != nil {
		existing.EndDate = req.EndDate
	}

	return s.repo.Update(ctx, existing)
}

// Удаляем подписку по ID: проверяем ID и делегируем репозиторию
func (s *SubscriptionService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}
	return s.repo.Delete(ctx, id)
}

// Рассчитываем общую сумму подписок с учётом фильтров через репозиторий
func (s *SubscriptionService) Sum(ctx context.Context, filters map[string]string) (int, error) {
	return s.repo.Sum(ctx, filters)
}
