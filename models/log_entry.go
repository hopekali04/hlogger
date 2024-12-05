package models

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
