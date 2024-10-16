package main

import (
    "bufio"
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

type LogParser interface {
    Parse(line string) (*LogEntry, error)
}

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

    switch logType {
    case "laravel":
        parser = &LaravelLogParser{}
    default:
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid log type"})
    }

    // TODO: GET LOG FROM ENV FILE
    logs, err := readLogs("laravel.log", parser)
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
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "[") {
            entry, err := parser.Parse(line)
            if err == nil {
                logs = append(logs, *entry)
            }
        }
        // Skip lines that don't start with "[" (stack traces)
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return logs, nil
}

type LaravelLogParser struct{}

func (p *LaravelLogParser) Parse(line string) (*LogEntry, error) {
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