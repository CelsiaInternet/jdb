# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

`jdb` (`github.com/celsiainternet/jdb`) is a Go database library: a declarative model layer plus a fluent query/command builder that a driver translates to SQL. It is a library, not an application. `cmd/` has no top-level `main.go`; it holds three independent tool binaries (see below), none of which is "the product."

**Drivers: PostgreSQL, SQLite and Oracle (19c+) are implemented; there is no `drivers/mysql`** (`MysqlDriver` is only a name constant in `jdb/drivers.go`). SQLite (`modernc.org/sqlite`, pure Go, no CGO) and Oracle (`sijms/go-ora`) mirror the postgres driver's behavior, including dynamic attributes (undeclared `New` keys stored in the source JSON unless the model is `Integrity`). Dialect differences that change behavior:
- **No schemas**: both flatten `schema.table` to a table named `schema_table` (Oracle: inside the connected user); index/constraint names are prefixed with the schema. All identifiers are double-quoted (Oracle rejects unquoted `_id`/`_data`).
- **Source JSON**: SQLite stores TEXT and updates attributes with `json_set` (top-level replace, same as postgres `||`). Oracle stores `CLOB CHECK (IS JSON)` and updates with `JSON_MERGEPATCH` (19c has no `JSON_TRANSFORM`): a `null` value **deletes** the key and nested objects are **merged**, not replaced.
- **Results**: both normalize the `result` column (returned as text) into objects and booleans (`0/1`) into `bool`. Oracle has no multi-row `RETURNING`: INSERT uses `RETURNING ROWID INTO`, UPDATE/DELETE first lock the target rows (`SELECT ROWID … FOR UPDATE`), then write and read by ROWID.
- **Safety**: both refuse UPDATE/DELETE with an empty WHERE (`MSG_WHERE_REQUIRED`) and alias the target table. Values in INSERT/UPDATE are bound args (`?N` / `:N`); WHERE values are still inlined like postgres.
- **Oracle specifics**: `''` is `NULL` (text columns get no DEFAULT); `LIKE` compares `UPPER()`; `Search` is `LIKE '%v%'` (no Oracle Text), same in SQLite; no `ON UPDATE CASCADE`; CLOB columns get no B-tree index; DDL auto-commits, so `MutateModel` is not atomic and `DROP … CASCADE CONSTRAINTS` drops foreign keys from other tables (as postgres `DROP … CASCADE` does).

`README.md` still advertises MySQL — it is stale (see "Doc drift").

Go version: **1.23.0** (see `.go-version`). Primary dependency: `github.com/celsiainternet/elvis` (utility, logging, JSON types `et.Json`, events, cache, middleware).

## Workspace (which `elvis` and which consumers)

`go env GOWORK` resolves to **`/Users/cgalvisleon/Projects/internet/go.work`** (`/Users/cesargalvisleon/...` on the old Mac) (two levels above this repo, not `library/`). It `use`s `./library/elvis`, `./library/jdb` and ~15 services, so:
- `elvis` is built from the **local `library/elvis` checkout**, not the `v1.1.321` pinned in `go.mod`. Edits in either repo are live for the other with no tag/replace.
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
- `cmd/install` — run **from a consuming project** (mirrors `elvis/cmd/install`): stops if there is no `go.mod`; runs `go mod tidy` only when the module already has Go packages (in an empty project tidy would drop the unused jdb requirement); calls `agentsguide.Install()` even if tidy failed; prints a hint to run `elvis/cmd/install` when `.claude/guides/elvis/` is missing. Exit code 1 if tidy or the guide failed. It no longer `go get`s a fixed dependency list (that pinned old versions and could downgrade).
- `cmd/create` — a Cobra CLI (`cmd/create/v2`, a fork of `elvis/create/v1` adapted to jdb) with four options: **Project** (`MkProject`: microservice + README/.env/.gitignore + `CLAUDE.md`/`AGENTS.md` rules from `template/agents.md` (marker `<!-- jdb:project-rules -->`) + `.claude/settings.json` denying `git commit`/`git push` + the jdb **and** elvis agent guides, then `go mod tidy` last), **Microservice**, **Modelo** (`MakeModel`: refuses if `pkg/<pkg>` is missing, or if it was created without schema; otherwise writes `internal/models/<pkg>/<model>.go` + `schema_<schema>.go`, appends `models.Define<Model>(db)` to `initModels` and inserts the routes before the `ctx := context.Background()` anchor in `router.go`) and **Rpc**. Generated layout: `cmd/<name>`, `internal/services/<name>` (HTTP server, `v1/api.go` loads config → cache → event → `StartRpcServer` → `jdb.Load()` with the postgres driver), `internal/models/<name>` (jdb models on `DefineProjectModel`: Get/Insert/Upsert/State/Delete/All/Query via `RunQuery`), `pkg/<name>` (router, `router-<model>.go` handlers + RPC method, controller, events, config). Templates are Go string constants in `cmd/create/v2/model.go` with `$1`/`$2` positional substitution (`file.MakeFile`, which never overwrites). Every project needs a `go.mod` first (module path comes from it). `go test ./cmd/create/...` checks every Go template is valid gofmt-clean Go, routes match handlers, and no template carries a JWT. Verified 2026-11-04 by generating projects (with/without schema, extra models in two schemas, schemaless model, rpc): build + vet clean in workspace mode and against published elvis v1.1.321 / jdb v1.0.112; the generated models ran end-to-end on SQLite. On postgres the generated update/state/delete work (the `Update … Where` alias defect was not reproduced on 2026-11-04).
- **Release blocker:** `agentsguide/guide.go` calls `elvisguide.InstallGuide`, which only exists in the local `elvis` checkout (not in the pinned/published `v1.1.321`), so `GOWORK=off go build ./...` fails and `go run github.com/celsiainternet/jdb/cmd/install|create` cannot work for consumers until elvis is tagged and `go.mod` bumped.

Tests: `go test ./drivers/...` (all pass at `ac5d257`; single test: `go test ./drivers/sqlite -run TestName -v`). `drivers/postgres/command_test.go` only checks generated SQL; `drivers/sqlite` runs end-to-end against an in-memory database; `drivers/oracle` runs end-to-end only when `ORA_TEST_HOST` (+ `ORA_TEST_PORT`, `ORA_TEST_USER`, `ORA_TEST_PASSWORD`, `ORA_TEST_SERVICE`) points at a server, otherwise it skips — e.g. `docker run -d -p 15210:1521 -e ORACLE_PASSWORD=x -e APP_USER=jdb -e APP_USER_PASSWORD=jdbtest gvenzl/oracle-free:slim`, service `FREEPDB1`. There are no tests for the `jdb` package itself. Commits on `main` are auto-generated "Backup:" snapshots, so `git log` says little; use `git log -L` / `git show` on the file to find when behavior changed.

### Checking generated SQL without a database

No DB is needed to see what SQL the driver builds: the driver stores it in `Ql.Sql` / `Command.Sql` *before* it executes, and with no connection `db.Db` is nil, so execution panics afterwards. Recipe used to verify this file:
1. Make a throwaway module in the scratchpad (not the repo) with a `go.work` that `use`s `.`, `library/elvis` and `library/jdb`; run with `GOWORK=<that file> go run .` (`-mod=mod` is not allowed in workspace mode).
2. `db, _ := jdb.NewDatabase("x", "postgres")` (blank-import `drivers/postgres`), `jdb.NewSchema`, `jdb.NewModel`, `DefineColumn`… — **skip `model.Init()`** (it hits the DB).
3. Wrap execution in `defer func(){ recover() }()` and print `q.Sql`. For commands, set `c.New = <data>` and call `db.Command(c)` directly — `c.Exec()` first runs a SELECT (`getCurrent`) and panics before reaching the write.

## Architecture

### Package layout

```
jdb/            Core library package (declarative intent: Model/Ql/Command/Field/Where)
drivers/        postgres, sqlite (modernc), oracle (go-ora, 19c+)
instances/      Feature singleton: generic keyed blob store
authorization/  Feature singleton: auth/session model (wires into elvis/middleware)
config/         Feature singleton: runtime config model
inbox/          Feature singleton: message inbox model
cmd/            Standalone tool binaries (test/install/create) — not the library
agentsguide/    JDB.md (agent guide: patterns + known defects) installed into consumers via elvis agentsguide.InstallGuide
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

### Rollups

`DefineRollup(name, with, fks, fields)` (`Show` attribs) / `DefineObject(...)` (`Show` objects) add a virtual `TpRollup` column. A query resolves it only when requested (`Select`/`Detail` with its name): `Ql.getDetailsTx` → `rollupValueTx` (`ql-query.go`) queries `With` per row by `Fk` (local field -> field in `With`, sorted, AND), plus the fixed `Rollup.Wheres` (`SetWheres(et.Json{"key": et.Json{"eq": "CC"}})`, validated operators), takes `Fields` from the row (explicit selects are ignored by the driver) and sets `item[Name]`: the value with attribs and one field, otherwise an object; missing key or no match -> nil fields. Rows with the same keys share one query (cache on the `Ql`, cloned per row). Serialized as `{ name, with, fk, fields, show: "attribs"|"objects", wheres }` (`ShowRollup` reads old numeric values); `getWith` rebuilds `With` from `WithName`. Rollups defined in data (not on the shared model, to avoid concurrent map writes while serving queries): `NewRollup(name, with, fks, fields, show)` + `StrToShowRollup`, passed per query with `Ql.AddRollups(...)` or `RunQuery(model, query, rollups...)`. Tests: `drivers/sqlite/rollup_test.go`.

### Fluent API

```go
items, err := jdb.From("schema.table").Where("active").Eq(true).Rows(10)   // From accepts *Model or a name
items, err := userModel.Where("active").Eq(true).All()
item, err  := userModel.Insert(data).Exec()          // Bulk(dataSlice) = Insert with N rows
items, err := userModel.Update(data).Where("id").Eq(id).Exec()
items, err := userModel.Delete("id").Eq(id).Exec()
items, err := userModel.Upsert(data).Exec()
```

`Where(field)` opens a condition; the next operator call (`Eq/Neg/In/NotIn/Like/More/Less/MoreEq/LessEq/Between/IsNull/NotNull`) fills the *last* condition. `Where`/`And`/`Or` resolve a plain string to a `*Field` via `resolveWhereField`; a string that doesn't resolve is passed through and the driver renders it as a **string literal** (`'age'`), not a column. JSON paths: when the name is not a declared field, `QlFroms.getField` tries `Model.getPathField` (`field-path.go`), which accepts Postgres arrows (`_data->'sync_external'->>'oss_order_id'`) or dots (`_data.sync_external.oss_order_id`, or `sync_external.oss_order_id` inside the source). The first segment must be a `TpColumn` with `TypeDataObject`; otherwise the whole path goes into `SourceField`. Keys must match `[A-Za-z0-9_-]+`, or the name is left unresolved (literal). The result is a `Field` with `Path`, used by where, order by, select and returning (not insert/update data), and rendered by `asField`: Postgres `A._data#>>'{a,b}'` (text, with `pathCast` adding `::numeric`/`::boolean` when the compared value, or the first item of In/Between, is a number or bool), SQLite `->> '$."a"."b"'` (typed) and Oracle `JSON_VALUE(...)`. Tests: `drivers/sqlite/json_path_test.go` and `drivers/postgres/json_path_test.go`. There is no `Limit(n)`/`Offset(n)` method (`Limit`/`Offset`/`Sheet` are struct fields): use `Rows(n)` (limit n, honoring `Page(p)`), `List(page, rows)` (adds a count, returns `et.List`), `All()`, `One()`. **`First(n)` returns the single item at index `n`, not "the first n rows"** (`Last(n)` counts from the end). Transactional variants: `AllTx/OneTx/...`, `ExecTx(tx)`. `Ql.Join(with, field, operator, value)` accepts `*Model` or `string` for `with` and `Operator` or a string (`"="`, `"eq"`, …) for `operator`; `LeftJoin/RightJoin/FullJoin` and `Model.Join` still take `string` + `Operator` only.

### Command execution and hooks

`Command.ExecTx` → `inserted()` / `updated(current)` / `deleted(current)` / `upsert()` (in `command-*.go`). Update/Delete first run a SELECT (`getCurrent`) to load `old` rows. `NewCommand` registers default trigger functions (`command-trigger.go`) that fill `IndexField`, `SystemKeyField`, `CreatedAt`/`UpdatedAt`. Hooks exist at two levels — `Model.BeforeInsert…/AfterInsert…` (`model-before.go`, `model-after.go`) and per-command `Command.BeforeInsert…` — each as plain (`DataFunctionTx`) and `*Trigger` (`old, new`) variants. `command-delete.go`'s `deleted()` is the reference-correct ordering (model before → command before → write → model after → command after); insert and update deviate (see defects). `Tx` (`tx.go`) has no public constructor (`newTx` is unexported): a command or query begins one lazily in `queryTx`, hooks receive it as `tx *jdb.Tx`, and callers propagate it with `ExecTx(tx)`/`AllTx(tx)`/`OneTx(tx)`.

### Driver registration and connecting

Each driver registers in its `init()` via `jdb.Register(name, factory, defaultParams)`; activate with a blank import (`_ "github.com/celsiainternet/jdb/drivers/postgres"`). Defaults come from env vars.

```go
db, err := jdb.Load()        // driver from DB_DRIVER (default "postgres"), params from env
db, err := jdb.ConnectTo(jdb.ConnectParams{Driver: "postgres", Name: "myapp", Params: &postgres.Connection{...}})
```

`ConnectParams` is `{Id, Driver, HostName, Name, UseCore, IsDebug, Params}` (`NodeId` is gone; only `DB.NodeId` exists, and nothing sets it). **Core schema** (re-verified 2026-11-04 on `ac5d257`): all three drivers' `init()` now set the top-level `ConnectParams.UseCore` from `USE_CORE` (default `true`), `load()` (`jdb/connected.go`) copies it, and `ConnectTo` checks `connection.UseCore`. So **`jdb.Load()`/`LoadTo()` call `createCore()` by default**, which defines the `core` schema models (`coreSchema`/`coreModel`/`coreSeries`, `jdb/core.go`) and so touches the DB. Set `USE_CORE=false` to skip it. `GetSeries`/`SetSeries` therefore get past `"Database not concurrent"` but still hit the Upsert-emits-no-SQL defect below. **`DB.UseCore`**, the field that `Model.Save`/`DB.Save`/`DB.Load` gate on, is still **never assigned**. Schemas and models copy it from the DB, so it stays `false` and those three remain no-ops.

### JSON queries and commands (JSQL)

[jsql.md](jsql.md) documents every JSON format with full examples, generated SQL and responses (verified 2026-11-04 on SQLite in memory + postgres SQL generation). Read it before touching these paths. There are two query dialects with different shapes:

- **`RunQuery(model, query, rollups...)`** (`jdb/jsql.go`) — the one consumers use (`octopus/api/internal/models/vcore/query.go`). `wheres` is an **ordered list** (`{campo:{op:v}}`, `{"and":{...}}`, `{"or":{...}}`, one condition per element), plus `selects`, `order_by {asc,desc}` (validated: unknown column → `MSG_QUERY_COLUMN_NOT_VALID`), `page`, `rows` (alias `limit`, default `QUERY_DEFAULT_LIMIT` 100, max `QUERY_MAX_LIMIT` 1000), and `join`/`left_join`/… Returns `et.List` (`all`, `count`, `page`, `start`, `end`, `result`).
- **`Ql.Query` / `Model.Query` / `DB.JQuery` / `ModelQuery`** (`ql-query.go` `queryTx`) — `where` is an **object** (plain keys + `and`/`or` arrays), plus `select`, `join`, `group_by`, `having`, `order_by`, `page`, `limit`, `debug`. Returns `{ok,count,result}` (or one object with `limit: 1`).

Operator names (both dialects, case-sensitive): `eq`, `neg`, `more`, `less`, `moreEq`, `lessEq`, `like`, `in`, `notIn`, `between`, `isNull`, `notNull`; `StrToOperator` also maps `=`, `!=`, `>`, `<`, `>=`, `<=`, `search`.

### HTTP handlers

`jdb/jdb.go` exposes `http.HandlerFunc`s for any Chi/std router: `ModelDefine` (body `{params:{type,name}}`), `ModelQuery` (runs a `Ql` from a JSON body via `Ql.Query`, model from `from`), `ModelCommand` (body `{"schema.model": {insert|bulk|update|delete|upsert, where}}`), `ModelDescribe` (body `{kind,name}`). `JSQL` is a stub that echoes the body.

### Package-level singleton pattern (`instances`, `authorization`, `config`, `inbox`)

Each holds an unexported package-level singleton (`inst`, `auth`, `cfg`, `inb`) with a `*jdb.Schema` + `*jdb.Model`, populated once by a `Load(db, schema, ...)` that no-ops if already loaded, built by an idempotent `Define(...)`. Each owns exactly one table and adds CRUD/HTTP/event helpers plus package-level wrapper funcs in `handler.go`. Signatures differ: `instances.Load(db, schema, name) (*Instance, error)`; the others are `Load(db, schema) error`. `authorization.Load` also calls `elvis/middleware.SetAuthorizationStore`. When adding a similar package, follow this shape.

## Known defects at HEAD

Verified 2026-09-22 (core items re-verified 2026-11-04 on `ac5d257`, now committed), by reading code and, where marked (SQL), by generating SQL as described above. **Re-verify before relying on or "fixing" any of these** — the repo changes daily. Full write-up and fix order: the `jdb-querybuilder-core-analysis` memory.

- ~~`Select("a","b")` renders `SELECT *`~~ — **no longer true** (re-verified 2026-11-04, SQL): postgres renders `jsonb_build_object('name', A.name, 'age', A.age) AS result`, also with `GROUP BY`; SQLite uses `json_object(...)`.
- ~~postgres `quote` did not escape JSON/array literals~~ — **fixed 2026-11-05** (`drivers/postgres/quote.go`): `et.Json`, `map[string]interface{}` and slices were inlined as `'<json>'` without doubling `'`, so a value with an apostrophe (e.g. a script `Et library's`) broke the INSERT/UPDATE (`syntax error at or near "s"`) and was an SQL-injection path. They now go through `quote(string)` like sqlite/oracle; `quote_test.go` covers it. `JsonQuote`-style formatters further down the file still use `fmt.Sprintf` without escaping.
- ~~`Update`/`Delete` with `Where` are invalid SQL~~ — **not reproduced 2026-11-04** against a real postgres: fluent `Model.Update(data).Where(k).Eq(v).And(...)` and `Model.Delete(k).Eq(v)` emit `UPDATE schema.table SET … WHERE _id='a' AND …` / `DELETE FROM … WHERE …` without alias and change the row. WHERE values are still inlined (quoted), not bound.
- ~~`Upsert` produces no SQL~~ — **fixed** (since `e310a29`, 2026-09-22): `Command.upsert()` now switches to `Insert` or `Update` (with the PK conditions from `getCurrent`) before executing. 
- **`Ql.Join("name", …)` is silently ignored** (SQL) when `name` is a string not already in `Froms` — the `s.GetModel(withName)` fallback was dropped. With a `*Model` it works.
- **Join RHS value is a string literal** (SQL): `ON A.id='user_id'`, never resolved to a field. `Ql.Having(x)` likewise doesn't resolve fields (`HAVING 'age'>3`).
- **JSON query path is partly broken** (SQL; affects `Ql.Query`, `Model.Query`, `ModelQuery`, `DB.JQuery`): `setHavings` calls `setWhere` instead of `setWheres`, so a non-empty `having` emits `HAVING '{}'`. The `and`/`or` groups of `QlWhere.setWheres` were fixed (2026-10-23: case-insensitive key, values read from each condition, and `in`/`notIn` expand JSON lists); an `or` group is not parenthesized, so mixing it with other conditions follows SQL precedence. Also `setLimitTx` with `limit > 1` calls `FirstTx(tx, limit)` and returns the row at that index: pass `limit: 0` for all rows.
  Re-verified 2026-11-04 by running it (SQLite in memory, see [jsql.md](jsql.md)): `limit > 1` — including the default 1000 when `limit` is omitted — **panics** (`index out of range [N] with length N`, `FirstTx` only guards `n > Count`) when at least N rows match, and returns empty otherwise. `page` with `limit: 0` emits `LIMIT 0 OFFSET 0` (no rows). Because `where` is a Go map, mixing an `or` group with other keys yields a **different SQL per run** (when the `or` condition lands first its connector is dropped and the `OR` becomes `AND`). `having` renders `HAVING '{"kind":{"eq":"lead"}}'`. `join` (`setJoins`) is ignored per the `Ql.Join` string defect above.
- **`RunQuery` (`jsql.go` `queryWhere`) silently drops conditions** (verified by running): a `wheres` key that is not a model field (`model.GetField` nil) is skipped, so a typo returns **all rows**; JSON paths (`_data.x`) are skipped the same way. Its operator switch has no `between`/`notIn`, but `ql.Where(column)` was already opened: `between` returns all rows and `notIn` returns none. `join` keys are ignored (same `Ql.Join` string defect). Fixed 2026-11-05 (`drivers/sqlite/concat_test.go` `TestRunQueryGroupedConditions`): an unknown key inside `{and}`/`{or}` no longer applies its operators to the previous condition (it turned `team = 't2'` into `team LIKE …`), and `{concat: {fields, <op>}}` (`Ql.WhereConcat`, one condition over several fields joined by a space) works at the top level and inside `and`/`or`.
- **`ModelCommand` `update`/`delete` are no-ops** (verified by running): `Command.setWheres` (`command-where.go`) does not call `resolveWhereField` (unlike `Ql.setWheres`), so the keys reach the SQL as string literals (`'_id'='c9'`): HTTP 200 with `{ok:false,count:0}` and nothing changes. `upsert` with an existing key goes to INSERT and fails with a UNIQUE violation. Only `insert` and `bulk` work. Fluent `Command.Where/And/Or` do resolve fields.
- **Hook wiring typos** (`command-insert.go`, `command-update.go`): `inserted()` runs `s.afterInsert*` *before* the write where `model.beforeInsert*` belongs (so `Model.BeforeInsert*` never fires) and runs the after-hooks again post-write; `updated()` iterates `s.afterUpdate` pre-write where `s.beforeUpdate` belongs, so per-command `BeforeUpdate` never fires (e.g. `GetSeries`'s `value + 1`).
- **`ExecTx(nil)` commits on hook failure** (`command.go`): inner `err :=` shadows the outer `err`, so the deferred `Commit` runs after a hook error and its own error is discarded. `Bulk` keeps only the last row in `Result`.
- **SQLite: `GetSeries` does not increment** (verified 2026-11-04 by running): on the SQLite driver the update branch of `GetSeries` (`value = CALC("value + 1")`) leaves `value` as is, so every call returns the same code (a series created by `SetSeries(…, 0)` keeps returning `…00000000`, and a new one returns 1 and then 0). On postgres it increments (1, 2, 3). Tests on SQLite must not expect consecutive codes.
- **`DB_RECORD_LIMIT` cap is dead code** (`From()` checks `Limit` before it is set); fluent queries default to `Limit == 0` (no `LIMIT`).
- **Core is only half-enabled.** `createCore()` runs by default (`USE_CORE=true`, see "Driver registration and connecting"), but `DB.UseCore` is never set, so `Model.Save`/`DB.Save`/`DB.Load` are no-ops.
- Panics: `jdb.GetSchema("noDot")` indexes `list[1]`; `From(<unknown name>)` dereferences a nil model; `Ql.Detail(<unknown field>)` dereferences a nil field.

## Doc drift

`README.md` (Spanish, user-facing) is out of sync with the code: it lists a MySQL driver and env values for it, shows `ConnectParams{UserCore: true, NodeId: 1}` — still wrong even after the `UseCore` restoration: the real field is spelled `UseCore` (no "r"), `NodeId` still doesn't exist on `ConnectParams` — and shows `user.Join(profile, "id", "=", "user_id")` (`Model.Join` takes a `string` and an `Operator`; that line does not compile — only `Ql.Join` gained the flexible `interface{}`/string-operator signature), and presents `First(10)` as pagination (it returns the item at index 10). Do not copy README examples without checking them against the code. `version.sh` rewrites the README version badge and the `elvis` version in its install snippet.

## Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `DB_NAME` | `jdb` | Database name |
| `DB_DRIVER` | `postgres` (in `jdb.Load`/`LoadTo`) | Driver to load: `postgres`, `sqlite` or `oracle` (blank-import the driver package) |
| `DB_HOST` | `localhost` | DB host |
| `DB_PORT` | `5432` | DB port |
| `DB_USER` | `admin` | DB username |
| `DB_PASSWORD` | `admin` | DB password |
| `APP_NAME` | `jdb` | Application name (used in PG connection string) |
| `DEBUG` | `false` | Enable debug logging (prints generated SQL) |
| `DB_VERSION` | `13` | PostgreSQL server version |
| `DB_RECORD_LIMIT` | `1000` | Intended max rows for `From()` (currently ineffective, see defects) |
| `USE_CORE` | `true` | Sets `ConnectParams.UseCore` in every driver's `init()`: `jdb.Load()` creates the `core` schema models. `false` skips it (`Save`/`Load` stay no-ops either way, see defects) |
| `SQLITE_PATH` | `<DB_NAME>.db` | SQLite file path (`:memory:` for an in-memory database) |
| `ORA_DB_HOST`, `ORA_DB_PORT` (`1521`), `ORA_DB_USER`, `ORA_DB_PASSWORD`, `ORA_DB_SERVICE_NAME_ORACLE`, `ORA_DB_SSL_ORACLE`, `ORA_DB_SSL_VERIFY_ORACLE`, `ORA_VERSION` (`19`) | — | Oracle connection |

(`NODE_ID` was listed previously; nothing in the code reads it any more.)
