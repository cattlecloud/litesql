package litesql

import "fmt"

const (
	// ExecFailure is a sentinel value used when an Exec statement did not
	// complete correctly and we have no resulting row ID.
	ExecFailure = -1

	// TxFailure is a sentinel value used when the creation of a transaction
	// fails and the query cannot continue, and we have no row ID.
	TxFailure = -2
)

// ID represents a unique ROW ID of a table.
//
// Most often, this is coming from an autoincrement index.
//
// A common pattern is to create a new type based on this type, so as to
// leverage the type system ensuring ID values are not co-mingled.
type ID int

func (id ID) String() string {
	switch id {
	case ExecFailure:
		return "exec-failure"
	case TxFailure:
		return "tx-failure"
	default:
		return fmt.Sprintf("id:%d", id)
	}
}
