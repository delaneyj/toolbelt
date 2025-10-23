package examples

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/delaneyj/toolbelt"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func SetupDB(ctx context.Context, dataFolder string, shouldClear bool) (*toolbelt.Database, error) {
	migrationsDir := "migrations"
	migrationsFiles, err := migrationsFS.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}
	slices.SortFunc(migrationsFiles, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	migrations := make([]string, len(migrationsFiles))
	for i, file := range migrationsFiles {
		fn := filepath.Join(migrationsDir, file.Name())
		fnts := filepath.ToSlash(fn)
		f, err := migrationsFS.Open(fnts)
		if err != nil {
			return nil, fmt.Errorf("failed to open migration file: %w", err)
		}
		defer f.Close()

		content, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file: %w", err)
		}

		migrations[i] = string(content)
	}

	dbFolder := filepath.Join(dataFolder, "database")
	if shouldClear {
		log.Printf("Clearing database folder: %s", dbFolder)
		if err := os.RemoveAll(dbFolder); err != nil {
			return nil, fmt.Errorf("failed to remove database folder: %w", err)
		}
	}
	dbFilename := filepath.Join(dbFolder, "examples.sqlite")
	db, err := toolbelt.NewDatabase(ctx, dbFilename, migrations)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	return db, nil
}
