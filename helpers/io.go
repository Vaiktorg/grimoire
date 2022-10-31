package helpers

import (
	"os"
)

func OpenFile(filepath string) (*os.File, error) {
	if _, err := os.Stat(filepath); err == nil {
		return os.OpenFile(filepath, os.O_RDWR, os.ModePerm)
	} else {
		return os.Create(filepath)
	}
}
