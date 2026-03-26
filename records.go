package litesql

const (
	// ExecFailure is a sentinel value used when an Exec statement did not
	// complete correctly and we have no resulting row ID.
	ExecFailure = -1
)

// ID represents a unique ROW ID of a table.
//
// Most often, this is coming from an autoincrement index.
//
// A common pattern is to create a new type based on this type, so as to
// leverage the type system ensuring ID values are not co-mingled.
type ID int
