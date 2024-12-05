package services

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hopekali04/hlogger/models"
)

func AddLogFile(fileInfo models.LogFileInfo) {
	models.Registry.Lock()
	models.Registry.Files[fileInfo.ID] = fileInfo
	models.Registry.Unlock()
	saveRegistry()
}

func GetAllLogFiles() []models.LogFileInfo {
	// Read the registry file
	path := getRegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading registry file: %v\n", err)
		return nil
	}

	// Unmarshal the JSON data into the registry struct
	var registry models.LogFileRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		fmt.Printf("Error unmarshalling registry data: %v\n", err)
		return nil
	}

	// Convert the Files map to a slice of LogFileInfo
	files := make([]models.LogFileInfo, 0, len(registry.Files))
	for _, file := range registry.Files {
		files = append(files, file)
	}

	fmt.Println("The files are:", files)
	return files
}

func LogFileExists(id string) bool {
	models.Registry.RLock()
	defer models.Registry.RUnlock()

	_, exists := models.Registry.Files[id]
	return exists
}

func GetLogFileByID(id string) (models.LogFileInfo, error) {
	// Read the registry file
	path := getRegistryPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return models.LogFileInfo{}, fmt.Errorf("error reading registry file: %v", err)
	}

	// Unmarshal the JSON data into the registry struct
	var registry models.LogFileRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return models.LogFileInfo{}, fmt.Errorf("error unmarshalling registry data: %v", err)
	}

	// Look up the file by ID
	fileInfo, exists := registry.Files[id]
	if !exists {
		return models.LogFileInfo{}, fmt.Errorf("log file not found")
	}

	return fileInfo, nil
}

func RemoveLogFile(id string) {
	models.Registry.Lock()
	delete(models.Registry.Files, id)
	models.Registry.Unlock()
	saveRegistry()
}

func IsNameUnique(name string) bool {
	models.Registry.RLock()
	defer models.Registry.RUnlock()

	for _, file := range models.Registry.Files {
		if strings.EqualFold(file.Name, name) {
			return false
		}
	}
	return true
}

func MakeNameUnique(name string) string {
	counter := 1
	newName := name
	for !IsNameUnique(newName) {
		newName = fmt.Sprintf("%s-%d", name, counter)
		counter++
	}
	return newName
}

func saveRegistry() {
	data, err := json.MarshalIndent(models.Registry, "", "  ")
	if err != nil {
		log.Printf("Error saving registry: %v", err)
		return
	}
	err = os.WriteFile(getRegistryPath(), data, 0644)
	if err != nil {
		log.Printf("Error writing registry to disk: %v", err)
	}
}

func getRegistryPath() string {
	return filepath.Join(".", "log_registry.json")
}
