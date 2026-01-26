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

	// TransactionLock configures the locking_mode pragma. Value must be one of:
	//
	//  - NORMAL
	//  - EXCLUSIVE
	//
	// https://sqlite.org/pragma.html#pragma_locking_mode
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
	// Not used for in-memory databases.
	//
	// https://sqlite.org/pragma.html#pragma_mmap_size
	MemoryMapSize int

	// MaxConnectionsOpen configures the maximum number of concurrent users
	// of the database.
	MaxConnectionsOpen int
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
	Mode:               "rwc",
	Encoding:           "utf8",
	BusyTimeout:        5000,
	TransactionLock:    "immediate",
	ForeignKeys:        true,
	JournalMode:        "WAL",
	CacheSize:          -65536,
	AutoVacuum:         "incremental",
	Synchronous:        "normal",
	MemoryMapSize:      67108864,
	MaxConnectionsOpen: 4,
}
