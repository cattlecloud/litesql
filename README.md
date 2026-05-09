# litesql

[![Go Reference](https://pkg.go.dev/badge/cattlecloud.net/go/litesql.svg)](https://pkg.go.dev/cattlecloud.net/go/litesql)
[![License](https://img.shields.io/github/license/cattlecloud/litesql?color=7C00D8&style=flat-square&label=License)](https://github.com/cattlecloud/litesql/blob/main/LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/cattlecloud/litesql/ci.yaml?style=flat-square&color=0FAA07&label=Tests)](https://github.com/cattlecloud/litesql/actions/workflows/ci.yaml)

`litesql` is a Go library for working with SQLite3, providing reasonable
defaults and an easy-to-use API for reliable and performant database access.

### Getting Started

The `litesql` package can be added to a Go project with `go get`.

```shell
go get cattlecloud.net/go/litesql@latest
```

```go
import "cattlecloud.net/go/litesql"
```

### Examples

#### Opening a SQLite database

The `litesql.TypicalConfiguration` contains reasonable defaults for many
general applications such as webapps. You may wish to use it as a reference
and fine-tune parameters for each use case.

```go
db, err := litesql.Open("/path/to/file", litesql.TypicalConfiguration)
// ...
defer db.Close()
```

#### Starting SQLite transactions

The `*LiteDB` returned by `Open` provides `StartRead` and `StartWrite` for
starting a read or write transaction. They make use of the `ReadConsistency`
and `WriteConsistency` package values to indicate isolation levels. A write
transaction must be ended with a call to `Commit`.

```go
// read transaction
tx, done, xerr := db.StartRead(ctx)
if xerr != nil {
    return xerr
}
defer done()

// ... use tx to execute queries ...
```

```go
// write transaction
tx, done, xerr = db.StartWrite(ctx)
if xerr != nil {
    return xerr
}
defer done()

// ... use tx to execute statements ...

// commit the write transaction
return tx.Commit()
```

#### Query rows

Convenience functions `QueryRow` and `QueryRows` exist at the package level
for abstracting much of the boiler-plate code for reading rows. By supplying
a `ScanFunc`, you can fetch row(s) without managing most of the query logic.

```go
func example(id int) ([]*record, error) {
  tx, done, xerr := db.StartRead(ctx)
  if xerr != nil {
    return nil, xerr
  }
  defer done()

  const statement = `select * from mytable where id > ?`

  f := func(sf litesql.ScanFunc) (*record, error) {
    r := new(record)
    err := sf(
      // &r.field1
      // &r.field2
      // ...
    )
    return r, err
  }

  return litesql.QueryRows(ctx, tx, f, statement, id)
}
```

#### Execute statement

The `ExecID` and `Exec` statements are for write transactions. The `Exec`
statement needs to know how many rows to expect to be modified, returning an
error if expectations are not met. There are special package constants for
indicating certain special cases. The `ExecID` method expects one row to be
changed, and will return the `ROWID` of the affected (or added) row.

```go
ExpectAnything   // do not enforce any expecation on number of rows changed
ExpectNonZero    // at least one row must be changed
ExpectOneOrZero  // exactly 0 or 1 row must be changed, useful for upserts
ExpectOne        // exactly 1 row must be changed
ExpectNone       // exactly 0 rows must be changed
```

A simple update example.

```go
func example(id int, value string) error {
  tx, done, xerr := db.StartWrite(ctx)
  if xerr != nil {
    return xerr
  }
  defer done()

  const statement = `update mytable set v = ? where id = ?`

  if err := db.Exec(ctx, tx, litesql.ExpectOneOrZero, statement, value, id); err != nil {
    return err
  }

  return tx.Commit()
}
```

#### Show pragma values

It is often helpful to dump the database pragma values on startup. This can
be done using the `Pragmas` method, which returns a `map[string]string` of
most common SQLite pragma configuration values.

```go
m, _ := db.Pragmas(ctx)

for k, v := range m {
  fmt.Println("pragma", k, "value", v)
}
```

#### Creating a snapshot

The `Snapshot` method creates a point-in-time copy of the database, useful
for backups or creating read-only copies for sharing. Snapshots are copied
page-by-page using SQLite's backup API, with configurable `Step` and `Gap`
options to control concurrency with writers.

```go
err := db.Snapshot(&litesql.SnapshotOptions{
    Directory:  "/path/to/snapshots",        // where to write the snapshot
    Retention:  3,                           // keep last 3 snapshots
    Step:       100,                         // copy 100 pages per step
    Gap:        1 * time.Millisecond,       // wait between steps
    Progress: func(pages, remaining int) {
        fmt.Printf("progress: %d/%d pages\n", pages-remaining, pages)
    },
})
```

The snapshot files are named `snapshot-<timestamp>.db` in the specified
directory, and older files are automatically cleaned up according to the
`Retention` policy.

### License

The `cattlecloud.net/go/litesql` module is opensource under the [BSD-3-Clause](LICENSE) license.
