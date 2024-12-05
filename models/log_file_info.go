package models

import "time"

type LogFileInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Type         string    `json:"type"`
	RegisteredAt time.Time `json:"registered_at"`
}
