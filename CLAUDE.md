# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

`jdb` (`github.com/celsiainternet/jdb`) is a Go database library: a declarative model layer plus a fluent query/command builder that a driver translates to SQL. It is a library, not an application. `cmd/` has no top-level `main.go`; it holds three independent tool binaries (see below), none of which is "the product."

**Only PostgreSQL is implemented.** `drivers/oracle` registers and connects (via `sijms/go-ora`) but every operation returns `"not implemented"`; `drivers/sqlite/` is an empty directory; there is no `drivers/mysql`. `MysqlDriver`/`SqliteDriver` exist only as name constants in `jdb/drivers.go`. `README.md` still advertises MySQL/SQLite — it is stale (see "Doc drift").

Go version: **1.23.0** (see `.go-version`). Primary dependency: `github.com/celsiainternet/elvis` (utility, logging, JSON types `et.Json`, events, cache, middleware).

## Workspace (which `elvis` and which consumers)

`go env GOWORK` resolves to **`/Users/cesargalvisleon/Projects/internet/go.work`** (two levels above this repo, not `library/`). It `use`s `./library/elvis`, `./library/jdb` and ~15 services, so:
- `elvis` is built from the **local `library/elvis` checkout**, not the `v1.1.311` pinned in `go.mod`. Edits in either repo are live for the other with no tag/replace.
- Workspace modules that import `jdb`: `catalog/api` (pins v1.0.92), `octopus/api` (v1.0.94), `services/publisher` (v1.0.76), `services/cleandb` (v1.0.15). In workspace mode they compile against *this* checkout regardless of the pinned version, so a public signature change here can break them — grep those directories before changing an exported API. (`services/maia`, `services/bulkfailures` also pin jdb but are not in `go.work`.)
- The outer `library/` directory is not a Git repo; run `git` from inside `jdb/`.

## Build and run commands

```bash
# Format and build everything (there is no package at ./cmd itself — use ./cmd/...)
gofmt -w . && go build ./...
go vet ./...        # clean at HEAD

# Run the dev sandbox (needs a live Postgres via env vars, see below)
gofmt -w . && go run --race ./cmd/test

# Release a patch version (increments X.Y.Z+1, rewrites the README badge, tags, pushes tags)
git add . && git commit -m 'Update version' && ./version.sh --r
./version.sh --n    # minor bump (X.Y+1.0)
./version.sh --m    # major bump (X+1.0.0)
```

`cmd/` contains three separate `main` packages:
- `cmd/test` — a hand-edited scratch program (currently just `jdb.Load()` + `jdb.GetSeries`). Treat it as a sandbox.
- `cmd/install` — installs a fixed list of third-party Go dependencies via `go get` (bootstrapping a new consumer project).
- `cmd/create` — a Cobra CLI (`cmd/create/v2`) that scaffolds microservice projects/models from templates under `cmd/create/v2/template`.

**There are no tests** (`*_test.go`). Commits on `main` are auto-generated "Backup:" snapshots, so `git log` says little; use `git log -L` / `git show` on the file to find when behavior changed.

### Checking generated SQL without a database

No DB is needed to see what SQL the driver builds: the driver stores it in `Ql.Sql` / `Command.Sql` *before* it executes, and with no connection `db.Db` is nil, so execution panics afterwards. Recipe used to verify this file:
1. Make a throwaway module in the scratchpad (not the repo) with a `go.work` that `use`s `.`, `library/elvis` and `library/jdb`; run with `GOWORK=<that file> go run .` (`-mod=mod` is not allowed in workspace mode).
2. `db, _ := jdb.NewDatabase("x", "postgres")` (blank-import `drivers/postgres`), `jdb.NewSchema`, `jdb.NewModel`, `DefineColumn`… — **skip `model.Init()`** (it hits the DB).
3. Wrap execution in `defer func(){ recover() }()` and print `q.Sql`. For commands, set `c.New = <data>` and call `db.Command(c)` directly — `c.Exec()` first runs a SELECT (`getCurrent`) and panics before reaching the write.

## Architecture

### Package layout

```
jdb/            Core library package (declarative intent: Model/Ql/Command/Field/Where)
drivers/        postgres (complete), oracle (stub, all "not implemented"), sqlite (empty dir)
instances/      Feature singleton: generic keyed blob store
authorization/  Feature singleton: auth/session model (wires into elvis/middleware)
config/         Feature singleton: runtime config model
inbox/          Feature singleton: message inbox model
cmd/            Standalone tool binaries (test/install/create) — not the library
```

### Two layers

1. **`jdb` package** records *intent* only: `Model`/`Column`/`Field`, `Ql` (read), `Command` (write), `QlWhere`/`QlCondition`/`Value`, `Agregation`. It emits no SQL.
2. **A driver** (`jdb.Driver` interface in `jdb/drivers.go`: `Name`, `Connect`, `LoadModel`, `DropModel`, `EmptyModel`, `MutateModel`, `Select`, `Count`, `Exists`, `Command`) turns that intent into SQL text, stores it on `ql.Sql`/`command.Sql`, and runs it through `jdb.QueryTx`. In `drivers/postgres`: `query.go` (SELECT/WHERE/JOIN/GROUP/ORDER/LIMIT), `command.go` (INSERT/UPDATE/DELETE), `quote.go` (value → SQL literal), `ddl-*.go`, `database.go`/`driver.go` (connection, `Params` type).

**Values are inlined, not parameterized.** `quote()` renders literals (strings only get `'` doubled); the only bound argument is the jsonb source blob (`$1`) on INSERT. `CALC(...)` and string operands of aggregations are embedded as raw SQL (`agregationValue`). Never pass untrusted input as a `CALC`/aggregation string or as an unresolved field name.

### Core types (`jdb/`)

- Global singleton `conn *JDB` (`jdb.go:init`) holds registered drivers/params and live `*DB`s.
- `DB` (`database.go`) — connection + `Driver` + schemas/models. `Schema` (`schema.go`) — namespace. `Model` (`model.go`, `model-define.go`) — table definition; `NewModel(schema, name, version)` dedupes by name, `Define*` methods add columns/keys/indices/relations/rollups, `Init()` creates the table via the driver. `Column` (`column.go`) has a `TypeColumn`: `TpColumn` (real column), `TpAtribute` (key inside the jsonb source column), `TpRollup`, `TpRelatedTo`, `TpDetail`, `TpCalc`.
- `Field` (`field.go`) is the resolved reference to a column/attribute: `{Model *QlFrom, Name, As, TypeColumn, ...}`. `QlFrom` = model + table alias.
- **Aliases:** `QlFroms.add` hands out table aliases `A`, `B`, … (starting at ASCII 65); `QlFroms.getField(name)` sets `Field.Model.As`; the driver's `asField` then renders `alias.column`. Every `Ql` *and* `Command` has an alias-`A` from.
- **Source models** (`DefineSourceField()`, jsonb column named `jdb.SOURCE`): attributes declared with `DefineAtribute` live inside the jsonb and render as `COALESCE(A._data#>>'{name}', default)`; SELECTs render `A._data || jsonb_build_object(...) AS result`. Undeclared attribute names are **not** auto-created when resolving a where/select (`getField(name, false)`).
- **`Ql` and `Command` are mutable builders**, not immutable: `Where/And/Or/Select/Join/...` mutate the receiver and return the same pointer, and execution state (`Sql`, `Limit`, `tx`) is stored on it. `Model.Where`/`Model.Select`/`Model.Join` create a fresh `Ql` each call; do not reuse a `Ql`/`Command` after executing it.

### Fluent API

```go
items, err := jdb.From("schema.table").Where("active").Eq(true).Rows(10)   // From accepts *Model or a name
items, err := userModel.Where("active").Eq(true).All()
item, err  := userModel.Insert(data).Exec()          // Bulk(dataSlice) = Insert with N rows
items, err := userModel.Update(data).Where("id").Eq(id).Exec()
items, err := userModel.Delete("id").Eq(id).Exec()
items, err := userModel.Upsert(data).Exec()
```

`Where(field)` opens a condition; the next operator call (`Eq/Neg/In/NotIn/Like/More/Less/MoreEq/LessEq/Between/IsNull/NotNull`) fills the *last* condition. `Where`/`And`/`Or` resolve a plain string to a `*Field` via `resolveWhereField`; a string that doesn't resolve is passed through and the driver renders it as a **string literal** (`'age'`), not a column. There is no `Limit(n)`/`Offset(n)` method (`Limit`/`Offset`/`Sheet` are struct fields): use `Rows(n)` (limit n, honoring `Page(p)`), `List(page, rows)` (adds a count, returns `et.List`), `All()`, `One()`. **`First(n)` returns the single item at index `n`, not "the first n rows"** (`Last(n)` counts from the end). Transactional variants: `AllTx/OneTx/...`, `ExecTx(tx)`. `Ql.Join(with, field, operator, value)` accepts `*Model` or `string` for `with` and `Operator` or a string (`"="`, `"eq"`, …) for `operator`; `LeftJoin/RightJoin/FullJoin` and `Model.Join` still take `string` + `Operator` only.

### Command execution and hooks

`Command.ExecTx` → `inserted()` / `updated(current)` / `deleted(current)` / `upsert()` (in `command-*.go`). Update/Delete first run a SELECT (`getCurrent`) to load `old` rows. `NewCommand` registers default trigger functions (`command-trigger.go`) that fill `IndexField`, `SystemKeyField`, `CreatedAt`/`UpdatedAt`. Hooks exist at two levels — `Model.BeforeInsert…/AfterInsert…` (`model-before.go`, `model-after.go`) and per-command `Command.BeforeInsert…` — each as plain (`DataFunctionTx`) and `*Trigger` (`old, new`) variants. `command-delete.go`'s `deleted()` is the reference-correct ordering (model before → command before → write → model after → command after); insert and update deviate (see defects). `Tx` (`tx.go`) begins lazily in `queryTx`, so `NewTx()` needs no explicit `Begin`.

### Driver registration and connecting

Each driver registers in its `init()` via `jdb.Register(name, factory, defaultParams)`; activate with a blank import (`_ "github.com/celsiainternet/jdb/drivers/postgres"`). Defaults come from env vars.

```go
db, err := jdb.Load()        // driver from DB_DRIVER (default "postgres"), params from env
db, err := jdb.ConnectTo(jdb.ConnectParams{Driver: "postgres", Name: "myapp", Params: &postgres.Connection{...}})
```

`ConnectParams` is `{Id, Driver, HostName, Name, UseCore, IsDebug, Params}` — `UseCore` was restored (uncommitted `jdb/connected.go` diff) after being dropped in 9c4cb53; `NodeId` is still gone (only `DB.NodeId` exists, and nothing ever sets it). The `ConnectTo` bug that made restoring it pointless — `if !result.UseCore` tested the *freshly-built, always-zero* `result` instead of the caller's `connection` — is also fixed (uncommitted `jdb/jdb.go` diff) to `if !connection.UseCore`. So `jdb.ConnectTo(jdb.ConnectParams{UseCore: true, ...})` now does reach `createCore()`, which sets the package-level `coreSchema`/`coreModel`/`coreSeries` — so `GetSeries`/`SetSeries` no longer fail with `"Database not concurrent"` this way (they still hit the separate Upsert-emits-no-SQL defect below). But two things still make "core" mostly non-functional:
  - **`DB.UseCore`** — the field `Model.Save`/`DB.Save`/`DB.Load` actually gate on (`!s.UseCore`) — is **never assigned** anywhere (not in `NewDatabase`, not in `Conected`, not in `createCore`), so those three stay no-ops even with `UseCore: true`.
  - **`jdb.Load()` (the env-var path) still never enables core.** `postgres.Connection` also grew a `UseCore` field (env `USE_CORE`, default `true`, uncommitted `drivers/postgres/driver.go` diff), but it's dead: `Chain`/`ToJson`/`Load`/`Validate` never read it, and it isn't what `ConnectTo` checks — that's the separate top-level `ConnectParams.UseCore`, which the postgres driver's `init()` → `jdb.Register(...)` call never sets (only `Params.UseCore` inside the nested `Connection`). So `jdb.Load()` builds a `ConnectParams` with `UseCore` still `false`; only an explicit `jdb.ConnectTo(ConnectParams{UseCore: true})` triggers `createCore()`.

### HTTP handlers

`jdb/jdb.go` exposes four `http.HandlerFunc`s for any Chi/std router: `ModelDefine`, `ModelQuery` (runs a `Ql` from a JSON body via `Ql.Query`), `ModelCommand`, `ModelDescribe`.

### Package-level singleton pattern (`instances`, `authorization`, `config`, `inbox`)

Each holds an unexported package-level singleton (`inst`, `auth`, `cfg`, `inb`) with a `*jdb.Schema` + `*jdb.Model`, populated once by a `Load(db, schema, ...)` that no-ops if already loaded, built by an idempotent `Define(...)`. Each owns exactly one table and adds CRUD/HTTP/event helpers plus package-level wrapper funcs in `handler.go`. Signatures differ: `instances.Load(db, schema, name) (*Instance, error)`; the others are `Load(db, schema) error`. `authorization.Load` also calls `elvis/middleware.SetAuthorizationStore`. When adding a similar package, follow this shape.

## Known defects at HEAD

Verified 2026-09-22 on `27a3772` + the uncommitted `ql-join.go`/`where.go`/`connected.go`/`jdb.go`/`drivers/postgres/driver.go` diff, by reading code and, where marked (SQL), by generating SQL as described above. **Re-verify before relying on or "fixing" any of these** — the repo changes daily. Full write-up and fix order: the `jdb-querybuilder-core-analysis` memory.

- **`Select("a","b")` renders `SELECT *`** (SQL). `sqlSelect` (`drivers/postgres/query.go`) only fills its local `selects` when `ql.Selects` is empty, so explicit selects are dropped; with `GroupBy` this yields `SELECT * … GROUP BY`.
- **`Update`/`Delete` with `Where` are invalid SQL** (SQL): `UPDATE public.users … WHERE A.id='u1'` — the table is not aliased in UPDATE/DELETE but `asField` always prefixes the alias.
- **`Upsert` produces no SQL** (SQL): `Command.upsert()` leaves `Command == Upsert` and the postgres `Command()` switch only handles Insert/Update/Delete. Affects `instances.Set`, `config.Set`, `inbox.UpsertInboxes`, `SetSeries/GetSeries`. (Effect on a live Postgres — empty query — is inferred, not run.) Its update path would also have an empty WHERE, since `getCurrent` puts the PK conditions on a temporary `Ql`, not on the command.
- **`Ql.Join("name", …)` is silently ignored** (SQL) when `name` is a string not already in `Froms` — the uncommitted change dropped the `s.GetModel(withName)` fallback. With a `*Model` it works.
- **Join RHS value is a string literal** (SQL): `ON A.id='user_id'`, never resolved to a field. `Ql.Having(x)` likewise doesn't resolve fields (`HAVING 'age'>3`).
- **JSON query path is broken** (SQL; affects `Ql.Query`, `Model.Query`, `ModelQuery`, `DB.JQuery`): `setHavings` calls `setWhere` instead of `setWheres`, so every JSON query emits `HAVING '{}'`; and `QlWhere.setWheres` compares `strings.ToLower(key) == "AND"` (never true), so `AND`/`OR` groups render as literals.
- **Hook wiring typos** (`command-insert.go`, `command-update.go`): `inserted()` runs `s.afterInsert*` *before* the write where `model.beforeInsert*` belongs (so `Model.BeforeInsert*` never fires) and runs the after-hooks again post-write; `updated()` iterates `s.afterUpdate` pre-write where `s.beforeUpdate` belongs, so per-command `BeforeUpdate` never fires (e.g. `GetSeries`'s `value + 1`).
- **`ExecTx(nil)` commits on hook failure** (`command.go`): inner `err :=` shadows the outer `err`, so the deferred `Commit` runs after a hook error and its own error is discarded. `Bulk` keeps only the last row in `Result`.
- **`DB_RECORD_LIMIT` cap is dead code** (`From()` checks `Limit` before it is set); fluent queries default to `Limit == 0` (no `LIMIT`).
- **Core is still effectively disabled, now for a different reason.** `ConnectParams.UseCore` is back and `ConnectTo`'s check is fixed (see "Driver registration and connecting" above), so `createCore()` *is* reachable — but `DB.UseCore` (what `Model.Save`/`DB.Save`/`DB.Load` gate on) is never set anywhere, and `jdb.Load()`'s env path never sets `ConnectParams.UseCore` at all (the new `postgres.Connection.UseCore`/`USE_CORE` env var is dead, unread by anything). Only `jdb.ConnectTo(ConnectParams{UseCore: true})` reaches `createCore()`, and even then `Save`/`Load` stay no-ops.
- Panics: `jdb.GetSchema("noDot")` indexes `list[1]`; `From(<unknown name>)` dereferences a nil model; `Ql.Detail(<unknown field>)` dereferences a nil field.

## Doc drift

`README.md` (Spanish, user-facing) is out of sync with the code: it lists MySQL/SQLite drivers and env values for them, shows `ConnectParams{UserCore: true, NodeId: 1}` — still wrong even after the `UseCore` restoration: the real field is spelled `UseCore` (no "r"), `NodeId` still doesn't exist on `ConnectParams` — and shows `user.Join(profile, "id", "=", "user_id")` (`Model.Join` takes a `string` and an `Operator`; that line does not compile — only `Ql.Join` gained the flexible `interface{}`/string-operator signature), and presents `First(10)` as pagination (it returns the item at index 10). Do not copy README examples without checking them against the code. `version.sh` rewrites the README version badge and the `elvis` version in its install snippet.

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `DB_NAME` | `jdb` | Database name |
| `DB_DRIVER` | `postgres` (in `jdb.Load`/`LoadTo`) | Driver to load (`postgres` is the only working one) |
| `DB_HOST` | `localhost` | DB host |
| `DB_PORT` | `5432` | DB port |
| `DB_USER` | `admin` | DB username |
| `DB_PASSWORD` | `admin` | DB password |
| `APP_NAME` | `jdb` | Application name (used in PG connection string) |
| `DEBUG` | `false` | Enable debug logging (prints generated SQL) |
| `DB_VERSION` | `13` | PostgreSQL server version |
| `DB_RECORD_LIMIT` | `1000` | Intended max rows for `From()` (currently ineffective, see defects) |
| `USE_CORE` | `true` | Read into `postgres.Connection.UseCore`, but **dead** — nothing consumes that field; it does not enable the `core` schema (see defects) |
| `ORA_DB_HOST`, `ORA_DB_PORT`, `ORA_DB_USER`, `ORA_DB_PASSWORD`, `ORA_DB_SERVICE_NAME_ORACLE`, `ORA_DB_SSL_ORACLE`, `ORA_DB_SSL_VERIFY_ORACLE`, `ORA_VERSION` | — | Oracle stub connection only |

(`NODE_ID` was listed previously; nothing in the code reads it any more.)
