package litesql

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"cattlecloud.net/go/scope"
)

var (
	// ReadConsistency provides options for general use transactional reads.
	//
	// Do not modify unless you are fully aware of the consequences.
	ReadConsistency = &sql.TxOptions{
		ReadOnly:  true,
		Isolation: sql.LevelReadCommitted,
	}

	// WriteConsistency provides options for general use transactional writes.
	//
	// Do not modify unless you are fully aware of the consequences.
	WriteConsistency = &sql.TxOptions{
		ReadOnly:  false,
		Isolation: sql.LevelWriteCommitted,
	}
)

// LiteDB is an interface over a sqlite3 database providing reasonable default
// values and an easy-to-use set of APIs for efficient and performant access.
type LiteDB struct {
	db *sql.DB
}

// Open the database of the given filename, using config to tune the PRAGMA
// values and connection string parameters.
func Open(filename string, config *Configuration) (*LiteDB, error) {
	ctx, cancel := scope.TTL(10 * time.Second)
	defer cancel()

	// compile the connection string using the given configuration
	parameters := strings.Join([]string{
		fmt.Sprintf("mode=%s", config.Mode),                 // nolint:perfsprint
		fmt.Sprintf("encoding=%s", config.Encoding),         // nolint:perfsprint
		fmt.Sprintf("_txlock=%s", config.TransactionLock),   // nolint:perfsprint
		fmt.Sprintf("_journal_mode=%s", config.JournalMode), // nolint:perfsprint
		fmt.Sprintf("_auto_vacuum=%s", config.AutoVacuum),   // nolint:perfsprint
		fmt.Sprintf("_synchronous=%s", config.Synchronous),  // nolint:perfsprint
		fmt.Sprintf("_busy_timeout=%d", config.BusyTimeout),
		fmt.Sprintf("_foreign_keys=%t", config.ForeignKeys),
		fmt.Sprintf("_cache_size=%d", config.CacheSize),
	}, "&")

	// open the database at filepath using our connection string
	uri := fmt.Sprintf("file:%s?%s", filename, parameters)
	db, oerr := sql.Open("sqlite3", uri)
	if oerr != nil {
		return nil, fmt.Errorf("litesql: unable to open database: %w", oerr)
	}

	// no connection string option for setting mmap_size; do it manually
	mmap := fmt.Sprintf("pragma mmap_size = %d;", config.MemoryMapSize)
	if _, err := db.ExecContext(ctx, mmap); err != nil {
		return nil, fmt.Errorf("litesql: unable to set mmap_size: %w", err)
	}

	// set connection behaviors; max conns is defacto read concurrency
	db.SetMaxOpenConns(config.MaxConnectionsOpen)
	db.SetConnMaxIdleTime(0) // disable
	db.SetConnMaxLifetime(0) // disable

	return &LiteDB{db: db}, nil
}

// Close the underlying database.
func (ldb *LiteDB) Close() error {
	return ldb.db.Close()
}

// Pragmas will query the database for the common set of PRAGMA values, so
// that one may look at them in all their magnificent glory.
func (ldb *LiteDB) Pragmas(ctx scope.C) (map[string]string, error) {
	result := make(map[string]string, 9)

	keys := []string{
		"encoding",
		"busy_timeout",
		"foreign_keys",
		"journal_mode",
		"cache_size",
		"auto_vacuum",
		"synchronous",
		"mmap_size",
		"page_size",
	}

	for _, key := range keys {
		statement := "pragma " + key
		row := ldb.db.QueryRowContext(ctx, statement)
		var value string
		if err := row.Scan(&value); err != nil {
			switch key {
			case "mmap_size":
				continue // skip; in-memory databases do not use
			default:
				return nil, fmt.Errorf("litesql: unable to query %q: %w", statement, err)
			}
		}
		result[key] = value
	}

	return result, nil
}

// CloseFunc should always be called before a transaction goes out of scope,
// even if the transaction has been committed (which turns into no-op).
type CloseFunc func()

// StartRead creates a transaction useful for reading from the database.
//
// No need to call sql.Tx.Commit; allow the rollback to close the transaction.
//
// The sql.Tx must not be used for writing.
//
// The CloseFunc must be called when done with the transaction.
func (ldb *LiteDB) StartRead(ctx scope.C) (*sql.Tx, CloseFunc, error) {
	tx, terr := ldb.db.BeginTx(ctx, ReadConsistency)
	if terr != nil {
		return nil, nil, fmt.Errorf("litesql: unable to create read transaction: %w", terr)
	}
	return tx, func() { _ = tx.Rollback() }, nil
}

// StartWrite creates a transaction useful for writing to the database.
//
// Must call sql.Tx.Commit to complete the transaction.
//
// The sql.Tx may be used for reading and/or writing. Note that using a write
// transaction just for reading will be slower than a read only transaction.
//
// The CloseFunc must be called when done with the transaction.
func (ldb *LiteDB) StartWrite(ctx scope.C) (*sql.Tx, CloseFunc, error) {
	tx, terr := ldb.db.BeginTx(ctx, WriteConsistency)
	if terr != nil {
		return nil, nil, fmt.Errorf("litesql: unable to create write transaction: %w", terr)
	}
	return tx, func() { _ = tx.Rollback() }, nil
}
