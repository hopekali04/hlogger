package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/hopekali04/hlogger/models"
)

func GenerateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func SanitizeFileName(name string) string {
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	return strings.Trim(name, "-")
}

func ValidateRegisterRequest(req models.RegisterLogRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("path is required")
	}
	if req.Type != "laravel" && req.Type != "fiber" {
		return fmt.Errorf("type must be either 'laravel' or 'fiber'")
	}
	return nil
}
