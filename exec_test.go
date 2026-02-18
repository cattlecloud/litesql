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
	Mode:               "rwc",
	Encoding:           "utf8",
	BusyTimeout:        1000,
	TransactionLock:    "immediate",
	ForeignKeys:        true,
	JournalMode:        "OFF",
	CacheSize:          -4000,
	AutoVacuum:         "incremental",
	Synchronous:        "normal",
	MemoryMapSize:      0,
	MaxConnectionsOpen: 1,
}

type user struct {
	ID       int
	Username string
	Email    string
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

func TestGlobal_QueryRow(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartRead(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `
	SELECT
		id, username, email
	FROM
		users
	WHERE
		ID = ?`

	f := func(sf ScanFunc) (*user, error) {
		u := new(user)
		err := sf(
			&u.ID,
			&u.Username,
			&u.Email,
		)
		return u, err
	}

	const id = 2
	user, uerr := QueryRow(ctx, tx, f, stmt, id)
	must.NoError(t, uerr)
	must.Eq(t, 2, user.ID)
	must.Eq(t, "Beth", user.Username)
	must.Eq(t, "beth@example.org", user.Email)
}

func TestGlobal_QueryRow_int(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartRead(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `
	SELECT standing_id FROM users WHERE id = ?`

	f := func(sf ScanFunc) (int, error) {
		var standing int
		err := sf(
			&standing,
		)
		return standing, err
	}

	const beth = 2 // (hardcode user id)
	standing, uerr := QueryRow(ctx, tx, f, stmt, beth)
	must.NoError(t, uerr)
	must.Eq(t, 3, standing)
}

func TestGlobal_QueryRows(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartRead(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `
	SELECT
		id, username, email
	FROM
		users
	ORDER BY id ASC`

	f := func(sf ScanFunc) (*user, error) {
		u := new(user)
		err := sf(
			&u.ID,
			&u.Username,
			&u.Email,
		)
		return u, err
	}

	users, uerr := QueryRows(ctx, tx, f, stmt)
	must.NoError(t, uerr)
	must.SliceLen(t, 5, users)
	must.Eq(t, &user{ID: 1, Username: "Admin", Email: "admin@example.org"}, users[0])
	must.Eq(t, &user{ID: 2, Username: "Beth", Email: "beth@example.org"}, users[1])
	must.Eq(t, &user{ID: 3, Username: "Carl", Email: "carl@example.org"}, users[2])
	must.Eq(t, &user{ID: 4, Username: "David", Email: "dave@example.org"}, users[3])
	must.Eq(t, &user{ID: 5, Username: "Eve", Email: "eve@example.org"}, users[4])
}

func TestGlobal_QueryRows_int(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), timeout)
	defer cancel()

	ldb := testSimple(t)

	tx, xdone, xerr := ldb.StartRead(ctx)
	must.NoError(t, xerr)
	defer xdone()

	const stmt = `SELECT id FROM oauth ORDER BY id DESC`

	f := func(sf ScanFunc) (int, error) {
		var id int
		err := sf(&id)
		return id, err
	}

	oauths, oerr := QueryRows(ctx, tx, f, stmt)
	must.NoError(t, oerr)
	must.Eq(t, []int{4, 3, 2, 1}, oauths)
}
