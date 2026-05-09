package litesql

import (
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"cattlecloud.net/go/scope"
	"github.com/shoenig/test/must"
)

func TestLiteDB_Snapshot(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), 10*time.Second)
	defer cancel()

	directory := t.TempDir()

	dbfile := filepath.Join(directory, "test.db")
	ldb, oerr := Open(dbfile, TypicalConfiguration)
	must.NoError(t, oerr)
	t.Cleanup(func() { _ = ldb.Close() })

	schema := `
	create table if not exists events (
		id integer primary key,
		data blob
	) strict;`
	_, xerr := ldb.db.ExecContext(ctx, schema)
	must.NoError(t, xerr)

	tx, xdone, xerr := ldb.StartWrite(ctx)
	must.NoError(t, xerr)

	for range 100 {
		eerr := ldb.Exec(ctx, tx, ExpectNonZero, "INSERT INTO events (data) VALUES (CAST(randomblob(1024) AS BLOB))")
		must.NoError(t, eerr)
	}
	must.NoError(t, tx.Commit())
	xdone()

	var progress atomic.Int64
	opts := &SnapshotOptions{
		Directory: directory,
		Step:      1,
		Gap:       1 * time.Millisecond,
		Progress:  func(_, _ int) { progress.Add(1) },
	}

	serr := ldb.Snapshot(ctx, opts)
	must.NoError(t, serr)
	must.Positive(t, progress.Load())

	matches, merr := filepath.Glob(filepath.Join(directory, "snapshot-*.db"))
	must.NoError(t, merr)
	must.SliceLen(t, 1, matches)
}

func TestLiteDB_SnapshotRetention(t *testing.T) {
	t.Parallel()

	ctx, cancel := scope.WithTTL(t.Context(), 10*time.Second)
	defer cancel()

	directory := t.TempDir()

	dbfile := filepath.Join(directory, "test.db")
	ldb, oerr := Open(dbfile, TypicalConfiguration)
	must.NoError(t, oerr)
	t.Cleanup(func() { _ = ldb.Close() })

	schema := `
	create table if not exists events (
		id integer primary key,
		data blob
	) strict;`
	_, xerr := ldb.db.ExecContext(ctx, schema)
	must.NoError(t, xerr)

	tx, xdone, xerr := ldb.StartWrite(ctx)
	must.NoError(t, xerr)

	for range 100 {
		eerr := ldb.Exec(ctx, tx, ExpectNonZero, "INSERT INTO events (data) VALUES (CAST(randomblob(1024) AS BLOB))")
		must.NoError(t, eerr)
	}
	must.NoError(t, tx.Commit())
	xdone()

	opts := &SnapshotOptions{
		Directory: directory,
		Retention: 2,
	}

	for i := range 5 {
		serr := ldb.Snapshot(ctx, opts)
		must.NoError(t, serr)

		time.Sleep(1100 * time.Millisecond)

		matches, merr := filepath.Glob(filepath.Join(directory, "snapshot-*.db"))
		must.NoError(t, merr)

		expected := min(i+1, opts.Retention+1)
		must.SliceLen(t, expected, matches)
	}
}
