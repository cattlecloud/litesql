package litesql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

// connector opens connections to the sqlite3 database, applying per-connection
// settings to each new connection.
type connector struct {
	// dsn is the connection string used to open the database.
	dsn string

	// init applies per-connection settings to a new connection.
	init func(context.Context, driver.Conn) error
}

// Connect opens a new connection to the database, applying the per-connection
// settings to it.
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := (&sqlite3.SQLiteDriver{}).Open(c.dsn)
	if err != nil {
		return nil, err
	}

	if err := c.init(ctx, conn); err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}

// Driver returns the driver used to open connections.
func (c *connector) Driver() driver.Driver {
	return &sqlite3.SQLiteDriver{}
}

// initialize applies the per-connection settings that cannot be expressed in
// the connection string.
func initialize(ctx context.Context, conn driver.Conn, config *Configuration) error {
	execer, ok := conn.(driver.ExecerContext)
	if !ok {
		return errors.New("litesql: connection does not support context execution")
	}

	// no connection string option for setting mmap_size; do it manually
	statement := fmt.Sprintf("pragma mmap_size = %d;", config.MemoryMapSize)
	if _, err := execer.ExecContext(ctx, statement, nil); err != nil {
		return fmt.Errorf("litesql: unable to set mmap_size: %w", err)
	}

	return nil
}

// openDB returns a database handle for the given connection string, with at
// most limit concurrent connections.
func openDB(uri string, config *Configuration, limit int) *sql.DB {
	connector := &connector{
		dsn: uri,
		init: func(ctx context.Context, conn driver.Conn) error {
			return initialize(ctx, conn, config)
		},
	}

	db := sql.OpenDB(connector)

	// set connection behaviors; max conns is defacto read concurrency
	db.SetMaxOpenConns(limit)
	db.SetConnMaxIdleTime(0) // disable
	db.SetConnMaxLifetime(0) // disable

	return db
}
