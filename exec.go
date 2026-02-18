package litesql

import (
	"database/sql"
	"errors"
	"fmt"

	"cattlecloud.net/go/scope"
)

const (
	// ExpectAnything indicates there is no expectation for the number of
	// rows that will be updated as a result of executing a statement.
	ExpectAnything = -(iota + 1)

	// ExpectNonZero indicates the expectation for the number of rows that will
	// be updated is non-zero.
	ExpectNonZero

	// ExpectOneOrZero indicates the expectation for the number of rows that
	// will be updated is exactly 0 or 1 (e.g. insert or ignore)
	ExpectOneOrZero
)

// ExecID executes the given sql query statement with args, and returns the
// resulting row id. The query must be intended to insert/modify exactly one
// row.
func (ldb *LiteDB) ExecID(ctx scope.C, tx *sql.Tx, stmt string, args ...any) (ID, error) {
	result, xerr := tx.ExecContext(ctx, stmt, args...)
	if xerr != nil {
		return ExecFailure, fmt.Errorf("litesql: failed to execute query: %w", xerr)
	}

	affected, aerr := result.RowsAffected()
	if aerr != nil {
		return ExecFailure, fmt.Errorf("litesql: failed to get rows affected: %w", aerr)
	}

	if affected != 1 {
		return ExecFailure, fmt.Errorf("litesql: expected to affect 1 row, actual was %d", affected)
	}

	inserted, ierr := result.LastInsertId()
	if ierr != nil {
		return ExecFailure, fmt.Errorf("litesql: failed to get last insert id: %w", ierr)
	}

	return ID(inserted), nil
}

// Exec executes the given sql query statement with args, and compares the
// number of rows affected with the given expectation. An error is returned if
// the number of rows does not match the given expectation. The constant values
// ExpectAnything, ExpectNonZero, and ExpectOneOrZero can be used for more
// complex, but common expected behaviors.
func (ldb *LiteDB) Exec(ctx scope.C, tx *sql.Tx, expectation int, stmt string, args ...any) error {
	result, xerr := tx.ExecContext(ctx, stmt, args...)
	if xerr != nil {
		return fmt.Errorf("litesql: failed to execute query: %w", xerr)
	}

	affected, aerr := result.RowsAffected()
	if aerr != nil {
		return fmt.Errorf("litesql: failed to get rows affected: %w", aerr)
	}

	switch expectation {
	case ExpectNonZero:
		if affected == 0 {
			return errors.New("litesql: expected to affect at least one row")
		}
	case ExpectOneOrZero:
		if affected != 0 && affected != 1 {
			return fmt.Errorf("litesql: expected to affect 0 or 1 row, actual: %d", affected)
		}
	case ExpectAnything:
		return nil
	default:
		if affected != int64(expectation) {
			return fmt.Errorf("litesql: expected to affect %d rows, actual: %d", expectation, affected)
		}
	}

	return nil
}

// QueryRow executes the given sql query statement with the expectation of
// returning exactly one row.
func (ldb *LiteDB) QueryRow(ctx scope.C, tx *sql.Tx, stmt string, args ...any) *sql.Row {
	return tx.QueryRowContext(ctx, stmt, args...)
}

// QueryRows executes the given sql query statement with the expectation of
// returning any number of rows.
//
// Must call the returned CloseFunc when finished; otherwise a connection will
// be consumed and not returned to the connection pool, causing future operations
// to hang indefinitely.
func (ldb *LiteDB) QueryRows(ctx scope.C, tx *sql.Tx, stmt string, args ...any) (*sql.Rows, CloseFunc, error) {
	cursor, cerr := tx.QueryContext(ctx, stmt, args...)
	closer := func() { _ = cursor.Close() }
	return cursor, closer, cerr
}

// ScanFunc represents the sql.Scan function from a sql.Rows object. This is
// used as the scan argument of the QueryRows package function, invoked on
// each element in the result set of the query to build the list of resulting
// items.
type ScanFunc func(args ...any) error

// QueryRow uses the given transaction tx and the scan function to extract a
// single row from the database, using the given stmnt query with args.
//
// The scan function is provided by the caller for custom extraction of column
// values into some type T.
func QueryRow[T any](ctx scope.C, tx *sql.Tx, scan func(ScanFunc) (T, error), stmt string, args ...any) (T, error) {
	row := tx.QueryRowContext(ctx, stmt, args...)

	t, terr := scan(row.Scan)
	if terr != nil {
		var zero T
		return zero, terr
	}

	return t, nil
}

// QueryRows uses the given transaction tx and the scan function to extract a
// set of rows from the database, using the given stmnt query with args.
//
// The scan function is provided by the caller for custom extraction of column
// values into some type T.
func QueryRows[T any](ctx scope.C, tx *sql.Tx, scan func(ScanFunc) (T, error), stmt string, args ...any) ([]T, error) {
	rows, rerr := tx.QueryContext(ctx, stmt, args...)
	if rerr != nil {
		return nil, rerr
	}
	defer func() { _ = rows.Close() }()

	items := make([]T, 0, 8)

	for rows.Next() {
		t, terr := scan(rows.Scan)
		if terr != nil {
			return nil, terr
		}
		items = append(items, t)
	}

	return items, rows.Err()
}
