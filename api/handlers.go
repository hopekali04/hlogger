package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/hopekali04/hlogger/models"
	"github.com/hopekali04/hlogger/services"
	"github.com/hopekali04/hlogger/utils"

	"path/filepath"
	"time"
)

func RegisterLogFile(c *fiber.Ctx) error {
	var req models.RegisterLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"msg":   "parsing error",
		})
	}

	if err := utils.ValidateRegisterRequest(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if _, err := utils.FileExists(req.Path); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File does not exist",
		})
	}

	id := utils.GenerateUniqueID()
	sanitizedName := utils.SanitizeFileName(req.Name)

	if !services.IsNameUnique(sanitizedName) {
		sanitizedName = services.MakeNameUnique(sanitizedName)
	}

	fileInfo := models.LogFileInfo{
		ID:           id,
		Name:         sanitizedName,
		Path:         filepath.Clean(req.Path),
		Type:         req.Type,
		RegisteredAt: time.Now(),
	}

	services.AddLogFile(fileInfo)

	return c.JSON(fiber.Map{
		"message": "Log file registered successfully",
		"file":    fileInfo,
	})
}

func GetRegisteredFiles(c *fiber.Ctx) error {
	files := services.GetAllLogFiles()

	return c.JSON(fiber.Map{
		"files": files,
	})
}

func DeleteLogFile(c *fiber.Ctx) error {
	id := c.Params("id")

	if !services.LogFileExists(id) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Log file not found",
		})
	}

	services.RemoveLogFile(id)

	return c.JSON(fiber.Map{
		"message": "Log file removed from registry",
	})
}

func GetLogByID(c *fiber.Ctx) error {
	fileID := c.Params("id")
	fileInfo, err := services.GetLogFileByID(fileID)

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Log file not found",
		})
	}

	logs, err := services.ReadLogs(fileInfo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": logs,
	})
}
