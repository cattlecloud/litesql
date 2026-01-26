// Package libsqltest enables convenient testing with a real sqlite3 database
// backed by memory for use in unit tests.
package litesqltest

import (
	"testing"
	"time"

	"cattlecloud.net/go/litesql"
	"cattlecloud.net/go/scope"
)

const timeout = 3 * time.Second

// TestConfiguration provides PRAGMA settings appropriate for an in-memory
// database used in unit tests.
//
// Do not modify.
var TestConfiguration = &litesql.Configuration{
	Mode:                   "rwc",
	Encoding:               "utf8",
	BusyTimeout:            500,
	TransactionLock:        "immediate",
	ForeignKeys:            true,
	JournalMode:            "OFF",
	CacheSize:              -4000,
	AutoVacuum:             "incremental",
	Synchronous:            "normal",
	MemoryMapSize:          0,
	MaxConnectionsOpen:     1,
	MaxConnectionsIdleTime: 0,
	MaxConnectionsLifeTime: 0,
}

func Open(t *testing.T) *litesql.LiteDB {
	t.Helper()

	_, cancel := scope.WithTTL(t.Context(), timeout)
	t.Cleanup(cancel)

	const filename = ":memory:"

	ldb, oerr := litesql.Open(filename, TestConfiguration)
	if oerr != nil {
		t.Log("unable to open database: " + oerr.Error())
		t.FailNow()
	}

	t.Cleanup(func() { _ = ldb.Close() })

	return ldb
}
