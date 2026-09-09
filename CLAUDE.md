# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

`jdb` (`github.com/celsiainternet/jdb`) is a Go database library that provides a unified interface over PostgreSQL, MySQL, and SQLite. It is a library, not an application. `cmd/` has no top-level `main.go`; it holds three independent tool binaries (see below), none of which is "the product."

Go version: **1.23.0** (see `.go-version`). Primary dependency: `github.com/celsiainternet/elvis` (utility, logging, JSON types, events).

## Build and run commands

```bash
# Format and build everything (there is no package at ./cmd itself — use ./cmd/...)
gofmt -w . && go build ./...

# Run the example/dev entry point (connects to a real DB via env vars, see below)
gofmt -w . && go run --race ./cmd/test

# Release a patch version (increments X.Y.Z+1, tags, and pushes tags)
git add . && git commit -m 'Update version' && ./version.sh --r

# Minor version bump (X.Y+1.0)
./version.sh --n

# Major version bump (X+1.0.0)
./version.sh --m
```

`cmd/` contains three separate `main` packages:
- `cmd/test` — a hand-edited scratch program for exercising the library against a live DB (loads cache/event/jdb, runs ad-hoc queries). Treat it as a sandbox, not a stable example.
- `cmd/install` — installs a fixed list of third-party Go dependencies via `go get` (used when bootstrapping a new consumer project, not for developing `jdb` itself).
- `cmd/create` — a Cobra CLI (`cmd/create/v2`) that scaffolds new microservice projects/models from templates under `cmd/create/v2/template`.

There are no test files in this repository (no `*_test.go`).

## Architecture

### Package layout

```
jdb/            Core library package
drivers/        One sub-package per DB engine (postgres, mysql, sqlite)
instances/      Singleton helper: dedupes a schema+model, generic CRUD
authorization/  Singleton feature package built on jdb: auth/session model
config/         Singleton feature package built on jdb: runtime config model
inbox/          Singleton feature package built on jdb: message inbox model
cmd/            Standalone tool binaries (test/install/create) — not the library
```

### Core package (`jdb/`)

The global singleton `conn *JDB` (initialized in `jdb/jdb.go:init`) holds all registered drivers and live `*DB` connections. Everything hangs off this singleton.

Key types:
- `DB` (`database.go`) — a live database connection; holds its `Driver`, `Schema` list, and `Model` list.
- `Schema` (`schema.go`) — a named namespace within a DB.
- `Model` (`model.go`) — a table definition with columns, indices, relations, and lifecycle hooks.
- `Column` (`column.go`) — a column/field definition with type and constraints.
- `Command` (`command.go`) — an immutable write operation (Insert/Update/Delete/Upsert); built by fluent methods on `Model`.
- `Ql` (`ql.go`) — an immutable read query; built by fluent methods on `Model` or `From()`.
- `Tx` (`tx.go`) — wraps `*sql.Tx` for transactional `Command` and `Ql` execution.
- `Driver` interface (`drivers.go`) — contract every engine implements: `Select`, `Count`, `Exists`, `Command`, `LoadModel`, `DropModel`, `EmptyModel`, `MutateModel`.

### Driver registration

Each driver package registers itself in its own `init()` via `jdb.Register(name, factoryFunc, defaultParams)`. Callers activate a driver with a blank import:

```go
import _ "github.com/celsiainternet/jdb/drivers/postgres"
```

Default connection parameters are read from environment variables inside `init()` (e.g., `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `APP_NAME`, `NODE_ID`, `DEBUG`).

### Connecting

```go
// From environment variables (uses the driver's init defaults)
db, err := jdb.Load()

// Explicit params
db, err := jdb.ConnectTo(jdb.ConnectParams{
    Driver:   "postgres",
    Name:     "myapp",
    UserCore: true,   // creates a "core" schema with metadata tables
    NodeId:   1,
    Params: &postgres.Connection{...},
})
```

When `UserCore: true`, `createCore()` builds a `core` schema used for internal metadata (model registry, sequences).

### Fluent query API

Queries are built via `From(model)` or directly on a `*Model`:

```go
// Package-level (model looked up by name)
items, err := jdb.From("schema.table").Where("active").Eq(true).Limit(10).All()

// Model-level
items, err := userModel.Where("active").Eq(true).All()
```

Write commands follow the same pattern:

```go
item, err  := userModel.Insert(data).Exec()
items, err := userModel.Update(data).Where("id").Eq(id).Exec()
items, err := userModel.Delete("id").Eq(id).Exec()
items, err := userModel.Upsert(data).Where("id").Eq(id).Exec()
items, err := userModel.Bulk(dataSlice).Exec()
```

Transactional variants use `ExecTx(tx)` instead of `Exec()`.

### HTTP handlers

`jdb.go` exposes four `http.HandlerFunc` values that can be mounted on any Chi (or standard) router:
- `ModelDefine` — describe a model/schema/db
- `ModelQuery` — run a `Ql` from a JSON body
- `ModelCommand` — run one or more commands from a JSON body
- `ModelDescribe` — describe an object by kind+name

### Package-level singleton pattern (`instances`, `authorization`, `config`, `inbox`)

These four packages all follow the same shape: an unexported package-level singleton (`inst`, `auth`, `cfg`, `inb`) holding a `*jdb.Schema` + `*jdb.Model`, populated once via a package `Load(db, schema, ...)` function that no-ops if already loaded, and built by a `Define(...)` function that defines the schema/model against `jdb` idempotently. Each package owns exactly one table/model and layers feature-specific behavior (CRUD helpers, HTTP handlers, events) on top — e.g. `authorization.Load` also wires itself into `elvis/middleware.SetAuthorizationStore`. When adding a similar feature package, follow this same `Load`/`Define`/singleton shape rather than inventing a new one.

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `DB_NAME` | `jdb` | Database name |
| `DB_DRIVER` | — | `postgres`, `mysql`, or `sqlite` |
| `DB_HOST` | `localhost` | DB host |
| `DB_PORT` | `5432` | DB port |
| `DB_USER` | `admin` | DB username |
| `DB_PASSWORD` | `admin` | DB password |
| `APP_NAME` | `jdb` | Application name (used in PG connection string) |
| `NODE_ID` | `0` | Node ID for distributed ID generation |
| `DEBUG` | `false` | Enable debug logging |
| `DB_VERSION` | `13` | PostgreSQL server version |
