# JDB - Go Database Library

[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/Version-v1.0.101-orange.svg)](https://github.com/celsiainternet/jdb/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-celsiainternet%2Fjdb-black.svg)](https://github.com/celsiainternet/jdb)

## Instalación

```bash
go get github.com/celsiainternet/jdb@v1.0.101
```

`jdb` depende de `elvis`, así que normalmente también necesitarás:

```bash
go get github.com/celsiainternet/elvis@v1.1.312
```

## Qué es

`jdb` es una librería de Go para trabajar con bases de datos: defines modelos de forma declarativa y construyes consultas (`Ql`) y comandos (`Command`) con una API fluida; un driver traduce esa intención a SQL.

- **Driver disponible:** solo **PostgreSQL** (`drivers/postgres`). `drivers/oracle` conecta pero sus operaciones devuelven `"not implemented"`. No hay drivers MySQL ni SQLite.
- **Dependencia:** [`elvis`](https://github.com/celsiainternet/elvis) (tipos `et.Json`/`et.Item`/`et.Items`, logging, eventos).
- **No es una aplicación:** `cmd/` contiene tres herramientas independientes (ver [Herramientas](#herramientas-de-cmd)).

## Paquetes

| Paquete | Import | Propósito |
|---|---|---|
| `jdb` | `github.com/celsiainternet/jdb/jdb` | Núcleo: conexión, modelos, consultas, comandos, hooks, handlers HTTP |
| `postgres` | `github.com/celsiainternet/jdb/drivers/postgres` | Driver PostgreSQL (se registra con import en blanco) |
| `instances` | `github.com/celsiainternet/jdb/instances` | Almacén genérico clave → objeto |
| `authorization` | `github.com/celsiainternet/jdb/authorization` | Permisos por proyecto/perfil/método/ruta |
| `config` | `github.com/celsiainternet/jdb/config` | Configuración por `tag` + `stage` |
| `inbox` | `github.com/celsiainternet/jdb/inbox` | Bandeja de mensajes/solicitudes |

---

## Paquete `jdb`

### Conexión y registro

| Función / método | Descripción |
|---|---|
| `Load() (*DB, error)` | Conecta usando el driver de `DB_DRIVER` (default `postgres`) y variables de entorno |
| `LoadTo(database, hostname...)` | Igual que `Load` pero con otro nombre de base de datos |
| `ConnectTo(ConnectParams) (*DB, error)` | Conecta con parámetros explícitos |
| `LoadConnectParams(et.Json)` | Construye `ConnectParams` desde JSON |
| `Register(name, factory, params)` | Registra un driver (lo usan los drivers en su `init()`) |
| `NewDatabase(name, driver)` | Crea un `*DB` sin conectar |
| `Jdb()`, `GetDB(name)`, `GetSchema("schema.x")`, `GetModel(name)` | Acceso al registro global |

`ConnectParams`: `Id`, `Driver`, `HostName`, `Name`, `UseCore`, `IsDebug`, `Params`.

```go
import (
    "github.com/celsiainternet/jdb/drivers/postgres"
    "github.com/celsiainternet/jdb/jdb"
)

db, err := jdb.ConnectTo(jdb.ConnectParams{
    Driver:  "postgres",
    Name:    "myapp",
    IsDebug: true,
    Params: &postgres.Connection{
        Host: "localhost", Port: 5432,
        Username: "postgres", Password: "password",
        Database: "myapp", App: "myapp",
    },
})
if err != nil {
    panic(err)
}
defer db.Disconected()

// o, desde variables de entorno:
// import _ "github.com/celsiainternet/jdb/drivers/postgres"
// db, err := jdb.Load()
```

### `*DB`

| Método | Descripción |
|---|---|
| `Ping()`, `HealthCheck()` | Verifica la conexión |
| `Conected(params)`, `Disconected()` | Abre / cierra la conexión |
| `GetSchema(name)`, `GetModel(name)`, `DropSchema(name)` | Schemas y modelos registrados |
| `LoadModel`, `MutateModel`, `DropModel`, `EmptyModel` | DDL sobre un modelo |
| `From(table) *Ql` | Inicia una consulta |
| `Select(ql)`, `Count(ql)`, `Exists(ql)`, `Command(cmd)` | Ejecuta un `Ql`/`Command` ya construido |
| `Query(sql, args...)`, `One(sql, args...)` | SQL crudo, devuelve `et.Items` / `et.Item` |
| `JQuery(et.Json)`, `QueryModel(et.Json)` | Consulta definida en JSON |
| `SetDebug(bool)`, `Debug()`, `Describe()` | Depuración y descripción |

```go
items, err := db.Query("SELECT * FROM public.users WHERE age > $1", 18)
```

### Schemas y modelos

| Función / método | Descripción |
|---|---|
| `NewSchema(db, name)` | Crea/obtiene un schema |
| `NewModel(schema, name, version)` | Crea/obtiene un modelo (tabla) |
| `NewTable(db, "schema.table")` | Modelo a partir de un nombre completo |
| `Model.Init()` | Crea/actualiza la tabla en la base de datos |
| `Model.Drop()`, `Model.Empty()` | Elimina / vacía la tabla |
| `Model.GetField(name)`, `Model.GetModel(name)` | Resuelve campos y modelos relacionados |
| `Model.GenId()`, `Model.GetId(id)` | Genera / normaliza identificadores |
| `Model.CheckRequired(data)` | Valida campos requeridos |
| `Model.Counted()`, `Model.Query(et.Json)` | Conteo y consulta JSON |
| `Model.On(channel, handler)`, `Model.Emit(channel, data)` | Eventos por modelo |
| `Model.Describe()`, `Model.Debug()` | Descripción y depuración |

### Definición de columnas (`Define*`)

| Método | Descripción |
|---|---|
| `DefineColumn(name, TypeData) *Column` | Columna real |
| `DefineAtribute(name, TypeData) *Column` | Atributo dentro de la columna jsonb fuente |
| `DefineSourceField()`, `DefineSource(name)` | Columna jsonb fuente (`_data`) |
| `DefinePrimaryKey(cols...)`, `DefinePrimaryKeyField()` | Llave primaria |
| `DefineForeignKey(fks, with, onDelete, onUpdate)` | Llave foránea |
| `DefineIndex(sort, cols...)`, `DefineUnique(cols...)` | Índices |
| `DefineRequired(cols...)`, `DefineHidden(cols...)` | Requeridos / ocultos en resultados |
| `DefineCreatedAtField()`, `DefineUpdatedAtField()`, `DefineStatusField()`, `DefineSystemKeyField()`, `DefineIndexField()`, `DefineProjectField()` | Campos de sistema (se rellenan automáticamente) |
| `DefineRelation(name, with, fks, limit)` | Relación uno-a-muchos |
| `DefineRollup(name, with, fks, fields)` | Agregación desde otra tabla |
| `DefineObject(name, with, fks, fields)` | Objeto embebido uno-a-uno |
| `DefineDetail(name, fks, limit)` | Detalle |
| `DefineCalc(name, fn)`, `DefineCalcTx(name, fn)` | Campo calculado en Go |
| `DefineModel()`, `DefineProjectModel()` | Conjunto estándar de campos de sistema |
| `DefineIntegrity()` | Activa integridad referencial |

Tipos de dato (`TypeData*`): `Text`, `Memo`, `ShortText`, `Key`, `Number`, `Int`, `Precision`, `DateTime`, `Checkbox`, `Bytes`, `Object`, `Select`, `MultiSelect`, `Geometry`, `FullText`, `State`, `User`, `FilesMedia`, `Url`, `Email`, `Phone`, `Address`, `Relation`, `Rollup`.

Modificadores de `*Column`: `SetDefaultValue`, `SetHidden`, `SetMax`, `SetMin`.

```go
schema := jdb.NewSchema(db, "public")

users := jdb.NewModel(schema, "users", 1)
users.DefineColumn("id", jdb.TypeDataKey)
users.DefineColumn("name", jdb.TypeDataText)
users.DefineColumn("email", jdb.TypeDataEmail)
users.DefineColumn("age", jdb.TypeDataInt).SetDefaultValue(0)
users.DefineCreatedAtField()
users.DefineUpdatedAtField()
users.DefinePrimaryKey("id")
users.DefineUnique("email")
users.DefineRequired("name", "email")

if err := users.Init(); err != nil {
    panic(err)
}
```

### Consultas (`*Ql`)

Se inicia con `jdb.From(model | "schema.table")`, `db.From(...)`, `Model.Where(...)`, `Model.Select(...)`, `Model.Data(...)` o `Model.Join(...)`. Los métodos mutan el `Ql` y lo devuelven; no reutilices un `Ql` después de ejecutarlo.

| Grupo | Métodos | Descripción |
|---|---|---|
| Condiciones | `Where`, `And`, `Or` | Abren una condición sobre un campo |
| Operadores | `Eq`, `Neg`, `In`, `NotIn`, `Like`, `More`, `Less`, `MoreEq`, `LessEq`, `Between`, `IsNull`, `NotNull` | Completan la última condición |
| Proyección | `Select`, `Data`, `Detail`, `Hidden` | Campos a devolver |
| Joins | `Join`, `LeftJoin`, `RightJoin`, `FullJoin` | Unión con otro modelo |
| Agrupación | `GroupBy`, `Having` | Agrupa y filtra grupos |
| Orden | `OrderBy`, `OrderByAsc`, `OrderByDesc` | Orden de resultados |
| Paginación | `Page(p)` + `Rows(n)`, `List(page, rows)` | `List` devuelve `et.List` con total |
| Ejecución | `All()`, `One()`, `Rows(n)`, `First(n)`, `Last(n)`, `Counted()`, `ItExists()` | `First(n)`/`Last(n)` devuelven **el ítem en la posición `n`** |
| Transacción | `AllTx`, `OneTx`, `RowsTx`, `FirstTx`, `LastTx`, `CountedTx`, `ItExistsTx` | Igual, dentro de un `*Tx` |
| JSON | `Query(et.Json)`, `QueryTx` | Ejecuta una consulta descrita en JSON |
| Otros | `Debug()`, `Describe()`, `Tx()` | Depuración |

Agregaciones: `SUM`, `COUNT`, `AVG`, `MIN`, `MAX`, `VALUE`, `CALC`, `EXTRACT_YEAR`/`MONTH`/`DAY`/`HOUR`/`MINUTE`/`SECOND`. Operadores de join aceptados como texto: `"="`, `"!="`, `"eq"`, `"neg"`, `"in"`, `"like"`, `"more"`, `"less"`, …

```go
// Todas las filas que cumplen la condición
items, err := users.Where("age").MoreEq(18).And("name").Like("%Ana%").All()

// Una sola fila
item, err := jdb.From(users).Where("email").Eq("ana@example.com").One()

// Página 2 de 20 filas, con total
list, err := users.Where("age").More(0).OrderByDesc("created_at").List(2, 20)

// Conteo y existencia
n, err := users.Where("age").Less(18).Counted()
ok, err := users.Where("email").Eq("x@y.com").ItExists()

// Join (con *Model)
items, err = jdb.From(users).Join(profiles, "id", "=", "user_id").All()
```

### Comandos (`*Command`)

| Inicio (`*Model`) | Descripción |
|---|---|
| `Insert(et.Json)` | Inserta una fila |
| `Bulk([]et.Json)` | Inserta varias filas |
| `Update(et.Json)` | Actualiza las filas que cumplen el `Where` |
| `Delete(field)` | Elimina; abre la condición sobre `field` |
| `Upsert(et.Json)` | Inserta o actualiza |

| Método | Descripción |
|---|---|
| `Where`, `And`, `Or` + operadores (`Eq`, `In`, `Like`, …) | Condiciones, igual que en `Ql` |
| `Returns(fields...)` | Campos a devolver |
| `Exec()`, `One()` | Ejecuta (abre y confirma su propia transacción) |
| `ExecTx(tx)`, `OneTx(tx)` | Ejecuta dentro de una transacción existente |
| `Debug()`, `Describe()` | Depuración |

```go
item, err := users.Insert(et.Json{"id": "u1", "name": "Ana", "email": "ana@example.com"}).One()

items, err := users.Update(et.Json{"age": 31}).Where("id").Eq("u1").Exec()

items, err = users.Delete("id").Eq("u1").Exec()

items, err = users.Bulk([]et.Json{
    {"id": "u2", "name": "Carlos", "email": "carlos@example.com"},
    {"id": "u3", "name": "María", "email": "maria@example.com"},
}).Exec()
```

### Hooks

Disponibles en `*Model` (aplican a todos los comandos) y en `*Command` (solo a ese comando).

| Método | Firma de la función |
|---|---|
| `BeforeInsert`, `BeforeUpdate`, `BeforeDelete`, `BeforeInsertOrUpdate` | `func(tx *jdb.Tx, data et.Json) error` |
| `AfterInsert`, `AfterUpdate`, `AfterDelete`, `AfterInsertOrUpdate` | `func(tx *jdb.Tx, data et.Json) error` |
| `Before…Trigger`, `After…Trigger` (mismas variantes) | `func(tx *jdb.Tx, old, new et.Json) error` |

```go
users.AfterInsert(func(tx *jdb.Tx, data et.Json) error {
    fmt.Println("nuevo usuario:", data.Str("id"))
    return nil
})

users.BeforeUpdateTrigger(func(tx *jdb.Tx, old, new et.Json) error {
    if old.Str("email") != new.Str("email") {
        return fmt.Errorf("el email no se puede cambiar")
    }
    return nil
})
```

### Transacciones

`jdb.Tx` envuelve un `*sql.Tx`. Sin transacción (`Exec()`), cada comando abre y confirma la suya.

```go
sqlTx, err := db.Db.Begin()
if err != nil {
    panic(err)
}
tx := &jdb.Tx{Tx: sqlTx}

if _, err := users.Insert(et.Json{"id": "u4", "name": "Luis", "email": "luis@example.com"}).ExecTx(tx); err != nil {
    sqlTx.Rollback()
    panic(err)
}
sqlTx.Commit()
```

### Series y utilidades

| Función | Descripción |
|---|---|
| `NewSerie(kind, tag, format, last)`, `SetSeries(...)`, `GetSeries(kind, tag)`, `ImportSeries(items)` | Consecutivos en el schema `core` (requiere `ConnectParams{UseCore: true}`) |
| `Query(db, sql, args...)`, `QueryTx(db, tx, sql, args...)` | SQL crudo |
| `RowsToItems(*sql.Rows)` | Convierte filas a `et.Items` |

### Handlers HTTP

| Handler | Descripción |
|---|---|
| `ModelDefine` | Define un modelo desde JSON |
| `ModelQuery` | Ejecuta un `Ql` descrito en el body JSON |
| `ModelCommand` | Ejecuta comandos descritos en el body JSON |
| `ModelDescribe` | Describe un objeto por tipo y nombre |

```go
r := chi.NewRouter()
r.Post("/model/query", jdb.ModelQuery)
r.Post("/model/command", jdb.ModelCommand)
r.Post("/model/define", jdb.ModelDefine)
r.Post("/model/describe", jdb.ModelDescribe)
```

---

## Paquete `drivers/postgres`

| Elemento | Descripción |
|---|---|
| `Connection{Database, Host, Port, Username, Password, App, Version, IsDebug}` | Parámetros de conexión (`Chain()`, `Validate()`, `Load(et.Json)`, `ToJson()`) |
| `Params.CreateUser`, `ChangePassword`, `DeleteUser`, `GrantPrivileges` | Gestión de usuarios de PostgreSQL |
| `Params.ExistDatabase`, `CreateDatabase`, `DropDatabase`, `DropSchema` | Gestión de bases de datos y schemas |
| `JsonQuote`, `EscapeJSON`, `Normalize` | Utilidades de escape de valores |

```go
import _ "github.com/celsiainternet/jdb/drivers/postgres" // registra el driver
```

---

## Paquetes de features

Los cuatro siguen el mismo patrón: `Load(db, schema, ...)` crea la tabla una sola vez y guarda un singleton; después se usan las funciones de paquete (o los métodos del objeto que devuelve `Define`).

### `instances`

| Función | Descripción |
|---|---|
| `Load(db, schema, name) (*Instance, error)` | Inicializa la tabla |
| `Set(id, tag, obj)` | Guarda un objeto serializable |
| `Get(id, &dest) (bool, error)` | Lee un objeto; `false` si no existe |
| `Delete(id)` | Elimina |
| `Query(et.Json)` | Consulta JSON |

```go
instances.Load(db, "core", "instances")
instances.Set("job-1", "jobs", myStruct)
var out MyStruct
found, err := instances.Get("job-1", &out)
```

### `authorization`

| Función | Descripción |
|---|---|
| `Load(db, schema)` | Inicializa y registra el store en `elvis/middleware` |
| `SetPath(method, path)`, `RemovePath(method, path)` | Registra / quita una ruta protegida |
| `SetAuthor(project, profile, method, path)`, `RemoveAuthor(...)` | Concede / revoca permiso |
| `Author(project, profile, method, path) (bool, error)` | Verifica permiso |
| `InitEvent(project, profiles)` | Inicializa permisos de un proyecto |
| `Query(et.Json)` | Consulta JSON |

```go
authorization.Load(db, "auth")
authorization.SetAuthor("p1", "admin", "GET", "/users")
ok, err := authorization.Author("p1", "admin", "GET", "/users")
```

### `config`

| Función | Descripción |
|---|---|
| `Load(db, schema)` | Inicializa la tabla |
| `Set(tag, stage, et.Json)` | Guarda una configuración |
| `Get(tag, stage) (et.Item, error)` | Lee una configuración |
| `Delete(tag, stage)` | Elimina |
| `Query(et.Json)` | Consulta JSON |

```go
config.Load(db, "core")
config.Set("smtp", "prod", et.Json{"host": "smtp.example.com", "port": 587})
item, err := config.Get("smtp", "prod")
```

### `inbox`

| Función | Descripción |
|---|---|
| `Load(db, schema)` | Inicializa la tabla |
| `UpsertInboxes(project, id, client, app, kind, data, user)` | Crea o actualiza un mensaje |
| `StateInboxes(id, status, user)` | Cambia el estado |
| `GetInboxesById(id)`, `GetInboxesByCode(kind, code)` | Lectura puntual |
| `GetInboxesByUserId(user, app, kind, status, page, rows)`, `GetInboxesByClientId(client, app, status, page, rows)` | Listados paginados |
| `GenInboxesCode(project)` | Genera un código consecutivo |
| `QueryInboxes(et.Json)` | Consulta JSON |

```go
inbox.Load(db, "core")
item, err := inbox.UpsertInboxes("p1", "", "c1", "app", "ticket", et.Json{"msg": "Hola"}, "u1")
items, err := inbox.GetInboxesByClientId("c1", "app", "pending", 1, 20)
```

---

## Variables de entorno

| Variable | Default | Propósito |
|---|---|---|
| `DB_DRIVER` | `postgres` | Driver que carga `jdb.Load()` |
| `DB_NAME` | `jdb` | Base de datos |
| `DB_HOST` | `localhost` | Host |
| `DB_PORT` | `5432` | Puerto |
| `DB_USER` | `admin` | Usuario |
| `DB_PASSWORD` | `admin` | Contraseña |
| `APP_NAME` | `jdb` | Nombre de la aplicación en la conexión |
| `DB_VERSION` | `13` | Versión de PostgreSQL |
| `DEBUG` | `false` | Imprime el SQL generado |

## Limitaciones conocidas

- Los valores se insertan como literales en el SQL (no parametrizados): no pases entrada de usuario en `CALC(...)` ni como nombre de campo.
- `Select("a", "b")` hoy genera `SELECT *`.
- `Update`/`Delete` con `Where` y `Upsert` tienen defectos en el SQL generado; verifica con `Debug()` antes de usarlos en producción.
- No hay tests (`*_test.go`) en el repositorio.

## Herramientas de `cmd/`

```bash
go run ./cmd/test                                       # sandbox contra una base real
go run github.com/celsiainternet/jdb/cmd/install        # instala dependencias de terceros
go run github.com/celsiainternet/jdb/cmd/create go      # CLI de scaffolding de proyectos/modelos
```

## Versionado

```bash
git add . && git commit -m 'Update version' && ./version.sh --r   # parche  X.Y.Z+1
./version.sh --n                                                  # menor    X.Y+1.0
./version.sh --m                                                  # mayor    X+1.0.0
```

`version.sh` actualiza la versión en este README y publica el tag.
