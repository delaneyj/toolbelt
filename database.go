package toolbelt

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
	"zombiezen.com/go/sqlite"
	"zombiezen.com/go/sqlite/sqlitemigration"
	"zombiezen.com/go/sqlite/sqlitex"
)

type Database struct {
	write *sqlitex.Pool
	read  *sqlitex.Pool
}

type TxFn func(tx *sqlite.Conn) error

func NewDatabase(ctx context.Context, dbFilename string, migrations []string) (*Database, error) {
	if err := os.MkdirAll(filepath.Dir(dbFilename), 0755); err != nil {
		return nil, fmt.Errorf("could not create database directory: %w", err)
	}

	uri := fmt.Sprintf("file:%s?_journal_mode=WAL&_synchronous=NORMAL", dbFilename)
	writePool, err := sqlitex.Open(uri, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("could not open write pool: %w", err)
	}

	readPool, err := sqlitex.Open(uri, 0, runtime.NumCPU())
	if err != nil {
		return nil, fmt.Errorf("could not open read pool: %w", err)
	}

	db := &Database{
		write: writePool,
		read:  readPool,
	}

	if err := db.WriteTX(ctx, func(tx *sqlite.Conn) error {
		foreignKeysStmt := tx.Prep("PRAGMA foreign_keys = ON")
		defer foreignKeysStmt.Finalize()
		if _, err := foreignKeysStmt.Step(); err != nil {
			return fmt.Errorf("failed to enable foreign keys: %w", err)
		}

		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	schema := sqlitemigration.Schema{Migrations: migrations}
	conn := db.write.Get(ctx)
	if err := sqlitemigration.Migrate(ctx, conn, schema); err != nil {
		db.write.Put(conn)
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}
	db.write.Put(conn)

	return db, nil
}

func (db *Database) Close() error {
	if err := db.write.Close(); err != nil {
		return fmt.Errorf("failed to close write pool: %w", err)
	}

	if err := db.read.Close(); err != nil {
		return fmt.Errorf("failed to close read pool: %w", err)
	}

	return nil
}

func (db *Database) WriteTX(ctx context.Context, fn TxFn) (err error) {
	conn := db.write.Get(ctx)
	if conn == nil {
		return fmt.Errorf("could not get write connection from pool")
	}
	defer db.write.Put(conn)

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
	conn := db.read.Get(ctx)
	if conn == nil {
		return fmt.Errorf("could not get read connection from pool")
	}
	defer db.read.Put(conn)

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

func JulianDayToTimestampStmt(stmt *sqlite.Stmt, colName string) *timestamppb.Timestamp {
	julianDays := stmt.GetFloat(colName)
	return JulianDayToTimestamp(julianDays)
}

func JulianDayToTimeStmt(stmt *sqlite.Stmt, colName string) time.Time {
	julianDays := stmt.GetFloat(colName)
	return JulianDayToTime(julianDays)
}

func Bytes(stmt *sqlite.Stmt, colName string) []byte {
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
