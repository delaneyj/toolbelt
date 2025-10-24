package examples

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupDB(t *testing.T) {
	tempDataDir := t.TempDir()

	db, err := SetupDB(context.Background(), tempDataDir, true)

	if err != nil {
		t.Fatalf("SetupDB failed: %v", err)
	}

	if db != nil {
		defer db.Close()
	}

	expectedDBPath := filepath.Join(tempDataDir, "database", "examples.sqlite")
	if _, err := os.Stat(expectedDBPath); os.IsNotExist(err) {
		t.Errorf("database file was not created at the expected path: %s", expectedDBPath)
	}
}
