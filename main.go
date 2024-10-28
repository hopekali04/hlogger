package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/template/html/v2"
)

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

type LogParser interface {
	Parse(line string) (*LogEntry, error)
}

type LaravelLogParser struct{}

type FiberLogParser struct{}

func main() {
	app := fiber.New(fiber.Config{
		Views: html.New("./views", ".html"),
	})

	app.Use(logger.New())
	app.Static("/", "./public")
	app.Get("/api/logs", getLogs)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Render("index", fiber.Map{
			"Title": "Log Viewer",
		})
	})

	log.Fatal(app.Listen(":3000"))
}

func getLogs(c *fiber.Ctx) error {
	logType := c.Query("type")
	var parser LogParser
	var logPath string

	switch logType {
	case "laravel":
		parser = &LaravelLogParser{}
		logPath = "laravel.log"
	case "fiber":
		parser = &FiberLogParser{}
		logPath = "canecc.log" 
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid log type"})
	}

	logs, err := readLogs(logPath, parser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"data": logs})
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
