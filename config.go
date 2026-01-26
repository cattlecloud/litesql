package litesql

// Configuration is the set of PRAGMA values used to configure the sqlite
// database, and also some connection tuning parameters.
//
// Most general purpose applications such as web servers and api servers may
// want to use or start with the TypicalConfiguration default values.
type Configuration struct {
	Mode string

	Encoding string

	BusyTimeout int

	TransactionLock string

	ForeignKeys bool

	JournalMode string

	CacheSize int

	AutoVacuum string

	Synchronous string

	MemoryMapSize int

	MaxConnectionsOpen int

	MaxConnectionsIdleTime int

	MaxConnectionsLifeTime int
}

// TypicalConfiguration should be suitable for use by most applications, such
// as web servers, api servers, and other general purpose scenarios. In
// particular it sets these parameter values:
//
//   - encoding: utf8
//   - journal_mode: WAL
//   - foreign_keys: true
//   - transaction_lock: immediate
//   - auto_vacuum: incremental
//   - synchronous: normal
//   - cache_size: 64 megabytes
//   - mmap_size: 64 megabytes
//
// Do not modify.
var TypicalConfiguration = &Configuration{
	Mode:                   "rwc",
	Encoding:               "utf8",
	BusyTimeout:            5000,
	TransactionLock:        "immediate",
	ForeignKeys:            true,
	JournalMode:            "WAL",
	CacheSize:              -65536,
	AutoVacuum:             "incremental",
	Synchronous:            "normal",
	MemoryMapSize:          67108864,
	MaxConnectionsOpen:     4,
	MaxConnectionsIdleTime: 0,
	MaxConnectionsLifeTime: 0,
}
