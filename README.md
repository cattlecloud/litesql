# litesql

[![Go Reference](https://pkg.go.dev/badge/cattlecloud.net/go/litesql.svg)](https://pkg.go.dev/cattlecloud.net/go/litesql)
[![License](https://img.shields.io/github/license/cattlecloud/litesql?color=7C00D8&style=flat-square&label=License)](https://github.com/cattlecloud/litesql/blob/main/LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/cattlecloud/litesql/ci.yaml?style=flat-square&color=0FAA07&label=Tests)](https://github.com/cattlecloud/litesql/actions/workflows/ci.yaml)

The `litesql` Go library provides a convenient interface for working with
SQLite3 in Go programs, so that they are reliable and performant, by making
use of reasonable defaults and providing an easy API to build on top of.

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
db.Close()
```

#### Starting SQLite transactions

The `*LiteDB` returned by `Open` provides `StartRead` and `StartWrite` for
starting a read or write transaction. They make use of the `ReadConsistency`
and `WriteConsistency` package values to indicate isolation levels. A write
transaction must be ended with a call to `Commit`.

#### Query rows

Convenience functions `QueryRow` and `QueryRows` exist at the package level
for abstracting much of the boiler-plate code for reading rows. By supplying
a `ScanFunc`, you can fetch row(s) without managing most of the query logic.

```go
func example(id int) ([]record, error) {
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
ExpectAnything // do not enforce any expecation on number of rows changed
ExpectNonZero // at least one row must be changed
ExpectOneOrZero // exactly 0 or 1 row must be changed, useful for upserts
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

### License

The `cattlecloud.net/go/litesql` module is opensource under the [BSD-3-Clause](LICENSE) license.
