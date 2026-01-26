package litesql

const (
	// ExecFailure is a sentinel value used when an Exec statement did not
	// complete correctly and we have no resulting row ID.
	ExecFailure = -1
)

type ID int
