package litesql

// Configuration is the set of PRAGMA values and/or connection parameters used
// to configure the sqlite database.
//
// Most general purpose applications such as web servers and api servers may
// want to start with the default values provided in TypicalConfiguration.
type Configuration struct {
	// Mode configures the read/write/create capability.
	//
	// Use "rwc" for most use cases.
	Mode string

	// Encoding configures the encoding pragma. Value must be one of:
	//
	//  - UTF-8
	//  - UTF-16
	//  - UTF-16le
	//  - UTF-16be
	//
	// https://sqlite.org/pragma.html#pragma_encoding
	Encoding string

	// BusyTimeout configures the busy_timeout pragma. Value in milliseconds.
	//
	// https://sqlite.org/pragma.html#pragma_busy_timeout
	BusyTimeout int

	// TransactionLock configures the transaction lock mode. Value must be one
	// of:
	//
	//  - deferred
	//  - immediate
	//  - exclusive
	//
	// This is the mode used by the write connection; read connections always
	// use the deferred mode.
	//
	// https://sqlite.org/pragma.html#pragma_txlock
	TransactionLock string

	// ForeighKeys configures the foreign_keys pragma. Value must be one of:
	//
	//  - true
	//  - false
	//
	// https://sqlite.org/pragma.html#pragma_foreign_keys
	ForeignKeys bool

	// JournalMode configures the journal_mode pragma. Value must be one of:
	//
	//  - DELETE
	//  - TRUNCATE
	//  - PERSIST
	//  - MEMORY
	//  - WAL
	//  - OFF
	//
	// https://sqlite.org/pragma.html#pragma_journal_mode
	JournalMode string

	// CacheSize configures the cache_size pragma.
	//
	// When set to a positive value, the size is interpreted in pages.
	// When set to a negative value, the size is interpreted in kibibytes.
	//
	// e.g. 100 => 4kb page size * 100 => 400,000 bytes of memory
	// e.g. -2000 => 2048000 bytes of memory
	//
	// The page cache is per connection, so the total memory used by the
	// database is this value multiplied by the number of connections.
	//
	// https://sqlite.org/pragma.html#pragma_cache_size
	CacheSize int

	// AutoVacuum configures the auto_vacuum pragma. Value must be one of:
	//
	//  - NONE
	//  - FULL
	//  - INCREMENTAL
	//
	// https://sqlite.org/pragma.html#pragma_auto_vacuum
	AutoVacuum string

	// Synchronous configures the synchronous pragma. Value must be one of:
	//
	//  - OFF
	//  - NORMAL
	//  - FULL
	//  - EXTRA
	//
	// https://sqlite.org/pragma.html#pragma_synchronous
	Synchronous string

	// MemoryMapSize configures the mmap_size pragma.
	//
	// Using mmap bypasses the kernel memory buffer, reducing the amount of
	// memory copying and syscall overhead. Understand the risks before using.
	//
	// The memory map is per connection, so the total address space used by
	// the database is this value multiplied by the number of connections.
	//
	// Not used for in-memory databases.
	//
	// https://sqlite.org/pragma.html#pragma_mmap_size
	MemoryMapSize int

	// MaxConnectionsOpen configures the maximum number of concurrent read
	// connections to the database.
	//
	// One additional connection is used for write transactions.
	MaxConnectionsOpen int
}

var (
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
	//   - cache_size: 16 megabytes per connection
	//   - mmap_size: 16 megabytes per connection
	//
	// Do not modify.
	TypicalConfiguration = &Configuration{
		Mode:               "rwc",
		Encoding:           "utf8",
		BusyTimeout:        5000,
		TransactionLock:    "immediate",
		ForeignKeys:        true,
		JournalMode:        "WAL",
		CacheSize:          -16384,
		AutoVacuum:         "incremental",
		Synchronous:        "normal",
		MemoryMapSize:      16777216,
		MaxConnectionsOpen: 4,
	}
)
