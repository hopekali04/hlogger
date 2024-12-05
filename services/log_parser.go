package services

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hopekali04/hlogger/models"
)

type LogParser interface {
	Parse(line string) (*models.LogEntry, error)
}

type LaravelLogParser struct{}
type FiberLogParser struct{}

func (p *LaravelLogParser) Parse(line string) (*models.LogEntry, error) {
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

	return &models.LogEntry{
		Timestamp: timestamp.Format(time.RFC3339),
		Level:     strings.ToUpper(matches[3]),
		Message:   matches[4],
	}, nil
}

func (p *FiberLogParser) Parse(line string) (*models.LogEntry, error) {
	var fiberLog models.FiberLogEntry
	if err := json.Unmarshal([]byte(line), &fiberLog); err != nil {
		return nil, err
	}

	timestamp, err := time.Parse(time.RFC3339, fiberLog.Time)
	if err != nil {
		return nil, err
	}

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

	return &models.LogEntry{
		Timestamp: timestamp.Format(time.RFC3339),
		Level:     strings.ToUpper(fiberLog.Level),
		Message:   message,
	}, nil
}

func ReadLogs(fileInfo models.LogFileInfo) ([]models.LogEntry, error) {
	file, err := os.Open(fileInfo.Path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var parser LogParser
	switch fileInfo.Type {
	case "laravel":
		parser = &LaravelLogParser{}
	case "fiber":
		parser = &FiberLogParser{}
	default:
		return nil, fmt.Errorf("invalid log type")
	}

	var logs []models.LogEntry
	scanner := bufio.NewScanner(file)

	const maxCapacity = 512 * 1024 // 512KB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		line := scanner.Text()
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
