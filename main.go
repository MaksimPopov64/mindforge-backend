package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Инициализация БД
	db, err := InitDB()
	if err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	defer db.Close()

	// Создаем слои приложения
	noteRepo := NewNoteRepository(db)
	noteService := NewNoteService(noteRepo)

	// Настройка роутера
	router := gin.Default()
	router.Use(corsMiddleware())

	// Настройка маршрутов
	setupRoutes(router, noteService)

	// Запуск сервера
	log.Println("Сервер запущен на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}

// setupRoutes настраивает все маршруты
func setupRoutes(router *gin.Engine, service *NoteService) {
	router.POST("/api/notes", createNoteHandler(service))
	router.GET("/api/notes", getNotesHandler(service))
	router.GET("/api/notes/:id", getNoteHandler(service))
	router.DELETE("/api/notes", deleteNotesHandler(service))
}

// CORS middleware остается без изменений
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
