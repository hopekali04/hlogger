package utils

import (
	"fmt"
	"os"
)

func FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, fmt.Errorf("file does not exist")
	}
	return true, nil
}
