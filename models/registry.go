package models

import "sync"

// LogFileRegistry stores registered log file information
type LogFileRegistry struct {
	sync.RWMutex
	Files map[string]LogFileInfo `json:"files"`
}

var Registry *LogFileRegistry

func init() {
	Registry = &LogFileRegistry{
		Files: make(map[string]LogFileInfo),
	}
}
