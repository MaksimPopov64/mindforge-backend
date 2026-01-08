package main

import (
	"log"
	"time"
)

// NoteService - сервис для работы с заметками
type NoteService struct {
	repo *NoteRepository
}

// NewNoteService создает новый сервис
func NewNoteService(repo *NoteRepository) *NoteService {
	return &NoteService{repo: repo}
}

// CreateNote обрабатывает и создает заметку
func (s *NoteService) CreateNote(input Note) (*Note, error) {
	// Обрабатываем контент
	processedContent, err := ProcessNoteContent(input.Text)
	if err != nil {
		log.Printf("⚠️ Ошибка обработки контента: %v", err)
		return nil, err
	}

	// Создаем новую заметку
	note := &Note{
		Title:       processedContent.Title,
		Text:        processedContent.PlainText,
		Content:     processedContent.HTMLContent,
		OriginalURL: processedContent.OriginalURL,
		Tags:        processedContent.GeneratedTags,
		CreatedAt:   time.Now(),
	}

	// Сохраняем в БД
	err = s.repo.Create(note)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Заметка сохранена (ID: %d). Теги: %s", note.ID, note.Tags)
	return note, nil
}

// GetAllNotes возвращает все заметки
func (s *NoteService) GetAllNotes() ([]Note, error) {
	return s.repo.GetAll()
}

// GetNote возвращает заметку по ID
func (s *NoteService) GetNote(id int) (*Note, error) {
	return s.repo.GetByID(id)
}

// DeleteAllNotes удаляет все заметки
func (s *NoteService) DeleteAllNotes() (int64, error) {
	return s.repo.DeleteAll()
}
