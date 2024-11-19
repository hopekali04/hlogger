package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

// LogFileRegistry stores registered log file information
type LogFileRegistry struct {
	sync.RWMutex
	Files map[string]LogFileInfo `json:"files"`
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

type FiberLogEntry struct {
	Error interface{} `json:"error"`
	Level string      `json:"level"`
	Msg   string      `json:"msg"`
	Time  string      `json:"time"`
}
type LaravelLogParser struct{}

type FiberLogParser struct{}

type LogParser interface {
	Parse(line string) (*LogEntry, error)
}
type LogFileInfo struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Type         string    `json:"type"` // "laravel" or "fiber"
	RegisteredAt time.Time `json:"registered_at"`
}

type RegisterLogRequest struct {
	Name string `json:""`
	Path string `json:""`
	Type string `json:""`
}

var registry *LogFileRegistry

func init() {
	registry = &LogFileRegistry{
		Files: make(map[string]LogFileInfo),
	}
	loadRegistry()
}

func main() {
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	app.Use(logger.New())
	app.Static("/", "./public")

	// Existing routes
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Log Viewer",
		})
	})

	app.Get("/api/logs", getLogs)

	// New routes for log file registration
	app.Post("/api/logs/register", registerLogFile)
	app.Get("/api/logs/files", getRegisteredFiles)
	app.Delete("/api/logs/files/:id", deleteLogFile)

	log.Fatal(app.Listen(":3000"))
}

func registerLogFile(c *fiber.Ctx) error {
	var req RegisterLogRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"msg":   "parsing error",
		})
	}

	// Validate request
	if err := validateRegisterRequest(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Check if file exists
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File does not exist",
		})
	}

	// Generate unique ID and sanitize name
	id := generateUniqueID()
	sanitizedName := sanitizeFileName(req.Name)

	// Check for name uniqueness
	if !isNameUnique(sanitizedName) {
		sanitizedName = makeNameUnique(sanitizedName)
	}

	// Create new log file info
	fileInfo := LogFileInfo{
		ID:           id,
		Name:         sanitizedName,
		Path:         filepath.Clean(req.Path),
		Type:         req.Type,
		RegisteredAt: time.Now(),
	}

	// Add to registry
	registry.Lock()
	registry.Files[id] = fileInfo
	registry.Unlock()

	// Save registry to disk
	if err := saveRegistry(); err != nil {
		log.Printf("Error saving registry: %v", err)
	}

	return c.JSON(fiber.Map{
		"message": "Log file registered successfully",
		"file":    fileInfo,
	})
}

func getRegisteredFiles(c *fiber.Ctx) error {
	registry.RLock()
	defer registry.RUnlock()

	files := make([]LogFileInfo, 0, len(registry.Files))
	for _, file := range registry.Files {
		files = append(files, file)
	}

	return c.JSON(fiber.Map{
		"files": files,
	})
}

func deleteLogFile(c *fiber.Ctx) error {
	id := c.Params("id")

	registry.Lock()
	defer registry.Unlock()

	if _, exists := registry.Files[id]; !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Log file not found",
		})
	}

	delete(registry.Files, id)
	if err := saveRegistry(); err != nil {
		log.Printf("Error saving registry: %v", err)
	}

	return c.JSON(fiber.Map{
		"message": "Log file removed from registry",
	})
}

// Updated getLogs function to work with registry
func getLogs(c *fiber.Ctx) error {
	fileID := c.Query("id")

	registry.RLock()
	fileInfo, exists := registry.Files[fileID]
	registry.RUnlock()

	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Log file not found",
		})
	}

	var parser LogParser
	switch fileInfo.Type {
	case "laravel":
		parser = &LaravelLogParser{}
	case "fiber":
		parser = &FiberLogParser{}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid log type",
		})
	}

	logs, err := readLogs(fileInfo.Path, parser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"data": logs,
	})
}

func readLogs(filePath string, parser LogParser) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logs []LogEntry
	scanner := bufio.NewScanner(file)

	// Increase scanner buffer size for large lines
	const maxCapacity = 512 * 1024 // 512KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		entry, err := parser.Parse(line)
		if err == nil {
			logs = append(logs, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// Helper functions
func generateUniqueID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func sanitizeFileName(name string) string {
	// Remove invalid characters and trim spaces
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '-'
	}, name)
	return strings.Trim(name, "-")
}

func isNameUnique(name string) bool {
	registry.RLock()
	defer registry.RUnlock()

	for _, file := range registry.Files {
		if strings.EqualFold(file.Name, name) {
			return false
		}
	}
	return true
}

func makeNameUnique(name string) string {
	counter := 1
	newName := name
	for !isNameUnique(newName) {
		newName = fmt.Sprintf("%s-%d", name, counter)
		counter++
	}
	return newName
}

func validateRegisterRequest(req RegisterLogRequest) error {
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

// Registry persistence
func getRegistryPath() string {
	return filepath.Join(".", "log_registry.json")
}

func saveRegistry() error {
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getRegistryPath(), data, 0644)
}

func loadRegistry() {
	data, err := os.ReadFile(getRegistryPath())
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Error reading registry file: %v", err)
		}
		return
	}

	if err := json.Unmarshal(data, registry); err != nil {
		log.Printf("Error parsing registry file: %v", err)
	}
}

func (p *LaravelLogParser) Parse(line string) (*LogEntry, error) {
	if !strings.HasPrefix(line, "[") {
		return nil, fmt.Errorf("not a log line")
	}

	re := regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})\] (\w+)\.(\w+): (.+)$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) != 5 {
		return nil, fmt.Errorf("invalid log format")
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		return nil, err
	}

	return &LogEntry{
		Timestamp: timestamp.Format(time.RFC3339),
		Level:     strings.ToUpper(matches[3]),
		Message:   matches[4],
	}, nil
}

func (p *FiberLogParser) Parse(line string) (*LogEntry, error) {
	var fiberLog FiberLogEntry
	if err := json.Unmarshal([]byte(line), &fiberLog); err != nil {
		return nil, err
	}

	// Parse the time string
	timestamp, err := time.Parse(time.RFC3339, fiberLog.Time)
	if err != nil {
		return nil, err
	}

	// Construct the message, including error if present
	message := fiberLog.Msg
	if fiberLog.Error != nil {
		switch e := fiberLog.Error.(type) {
		case string:
			if e != "" {
				message = fmt.Sprintf("%s (Error: %s)", message, e)
			}
		case map[string]interface{}:
			errorJSON, _ := json.Marshal(e)
			message = fmt.Sprintf("%s (Error: %s)", message, string(errorJSON))
		}
	}

	return &LogEntry{
		Timestamp: timestamp.Format(time.RFC3339),
		Level:     strings.ToUpper(fiberLog.Level),
		Message:   message,
	}, nil
}
