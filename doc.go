// Package litesql implements a SQLite3 interface with reasonable defaults,
// making interacting with sqlite3 databases easy, reliable, and performant in
// Go programs.
package litesql

import _ "github.com/mattn/go-sqlite3" // must include the underlying sqlite3 driver implementation
