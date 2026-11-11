package litesql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cattlecloud.net/go/scope"
	"github.com/shoenig/lang"
)

var (
	// ReadConsistency provides options for general use transactional reads.
	//
	// Note that the sqlite3 driver does not implement transaction isolation
	// levels; the behavior of a read transaction is determined by the
	// transaction lock mode of the connection, which is deferred, providing
	// snapshot isolation when the database is in WAL mode.
	//
	// This value has no effect.
	ReadConsistency = &sql.TxOptions{
		ReadOnly:  true,
		Isolation: sql.LevelSnapshot,
	}

	// WriteConsistency provides options for general use transactional writes.
	//
	// Note that the sqlite3 driver does not implement transaction isolation
	// levels; the behavior of a write transaction is determined by the
	// transaction lock mode of the connection, which is immediate, taking
	// the write lock when the transaction begins.
	//
	// This value has no effect.
	WriteConsistency = &sql.TxOptions{
		ReadOnly:  false,
		Isolation: sql.LevelSerializable,
	}
)

// LiteDB is a library around an sqlite3 database providing reasonable default
// values and an easy-to-use set of APIs for efficient and performant access.
//
// Write transactions are performed using a dedicated connection, so that only
// one write transaction is active at a time. Read transactions are performed
// using a pool of connections.
type LiteDB struct {
	reads *sql.DB
	write *sql.DB
}

// Open the database of the given filename, using config to tune the PRAGMA
// values and connection string parameters.
func Open(filename string, config *Configuration) (*LiteDB, error) {
	ctx, cancel := scope.TTL(10 * time.Second)
	defer cancel()

	// the write connection is opened first, so that the database is created
	// and its journal mode is established before any read connections are
	// opened
	write := openDB(uri(filename, config, true), config, 1)
	if err := write.PingContext(ctx); err != nil {
		_ = write.Close()
		return nil, fmt.Errorf("litesql: unable to open database: %w", err)
	}

	reads := openDB(uri(filename, config, false), config, config.MaxConnectionsOpen)
	if err := reads.PingContext(ctx); err != nil {
		_ = reads.Close()
		_ = write.Close()
		return nil, fmt.Errorf("litesql: unable to open database: %w", err)
	}

	return &LiteDB{reads: reads, write: write}, nil
}

// uri compiles the connection string for the given filename and configuration.
//
// Read connections use deferred transactions and are restricted to reading.
// The write connection uses immediate transactions, taking the write lock
// when the transaction begins.
func uri(filename string, config *Configuration, writer bool) string {
	lock := lang.Maybe(writer, config.TransactionLock, "deferred")

	parameters := []string{
		fmt.Sprintf("mode=%s", config.Mode),                 // nolint:perfsprint
		fmt.Sprintf("encoding=%s", config.Encoding),         // nolint:perfsprint
		fmt.Sprintf("_txlock=%s", lock),                     // nolint:perfsprint
		fmt.Sprintf("_journal_mode=%s", config.JournalMode), // nolint:perfsprint
		fmt.Sprintf("_auto_vacuum=%s", config.AutoVacuum),   // nolint:perfsprint
		fmt.Sprintf("_synchronous=%s", config.Synchronous),  // nolint:perfsprint
		fmt.Sprintf("_busy_timeout=%d", config.BusyTimeout),
		fmt.Sprintf("_foreign_keys=%t", config.ForeignKeys),
		fmt.Sprintf("_cache_size=%d", config.CacheSize),
		fmt.Sprintf("_query_only=%t", !writer),
	}

	return fmt.Sprintf("file:%s?%s", filename, strings.Join(parameters, "&"))
}

// Close the underlying database.
func (ldb *LiteDB) Close() error {
	return errors.Join(ldb.reads.Close(), ldb.write.Close())
}

// Pragmas will query the database for the common set of PRAGMA values, so
// that one may look at them in all their magnificent glory.
//
// The values are those of a read connection.
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
		row := ldb.reads.QueryRowContext(ctx, statement)
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
	tx, terr := ldb.reads.BeginTx(ctx, ReadConsistency)
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
// Write transactions use a dedicated connection, so that only one write
// transaction is active at a time; a second write transaction waits for the
// first to complete. A write transaction that is neither committed nor rolled
// back blocks all subsequent write transactions.
//
// The CloseFunc must be called when done with the transaction.
func (ldb *LiteDB) StartWrite(ctx scope.C) (*sql.Tx, CloseFunc, error) {
	tx, terr := ldb.write.BeginTx(ctx, WriteConsistency)
	if terr != nil {
		return nil, nil, fmt.Errorf("litesql: unable to create write transaction: %w", terr)
	}
	return tx, func() { _ = tx.Rollback() }, nil
}
