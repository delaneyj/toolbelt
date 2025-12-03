package toolbelt

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/benbjohnson/litestream/file"
	"github.com/stretchr/testify/require"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitex"
)

func TestDatabaseLitestreamFileReplicaRestore(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tmp := filepath.Join(".", t.TempDir())
	// tmp := filepath.Join(".", "data") //t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	replicaPath := filepath.Join(tmp, "replica")

	log.Println("dbPath:", dbPath)
	log.Println("replicaPath:", replicaPath)

	require.NoError(t, os.MkdirAll(replicaPath, 0o755))

	migrations := []string{
		"CREATE TABLE IF NOT EXISTS foo (id INTEGER PRIMARY KEY, val TEXT);",
	}

	cfg := DatabaseLitestreamConfig{
		ReplicaClient:   file.NewReplicaClient(replicaPath),
		MonitorInterval: 0,
		SyncInterval:    0,
	}

	db, err := NewDatabase(
		ctx,
		DatabaseWithFilename(dbPath),
		DatabaseWithMigrations(migrations),
		WithDatabaseLitestreamConfig(cfg),
	)
	require.NoError(t, err)

	err = db.WriteTX(ctx, func(conn *sqlite.Conn) error {
		return sqlitex.ExecuteTransient(conn, "INSERT INTO foo (val) VALUES ('hello')", nil)
	})
	require.NoError(t, err)

	lsdb := db.Litestream()
	require.NotNil(t, lsdb)
	require.NotNil(t, lsdb.Replica)

	require.NoError(t, lsdb.Sync(ctx))
	require.NoError(t, lsdb.Replica.Sync(ctx))
	require.NoError(t, db.Close())

	dbFiles, err := filepath.Glob(dbPath + "*")
	require.NoError(t, err)
	for _, f := range dbFiles {
		require.NoError(t, os.Remove(f))
	}

	dbRestored, err := NewDatabase(
		ctx,
		DatabaseWithFilename(dbPath),
		DatabaseWithMigrations(migrations),
		WithDatabaseLitestreamConfig(cfg),
	)
	require.NoError(t, err)
	defer func() {
		_ = dbRestored.Close()
	}()

	var count int
	err = dbRestored.ReadTX(ctx, func(conn *sqlite.Conn) error {
		stmt, err := conn.Prepare("SELECT COUNT(*) FROM foo")
		if err != nil {
			return err
		}
		defer stmt.Finalize()

		hasRow, err := stmt.Step()
		if err != nil {
			return err
		}
		if !hasRow {
			return fmt.Errorf("no rows returned")
		}

		count = stmt.ColumnInt(0)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, count)
}
