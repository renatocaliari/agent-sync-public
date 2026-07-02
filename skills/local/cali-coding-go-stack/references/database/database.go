package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
}

type databaseOptions struct {
	filename    string
	migrations  []string
	pragmas     []string
	shouldClear bool
}

type DatabaseOption func(*databaseOptions)

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

func NewDatabase(ctx context.Context, opts ...DatabaseOption) (*Database, error) {
	options := databaseOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	if len(options.pragmas) == 0 {
		options.pragmas = []string{"foreign_keys = ON"}
	}

	if options.filename == "" {
		options.filename = "database.sqlite"
	}

	db := &Database{
		filename:   options.filename,
		migrations: options.migrations,
		pragmas:    options.pragmas,
	}

	if err := db.Reset(ctx, options.shouldClear); err != nil {
		return nil, fmt.Errorf("failed to reset database: %w", err)
	}

	return db, nil
}

func (db *Database) Path() string {
	return db.filename
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

	return nil
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
