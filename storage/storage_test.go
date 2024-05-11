package storage

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestNew(t *testing.T) {
	t.Run("Create storate", func(t *testing.T) {
		file, err := os.CreateTemp("", "storage.*.db")
		if err != nil {
			t.Errorf("Get temp file error = %v", err)
			return
		}
		defer os.Remove(file.Name())

		_, err = New(file.Name())
		if err != nil {
			t.Errorf("New() error = %v", err)
			return
		}
	})
}
