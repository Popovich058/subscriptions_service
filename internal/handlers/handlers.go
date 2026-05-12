package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"subscriptions_service/internal/config"
	"subscriptions_service/internal/models"
	"subscriptions_service/internal/service"
)

// SubscriptionHandler — обработчик HTTP-запросов для подписок
type SubscriptionHandler struct {
	svc *service.SubscriptionService 
	cfg *config.Config             
}

// Создаём новый экземпляр SubscriptionHandler
func NewSubscriptionHandler(svc *service.SubscriptionService, cfg *config.Config) *SubscriptionHandler {
	return &SubscriptionHandler{svc: svc, cfg: cfg}
}

// Обрабатываем создание новой подписки
func (h *SubscriptionHandler) Create(c *gin.Context) {
	var req models.CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Ошибка валидации JSON
		return
	}

	sub, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Ошибка создания
		return
	}

	c.JSON(http.StatusCreated, sub) // Успешное создание (201)
}

// Обрабатываем получение подписки по ID
func (h *SubscriptionHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"}) // ID не указан
		return
	}

	sub, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"}) // Подписка не найдена
		return
	}

	c.JSON(http.StatusOK, sub) // Успешный ответ (200)
}

// Обрабатываем получение списка подписок с фильтрацией
func (h *SubscriptionHandler) List(c *gin.Context) {
	filters := map[string]string{}

	// Добавляем фильтры из query‑параметров
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if serviceName := c.Query("service_name"); serviceName != "" {
		filters["service_name"] = serviceName
	}

	subs, err := h.svc.List(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Ошибка получения списка
		return
	}

	if subs == nil {
		subs = []models.Subscription{} // Пустой массив вместо nil
	}

	c.JSON(http.StatusOK, subs) // Возвращаем список подписок (200)
}

// Обрабатываем обновление подписки по ID
func (h *SubscriptionHandler) Update(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"}) // ID не указан
		return
	}

	var req models.UpdateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}) // Ошибка валидации JSON
		return
	}

	sub, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Ошибка обновления
		return
	}

	c.JSON(http.StatusOK, sub) // Успешное обновление (200)
}

// Обрабатываем удаление подписки по ID
func (h *SubscriptionHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"}) // ID не указан
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Ошибка удаления
		return
	}

	c.Status(http.StatusNoContent) // Успешное удаление (204 — без тела ответа)
}

// Обрабатываем расчёт суммы подписок с фильтрацией по параметрам
func (h *SubscriptionHandler) Sum(c *gin.Context) {
	filters := map[string]string{}

	// Собираем фильтры из query-параметров (user_id, service_name, даты)
	if userID := c.Query("user_id"); userID != "" {
		filters["user_id"] = userID
	}
	if serviceName := c.Query("service_name"); serviceName != "" {
		filters["service_name"] = serviceName
	}
	if startDate := c.Query("start_date"); startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate := c.Query("end_date"); endDate != "" {
		filters["end_date"] = endDate
	}

	total, err := h.svc.Sum(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Ошибка расчёта суммы
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": total}) // Возвращаем итоговую сумму (200)
}
