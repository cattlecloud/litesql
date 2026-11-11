// Package litesqltest enables convenient testing with a real sqlite3 database
// for use in unit tests.
package litesqltest

import (
	"path/filepath"
	"testing"

	"cattlecloud.net/go/litesql"
)

var (
	// TestConfiguration provides PRAGMA settings appropriate for a database used
	// in unit tests. This configuration uses a smaller page cache and disables
	// memory-mapped I/O to keep resource usage low.
	//
	// Do not modify.
	TestConfiguration = &litesql.Configuration{
		Mode:               "rwc",
		Encoding:           "utf8",
		BusyTimeout:        500,
		TransactionLock:    "immediate",
		ForeignKeys:        true,
		JournalMode:        "WAL",
		CacheSize:          -4000,
		AutoVacuum:         "incremental",
		Synchronous:        "normal",
		MemoryMapSize:      0,
		MaxConnectionsOpen: 2,
	}
)

// Open creates a database in a temporary directory for use in unit tests.
//
// The database is backed by a file, which is removed when the test completes.
func Open(t *testing.T) *litesql.LiteDB {
	t.Helper()

	filename := filepath.Join(t.TempDir(), "test.db")

	ldb, oerr := litesql.Open(filename, TestConfiguration)
	if oerr != nil {
		t.Log("unable to open database: " + oerr.Error())
		t.FailNow()
	}

	t.Cleanup(func() { _ = ldb.Close() })

	return ldb
}
