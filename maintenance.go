package litesql

import (
	"fmt"

	"cattlecloud.net/go/scope"
)

// Optimize runs SQLite's PRAGMA optimize command, which updates internal schema
// metadata used by the query planner. It should be called periodically (e.g., weekly
// or after bulk inserts/deletes) to maintain optimal query performance. The operation
// is fast and non-blocking.
func (ldb *LiteDB) Optimize(ctx scope.C) error {
	if _, err := ldb.db.ExecContext(ctx, `pragma optimize`); err != nil {
		return fmt.Errorf("problem running 'pragma optimize': %w", err)
	}
	return nil
}

// Analyze runs SQLite's ANALYZE command, which gathers statistics about table and
// index data distribution. These statistics help the query planner choose efficient
// execution plans. Run after significant data changes (e.g., large inserts, deletes,
// or bulk updates) to ensure the planner has accurate information. Fast and non-blocking.
func (ldb *LiteDB) Analyze(ctx scope.C) error {
	if _, err := ldb.db.ExecContext(ctx, `analyze`); err != nil {
		return fmt.Errorf("problem running 'analyze': %w", err)
	}
	return nil
}

// Vacuum rebuilds the database file, reclaiming unused space and defragmenting data.
// It should be run sparingly (e.g., monthly or after many deletions) as it requires
// exclusive database access and can be slow for large databases. Improves performance
// by reducing file size and greatly improving sequential I/O efficiency.
func (ldb *LiteDB) Vacuum(ctx scope.C) error {
	if _, err := ldb.db.ExecContext(ctx, `vacuum`); err != nil {
		return fmt.Errorf("problem running 'vacuum': %w", err)
	}
	return nil
}
