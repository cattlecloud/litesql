package litesql

import (
	"io"
	"os"
	"testing"
	"time"

	"cattlecloud.net/go/scope"
	"github.com/shoenig/test/must"
)

const (
	timeout = 3 * time.Second
)

var testConfiguration = &Configuration{
	Mode:                   "rwc",
	Encoding:               "utf8",
	BusyTimeout:            1000,
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

func testSimple(t *testing.T) *LiteDB {
	t.Helper()

	// give ourselves a short timeout; all in memory
	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	t.Cleanup(cancel)

	// create the underlying sqlite3 database
	ldb, oerr := Open(":memory:", testConfiguration)
	must.NoError(t, oerr, must.Sprint("unable to open test database"))

	// open the sample schema
	f, ferr := os.Open("hack/simple.sql")
	must.NoError(t, ferr, must.Sprint("unable to open simple.sql file"))

	b, berr := io.ReadAll(f)
	must.NoError(t, berr, must.Sprint("unable to read simple.sql file"))
	stmt := string(b)

	// execute the sample schema and populate our test database
	_, xerr := ldb.db.ExecContext(ctx, stmt)
	must.NoError(t, xerr)

	// make sure we close the database once test is complete
	t.Cleanup(func() { _ = ldb.Close() })

	return ldb
}

func TestLiteDB_QueryRow(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartRead(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `SELECT COUNT(id) FROM users`

	row := ldb.QueryRow(ctx, tx, stmt)

	var count int
	serr := row.Scan(&count)
	must.NoError(t, serr)
	rerr := row.Err()
	must.NoError(t, rerr)

	must.Eq(t, 5, count)
}

type pair[T, U any] struct {
	A T
	B U
}

func TestLiteDB_QueryRows(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartRead(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `
	SELECT
		id, provider
	FROM
		oauth
	ORDER BY id DESC`

	rows, rdone, rerr := ldb.QueryRows(ctx, tx, stmt) // nolint:sqlclosecheck
	must.NoError(t, rerr)
	defer rdone()

	results := make([]pair[int, string], 0, 4)

	for rows.Next() {
		var (
			id       int
			provider string
		)

		serr := rows.Scan(&id, &provider)
		must.NoError(t, serr, must.Sprint("failure to scan row"))
		results = append(results, pair[int, string]{A: id, B: provider})
	}

	must.NoError(t, rows.Err())
	must.Eq(t, []pair[int, string]{
		{A: 4, B: "Microsoft"},
		{A: 3, B: "Google"},
		{A: 2, B: "Apple"},
		{A: 1, B: "Testing"},
	}, results)
}

func TestLiteDB_ExecID(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartWrite(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `
	INSERT INTO users (
		oauth_id, username, email
	) VALUES (?, ?, ?)`

	const (
		oid      = 2
		username = "ned"
		email    = "ned@example.org"
	)
	id, ierr := ldb.ExecID(ctx, tx, stmt, oid, username, email)
	must.NoError(t, ierr)
	must.Eq(t, 6, id) // 6th user insertion

	cerr := tx.Commit()
	must.NoError(t, cerr)
}

func TestLiteDB_Exec(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartWrite(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `
	DELETE FROM users
	WHERE
		oauth_id = ?`

	const oid = 4
	derr := ldb.Exec(ctx, tx, ExpectNonZero, stmt, oid)
	must.NoError(t, derr)

	cerr := tx.Commit()
	must.NoError(t, cerr)
}
