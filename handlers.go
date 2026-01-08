package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// createNoteHandler создает новую заметку
func createNoteHandler(service *NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var note Note

		// Получаем данные из запроса
		if err := c.ShouldBindJSON(&note); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
			return
		}

		// Используем сервис для создания заметки
		createdNote, err := service.CreateNote(note)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось создать заметку",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Заметка успешно создана",
			"note":    createdNote,
		})
	}
}

// getNotesHandler возвращает все заметки
func getNotesHandler(service *NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		notes, err := service.GetAllNotes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось получить заметки",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count": len(notes),
			"notes": notes,
		})
	}
}

// getNoteHandler возвращает заметку по ID
func getNoteHandler(service *NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Неверный формат ID",
			})
			return
		}

		note, err := service.GetNote(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Заметка не найдена",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"note": note,
		})
	}
}

// deleteNotesHandler удаляет все заметки
func deleteNotesHandler(service *NoteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		deletedCount, err := service.DeleteAllNotes()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Не удалось удалить заметки",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Все заметки успешно удалены",
			"deleted_count": deletedCount,
		})
	}
}
