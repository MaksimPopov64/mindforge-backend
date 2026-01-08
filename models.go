package main

import (
	"time"
)

type Note struct {
	ID          int       `json:"id"`
	Title       string    `json:"title,omitempty"`
	Text        string    `json:"text"`
	Content     string    `json:"content,omitempty"`
	OriginalURL string    `json:"original_url,omitempty"`
	Tags        string    `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
}
