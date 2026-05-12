package middleware

import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
)

// Logger возвращает middleware-обработчик для логирования HTTP-запросов
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()           // Фиксируем время начала запроса
		path := c.Request.URL.Path  // Получаем путь запроса
		method := c.Request.Method  // Получаем HTTP-метод (GET, POST и т.д.)

		c.Next()  // Передаём управление следующему обработчику в цепочке

		latency := time.Since(start) // Рассчитываем время выполнения запроса
		status := c.Writer.Status()  // Получаем статус-код ответа

		// Логируем данные о запросе: время, метод, путь, статус, время выполнения
		log.Printf("[%s] %s %s %d %v",
			time.Now().Format("2006-01-02 15:04:05"), // Форматированное время
			method,   // HTTP-метод
			path,     // Путь запроса
			status,   // Статус-код ответа
			latency,  // Время выполнения
		)
	}
}
