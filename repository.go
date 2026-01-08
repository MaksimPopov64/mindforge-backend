package main

import (
	"database/sql"
	"log"
)

// NoteRepository - репозиторий для работы с заметками
type NoteRepository struct {
	db *sql.DB
}

// NewNoteRepository создает новый репозиторий
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// Create создает новую заметку
func (r *NoteRepository) Create(note *Note) error {
	result, err := r.db.Exec(`
		INSERT INTO notes (title, text, content, original_url, tags, created_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		note.Title,
		note.Text,
		note.Content,
		note.OriginalURL,
		note.Tags,
		note.CreatedAt,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	note.ID = int(id)
	return nil
}

// GetAll возвращает все заметки
func (r *NoteRepository) GetAll() ([]Note, error) {
	rows, err := r.db.Query(`
		SELECT id, title, text, tags, created_at 
		FROM notes 
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []Note
	for rows.Next() {
		var note Note
		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Text,
			&note.Tags,
			&note.CreatedAt,
		)
		if err != nil {
			log.Printf("Ошибка чтения строки: %v", err)
			continue
		}
		notes = append(notes, note)
	}

	return notes, rows.Err()
}

// GetByID возвращает заметку по ID
func (r *NoteRepository) GetByID(id int) (*Note, error) {
	var note Note
	err := r.db.QueryRow(`
		SELECT id, title, text, content, original_url, tags, created_at 
		FROM notes 
		WHERE id = ?`, id).Scan(
		&note.ID,
		&note.Title,
		&note.Text,
		&note.Content,
		&note.OriginalURL,
		&note.Tags,
		&note.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &note, nil
}

// DeleteAll удаляет все заметки
func (r *NoteRepository) DeleteAll() (int64, error) {
	result, err := r.db.Exec("DELETE FROM notes")
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}
