package toolbelt

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Database struct {
	filename   string
	migrations []string
	writePool  *sqlitex.Pool
	readPool   *sqlitex.Pool
	pragmas    []string
	flushWAL   bool
}

type databaseOptions struct {
	filename    string
	migrations  []string
	pragmas     []string
	shouldClear bool
	flushWAL    bool
}

type DatabaseOption func(*databaseOptions)

const defaultDatabaseFilename = "database.sqlite"

func DatabaseWithFilename(filename string) DatabaseOption {
	return func(o *databaseOptions) {
		o.filename = filename
	}
}

func DatabaseWithMigrations(migrations []string) DatabaseOption {
	cp := append([]string(nil), migrations...)
	return func(o *databaseOptions) {
		o.migrations = cp
	}
}

func DatabaseWithPragmas(pragmas ...string) DatabaseOption {
	cp := append([]string(nil), pragmas...)
	return func(o *databaseOptions) {
		o.pragmas = cp
	}
}

func DatabaseWithShouldClear(shouldClear bool) DatabaseOption {
	return func(o *databaseOptions) {
		o.shouldClear = shouldClear
	}
}

func DatabaseWithFlushWAL() DatabaseOption {
	return func(o *databaseOptions) {
		o.flushWAL = true
	}
}

func DatabaseWithoutFlushWAL() DatabaseOption {
	return func(o *databaseOptions) {
		o.flushWAL = false
	}
}

func normalizePragma(pragma string) string {
	s := strings.TrimSpace(pragma)
	if !strings.HasPrefix(strings.ToUpper(s), "PRAGMA ") {
		s = "PRAGMA " + s
	}
	if !strings.HasSuffix(s, ";") {
		s += ";"
	}
	return s
}

type TxFn func(tx *sqlite.Conn) error

func flushWAL(conn *sqlite.Conn) error {
	if !conn.AutocommitEnabled() {
		if err := sqlitex.ExecuteTransient(conn, "ROLLBACK;", nil); err != nil {
			return fmt.Errorf("reset transaction before flush: %w", err)
		}
	}
	if err := sqlitex.ExecuteTransient(conn, "VACUUM;", nil); err != nil {
		return fmt.Errorf("vacuum database: %w", err)
	}
	if err := sqlitex.ExecuteTransient(conn, "PRAGMA wal_checkpoint(TRUNCATE);", nil); err != nil {
		return fmt.Errorf("execute WAL checkpoint: %w", err)
	}
	return nil
}

func NewDatabase(ctx context.Context, opts ...DatabaseOption) (*Database, error) {
	options := databaseOptions{
		flushWAL: true,
	}
	for _, opt := range opts {
		opt(&options)
	}

	if len(options.pragmas) == 0 {
		options.pragmas = []string{"foreign_keys = ON"}
	}

	if options.filename == "" {
		options.filename = defaultDatabaseFilename
	}

	db := &Database{
		filename:   options.filename,
		migrations: options.migrations,
		pragmas:    options.pragmas,
		flushWAL:   options.flushWAL,
	}

	if err := db.Reset(ctx, options.shouldClear); err != nil {
		return nil, fmt.Errorf("failed to reset database: %w", err)
	}

	return db, nil
}

func (db *Database) Path() string {
	return db.filename
}

func (db *Database) WriteWithoutTx(ctx context.Context, fn TxFn) error {
	conn, err := db.writePool.Take(ctx)
	if err != nil {
		return fmt.Errorf("failed to take write connection: %w", err)
	}
	if conn == nil {
		return fmt.Errorf("could not get write connection from pool")
	}
	defer db.writePool.Put(conn)

	if err := fn(conn); err != nil {
		return fmt.Errorf("could not execute write transaction: %w", err)
	}

	return nil
}

func (db *Database) Reset(ctx context.Context, shouldClear bool) (err error) {
	if err := db.Close(); err != nil {
		return fmt.Errorf("could not close database: %w", err)
	}

	if shouldClear {
		dbFiles, err := filepath.Glob(db.filename + "*")
		if err != nil {
			return fmt.Errorf("could not glob database files: %w", err)
		}
		for _, file := range dbFiles {
			if err := os.Remove(file); err != nil {
				return fmt.Errorf("could not remove database file: %w", err)
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(db.filename), 0o755); err != nil {
		return fmt.Errorf("could not create database directory: %w", err)
	}

	uri := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL", db.filename)

	db.writePool, err = sqlitex.NewPool(uri, sqlitex.PoolOptions{
		PoolSize: 1,
		PrepareConn: func(conn *sqlite.Conn) error {
			for _, pragma := range db.pragmas {
				stmt := normalizePragma(pragma)
				if err := sqlitex.ExecuteTransient(conn, stmt, nil); err != nil {
					return fmt.Errorf("apply pragma %q: %w", pragma, err)
				}
			}
			return nil
		},
	})
	if err != nil {
		return fmt.Errorf("could not open write pool: %w", err)
	}

	db.readPool, err = sqlitex.NewPool(uri, sqlitex.PoolOptions{
		PoolSize: runtime.NumCPU(),
	})

	schema := sqlitemigration.Schema{Migrations: db.migrations}
	conn, err := db.writePool.Take(ctx)
	if err != nil {
		return fmt.Errorf("failed to take write connection: %w", err)
	}
	defer func() {
		if conn != nil {
			db.writePool.Put(conn)
		}
	}()

	if err := sqlitemigration.Migrate(ctx, conn, schema); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	db.writePool.Put(conn)
	conn = nil

	if db.flushWAL {
		if err := db.WriteWithoutTx(ctx, func(conn *sqlite.Conn) error {
			return flushWAL(conn)
		}); err != nil {
			return fmt.Errorf("failed to flush WAL: %w", err)
		}
	}

	return nil
}

func (db *Database) Close() error {
	errs := []error{}
	if db.writePool != nil {
		errs = append(errs, db.writePool.Close())
	}

	if db.readPool != nil {
		errs = append(errs, db.readPool.Close())
	}

	return errors.Join(errs...)
}

func (db *Database) FlushWAL(ctx context.Context) error {
	if db.writePool == nil {
		return fmt.Errorf("write pool is not initialized")
	}

	return db.WriteWithoutTx(ctx, func(conn *sqlite.Conn) error {
		return flushWAL(conn)
	})
}

func (db *Database) WriteTX(ctx context.Context, fn TxFn) (err error) {
	conn, err := db.writePool.Take(ctx)
	if err != nil {
		return fmt.Errorf("failed to take write connection: %w", err)
	}
	if conn == nil {
		return fmt.Errorf("could not get write connection from pool")
	}
	defer db.writePool.Put(conn)

	endFn, err := sqlitex.ImmediateTransaction(conn)
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer endFn(&err)

	if err := fn(conn); err != nil {
		return fmt.Errorf("could not execute write transaction: %w", err)
	}

	return nil
}

func (db *Database) ReadTX(ctx context.Context, fn TxFn) (err error) {
	conn, err := db.readPool.Take(ctx)
	if err != nil {
		return fmt.Errorf("failed to take read connection: %w", err)
	}
	if conn == nil {
		return fmt.Errorf("could not get read connection from pool")
	}
	defer db.readPool.Put(conn)

	endFn := sqlitex.Transaction(conn)
	defer endFn(&err)

	if err := fn(conn); err != nil {
		return fmt.Errorf("could not execute read transaction: %w", err)
	}

	return nil
}

const (
	secondsInADay      = 86400
	UnixEpochJulianDay = 2440587.5
)

var (
	JulianZeroTime = JulianDayToTime(0)
)

// TimeToJulianDay converts a time.Time into a Julian day.
func TimeToJulianDay(t time.Time) float64 {
	return float64(t.UTC().Unix())/secondsInADay + UnixEpochJulianDay
}

// JulianDayToTime converts a Julian day into a time.Time.
func JulianDayToTime(d float64) time.Time {
	return time.Unix(int64((d-UnixEpochJulianDay)*secondsInADay), 0).UTC()
}

func JulianNow() float64 {
	return TimeToJulianDay(time.Now())
}

func TimestampJulian(ts *timestamppb.Timestamp) float64 {
	return TimeToJulianDay(ts.AsTime())
}

func JulianDayToTimestamp(f float64) *timestamppb.Timestamp {
	t := JulianDayToTime(f)
	return timestamppb.New(t)
}

func StmtJulianToTimestamp(stmt *sqlite.Stmt, colName string) *timestamppb.Timestamp {
	julianDays := stmt.GetFloat(colName)
	return JulianDayToTimestamp(julianDays)
}

func StmtJulianToTime(stmt *sqlite.Stmt, colName string) time.Time {
	julianDays := stmt.GetFloat(colName)
	return JulianDayToTime(julianDays)
}

func DurationToMilliseconds(d time.Duration) int64 {
	return int64(d / time.Millisecond)
}

func MillisecondsToDuration(ms int64) time.Duration {
	return time.Duration(ms) * time.Millisecond
}

func StmtBytes(stmt *sqlite.Stmt, colName string) []byte {
	bl := stmt.GetLen(colName)
	if bl == 0 {
		return nil
	}

	buf := make([]byte, bl)
	if writtent := stmt.GetBytes(colName, buf); writtent != bl {
		return nil
	}

	return buf
}

func StmtBytesByCol(stmt *sqlite.Stmt, col int) []byte {
	bl := stmt.ColumnLen(col)
	if bl == 0 {
		return nil
	}

	buf := make([]byte, bl)
	if writtent := stmt.ColumnBytes(col, buf); writtent != bl {
		return nil
	}

	return buf
}

func MigrationsFromFS(migrationsFS embed.FS, migrationsDir string) ([]string, error) {
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

	return migrations, nil
}
