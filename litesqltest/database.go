package libsqltest

import (
	"testing"
	"time"

	"cattlecloud.net/go/litesql"
	"cattlecloud.net/go/scope"
)

const timeout = 3 * time.Second

var TestConfiguration = &litesql.Configuration{
	Mode:                   "rwc",
	Encoding:               "utf8",
	BusyTimeout:            500,
	TransactionLock:        "immediate",
	ForeignKeys:            true,
	JournalMode:            "OFF",
	CacheSize:              -65536,
	AutoVacuum:             "incremental",
	Synchronous:            "normal",
	MemoryMapSize:          4000,
	MaxConnectionsOpen:     1,
	MaxConnectionsIdleTime: 0,
	MaxConnectionsLifeTime: 0,
}

func Open(t *testing.T) *litesql.LiteDB {
	t.Helper()

	_, cancel := scope.WithTTL(t.Context(), timeout)
	t.Cleanup(cancel)

	filename := ":memory:"
	ldb, oerr := litesql.Open(filename, litesql.TypicalConfiguration)
	if oerr != nil {
		t.Log("unable to open database: " + oerr.Error())
		t.FailNow()
	}

	return ldb
}
