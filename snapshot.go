package litesql

import (
	"cmp"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cattlecloud.net/go/scope"
	"github.com/hashicorp/go-set/v3"
	"github.com/mattn/go-sqlite3"
	"github.com/shoenig/lang"
)

// SnapshotOptions defines options for creating a point-in-time snapshot of the
// database. Snapshots create a full copy of the database at a specific moment in time,
// which can be useful for backups, point-in-time recovery, or creating read-only
// copies of the database for sharing.
//
// The snapshot is performed by copying the database page-by-page using SQLite's
// backup API. By default, the entire database is copied in a single operation,
// which is fast but may cause noticeable I/O pauses for other write operations.
// To enable concurrent writers to make progress, configure Step to copy a smaller
// number of pages at a time with a Gap between each step.
//
// Snapshot database files are created in the specified Directory with timestamps
// in the filename (snapshot-<timestamp>.db). The Retention option controls how many
// previous snapshot files to keep around after each snapshot completes.
type SnapshotOptions struct {
	// The Directory in which to create the snapshot database file(s).
	//
	// Snapshot database filenames are in the form "snapshot-<timestamp>.db".
	// and the retention policy makes use of these timestamps to determine
	// which files to keep around or purge.
	Directory string

	// The Retention policy indicates the number of additional previous snapshot
	// database files to keep around after completion.
	//
	// Defaults to 0 (purge all but the latest snapshot).
	Retention int

	// Step indicates the number of pages to copy in each operation.
	//
	// A smaller step size (<128) will take longer to create the snapshot, but
	// will enable writers to make concurrent progress.
	//
	// A larger step size (>1024) will be fast, but may cause noticeable
	// pauses for other write operations.
	//
	// Use -1 to snapshot the entire database in a single operation. Default.
	Step int

	// Gap is amount of time to wait in between each step operation.
	//
	// A larger gap enables more write operations to happen concurrently, but
	// will cause more duplicate work, as each page that gets written to will
	// need to be copied again.
	//
	// Defaults to 1 millisecond.
	Gap time.Duration

	// Progress is an optional callback that is invoked after each step interval
	// for updating, providing information about the total number of pages and
	// the number of pages remaining to be copied.
	Progress ProgressFunc
}

func (opts SnapshotOptions) defaults() *SnapshotOptions {
	result := new(SnapshotOptions)
	result.Directory = opts.Directory
	result.Retention = opts.Retention
	result.Step = lang.Maybe(opts.Step == 0, -1, opts.Step)
	result.Gap = lang.Maybe(opts.Gap == 0, 1*time.Millisecond, opts.Gap)
	result.Progress = lang.Maybe(opts.Progress == nil, func(int, int) {}, opts.Progress)
	return result
}

// ProgressFunc is a callback invoked between each snapshot step operation
// indicating the total number of pages, and the number of pages to be copied
// remaining.
type ProgressFunc func(pages, remaining int)

// Snapshot creates a point-in-time snapshot of the database, copying all pages to
// a new database file in the specified directory.
//
// The snapshot is performed by acquiring a connection to the database and using
// SQLite's backup API to copy pages to the new database. The Progress callback
// (if provided) is invoked after each step interval with the total number of
// pages and the number of pages remaining to be copied.
//
// After the snapshot completes, older snapshot files are automatically cleaned up
// according to the Retention policy.
//
// Example:
//
//	ctx := context.Background()
//	err := db.Snapshot(ctx, &SnapshotOptions{
//	    Directory:  "/path/to/snapshots",
//	    Retention: 3,
//	    Step:      100,
//	    Gap:       1 * time.Millisecond,
//	    Progress: func(pages, remaining int) {
//	        fmt.Printf("copied %d of %d pages\n", pages-remaining, pages)
//	    },
//	})
func (ldb *LiteDB) Snapshot(ctx scope.C, opts *SnapshotOptions) error {
	connection, cerr := ldb.db.Conn(ctx)
	if cerr != nil {
		return fmt.Errorf("unable to acquire sqlite connection: %w", cerr)
	}
	defer func() { _ = connection.Close() }()

	return connection.Raw(func(raw any) error {
		if scon, ok := raw.(*sqlite3.SQLiteConn); ok {
			return ldb.create(ctx, opts.defaults(), time.Now(), scon)
		}
		return errors.New("not a sqlite3 connection")
	})
}

func (ldb *LiteDB) uri(directory string, now time.Time) (string, string) {
	parameters := strings.Join([]string{
		"mode=rwc",
		"encoding=utf8",
		"_txlock=immediate",
		"_foreign_keys=true",
	}, "&")
	file := fmt.Sprintf("%s/snapshot-%d.db", directory, now.Unix())
	uri := fmt.Sprintf("file:%s?%s", file, parameters)
	return file, uri
}

func (ldb *LiteDB) create(ctx scope.C, opts *SnapshotOptions, now time.Time, scon *sqlite3.SQLiteConn) error {
	file, uri := ldb.uri(opts.Directory, now)

	if err := ldb.touch(file); err != nil {
		return fmt.Errorf("unable to touch snapshot file: %w", err)
	}

	dc, oerr := sql.Open("sqlite3", uri)
	if oerr != nil {
		return fmt.Errorf("unable to create snapshot database: %w", oerr)
	}

	dc.SetMaxOpenConns(1)    // only one writer
	dc.SetConnMaxIdleTime(0) // do not close connection
	dc.SetConnMaxLifetime(0) // do not close connection

	dcc, derr := dc.Conn(ctx)
	if derr != nil {
		return fmt.Errorf("unable to open snapshot database: %w", derr)
	}

	if err := dcc.Raw(func(raw any) error {
		if dcon, ok := raw.(*sqlite3.SQLiteConn); ok {
			return ldb.clone(scon, dcon, opts)
		}
		return nil // not a sqlite3 connection; nothing to do
	}); err != nil {
		return fmt.Errorf("unable to acqurie raw snapshot connection: %w", err)
	}

	if err := dcc.Close(); err != nil {
		return fmt.Errorf("unable to close snapshot database: %w", err)
	}

	return ldb.cleanup(opts)
}

func (ldb *LiteDB) touch(path string) error {
	// ensure the snapshot file exists with owner-only permissions
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	return file.Close()
}

func (ldb *LiteDB) clone(primary, snapshot *sqlite3.SQLiteConn, opts *SnapshotOptions) error {
	bu, berr := snapshot.Backup("main", primary, "main")
	if berr != nil {
		return fmt.Errorf("unable to start snapshot backup: %w", berr)
	}

	for range time.Tick(opts.Gap) {
		done, err := bu.Step(opts.Step)
		if err != nil {
			return fmt.Errorf("unable to step snapshot backup: %w", err)
		}

		if done {
			break
		}

		opts.Progress(bu.PageCount(), bu.Remaining())
	}

	return bu.Finish()
}

func (ldb *LiteDB) cleanup(opts *SnapshotOptions) error {
	matches, err := filepath.Glob(filepath.Join(opts.Directory, "snapshot-*.db"))
	if err != nil {
		return fmt.Errorf("unable to list snapshot files: %w", err)
	}

	tree := set.TreeSetFrom(matches, cmp.Compare[string])
	remove := max(0, tree.Size()-opts.Retention-1)
	deletions := tree.TopK(remove)

	for _, filename := range deletions {
		_ = os.Remove(filename)
	}

	return nil
}
