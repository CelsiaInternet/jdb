# JDB - Go Database Library

[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/Version-v1.0.120-orange.svg)](https://github.com/celsiainternet/jdb/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-celsiainternet%2Fjdb-black.svg)](https://github.com/celsiainternet/jdb)

## Instalación

```bash
# En una carpeta nueva
go mod init github.com/<organizacion>/<proyecto>

go get github.com/celsiainternet/jdb@v1.0.122
go run github.com/celsiainternet/jdb/cmd/install
go run github.com/celsiainternet/jdb/cmd/create go
```

`jdb` depende de `elvis` (llega como dependencia de jdb). `cmd/install` completa el `go.mod` con `go mod tidy` (solo si el proyecto ya tiene código; no fija versiones) e instala la guía de agentes de jdb en `.claude/guides/jdb/`, referenciada desde `CLAUDE.md` y `AGENTS.md`. Para la guía de elvis: `go run github.com/celsiainternet/elvis/cmd/install`.

La opción **Project** de `cmd/create` genera un microservicio completo sobre elvis y jdb (`cmd/`, `internal/services/`, `internal/models/`, `pkg/`, `.env`, `README.md`, reglas para agentes y las guías de jdb y elvis) y termina con `go mod tidy`. **Modelo** agrega un modelo jdb con su CRUD y registra sus rutas en `router.go`; **Rpc** agrega `rpc.go` a un paquete existente.

## Qué es

`jdb` es una librería de Go para trabajar con bases de datos: defines modelos de forma declarativa y construyes consultas (`Ql`) y comandos (`Command`) con una API fluida; un driver traduce esa intención a SQL.

- **Drivers:** **PostgreSQL** (`drivers/postgres`), **SQLite** (`drivers/sqlite`, `modernc.org/sqlite`, Go puro sin CGO) y **Oracle 19c+** (`drivers/oracle`, `sijms/go-ora`). No hay driver MySQL (`MysqlDriver` es solo una constante de nombre).
- **Dependencia:** [`elvis`](https://github.com/celsiainternet/elvis) (tipos `et.Json`/`et.Item`/`et.Items`, logging, eventos).
- **No es una aplicación:** `cmd/` contiene tres herramientas independientes (ver [Herramientas](#herramientas-de-cmd)).

## Paquetes

| Paquete         | Import                                           | Propósito                                                            |
| --------------- | ------------------------------------------------ | -------------------------------------------------------------------- |
| `jdb`           | `github.com/celsiainternet/jdb/jdb`              | Núcleo: conexión, modelos, consultas, comandos, hooks, handlers HTTP |
| `postgres`      | `github.com/celsiainternet/jdb/drivers/postgres` | Driver PostgreSQL (se registra con import en blanco)                 |
| `sqlite`        | `github.com/celsiainternet/jdb/drivers/sqlite`   | Driver SQLite: archivo o `:memory:` (import en blanco)               |
| `oracle`        | `github.com/celsiainternet/jdb/drivers/oracle`   | Driver Oracle 19c+ (import en blanco)                                |
| `instances`     | `github.com/celsiainternet/jdb/instances`        | Almacén genérico clave → objeto                                      |
| `authorization` | `github.com/celsiainternet/jdb/authorization`    | Permisos por proyecto/perfil/método/ruta                             |
| `config`        | `github.com/celsiainternet/jdb/config`           | Configuración por `tag` + `stage`                                    |
| `inbox`         | `github.com/celsiainternet/jdb/inbox`            | Bandeja de mensajes/solicitudes                                      |

---

## Paquete `jdb`

### Conexión y registro

| Función / método                                                  | Descripción                                                                         |
| ----------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `Load() (*DB, error)`                                             | Conecta usando el driver de `DB_DRIVER` (default `postgres`) y variables de entorno |
| `LoadTo(database, hostname...)`                                   | Igual que `Load` pero con otro nombre de base de datos                              |
| `ConnectTo(ConnectParams) (*DB, error)`                           | Conecta con parámetros explícitos                                                   |
| `LoadConnectParams(et.Json)`                                      | Construye `ConnectParams` desde JSON                                                |
| `Register(name, factory, params)`                                 | Registra un driver (lo usan los drivers en su `init()`)                             |
| `NewDatabase(name, driver)`                                       | Crea un `*DB` sin conectar                                                          |
| `Jdb()`, `GetDB(name)`, `GetSchema("schema.x")`, `GetModel(name)` | Acceso al registro global                                                           |

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

// o, desde variables de entorno (DB_DRIVER elige postgres, sqlite u oracle):
// import _ "github.com/celsiainternet/jdb/drivers/postgres"
// db, err := jdb.Load()
```

`Load()`/`LoadTo()` crean por defecto los modelos del schema `core` (series); con `USE_CORE=false` no lo hacen.

### `*DB`

| Método                                                  | Descripción                                |
| ------------------------------------------------------- | ------------------------------------------ |
| `Ping()`, `HealthCheck()`                               | Verifica la conexión                       |
| `Conected(params)`, `Disconected()`                     | Abre / cierra la conexión                  |
| `GetSchema(name)`, `GetModel(name)`, `DropSchema(name)` | Schemas y modelos registrados              |
| `LoadModel`, `MutateModel`, `DropModel`, `EmptyModel`   | DDL sobre un modelo                        |
| `From(table) *Ql`                                       | Inicia una consulta                        |
| `Select(ql)`, `Count(ql)`, `Exists(ql)`, `Command(cmd)` | Ejecuta un `Ql`/`Command` ya construido    |
| `Query(sql, args...)`, `One(sql, args...)`              | SQL crudo, devuelve `et.Items` / `et.Item` |
| `JQuery(et.Json)`, `QueryModel(et.Json)`                | Consulta definida en JSON                  |
| `SetDebug(bool)`, `Debug()`, `Describe()`               | Depuración y descripción                   |

```go
items, err := db.Query("SELECT * FROM public.users WHERE age > $1", 18)
```

### Schemas y modelos

| Función / método                                          | Descripción                                 |
| --------------------------------------------------------- | ------------------------------------------- |
| `NewSchema(db, name)`                                     | Crea/obtiene un schema                      |
| `NewModel(schema, name, version)`                         | Crea/obtiene un modelo (tabla)              |
| `NewTable(db, "schema.table")`                            | Modelo a partir de un nombre completo       |
| `Model.Init()`                                            | Crea/actualiza la tabla en la base de datos |
| `Model.Drop()`, `Model.Empty()`                           | Elimina / vacía la tabla                    |
| `Model.GetField(name)`, `Model.GetModel(name)`            | Resuelve campos y modelos relacionados      |
| `Model.GenId()`, `Model.GetId(id)`                        | Genera / normaliza identificadores          |
| `Model.CheckRequired(data)`                               | Valida campos requeridos                    |
| `Model.Counted()`, `Model.Query(et.Json)`                 | Conteo y consulta JSON                      |
| `Model.On(channel, handler)`, `Model.Emit(channel, data)` | Eventos por modelo                          |
| `Model.Describe()`, `Model.Debug()`                       | Descripción y depuración                    |

### Definición de columnas (`Define*`)

| Método                                                                                                                                            | Descripción                                     |
| ------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------- |
| `DefineColumn(name, TypeData) *Column`                                                                                                            | Columna real                                    |
| `DefineAtribute(name, TypeData) *Column`                                                                                                          | Atributo dentro de la columna jsonb fuente      |
| `DefineSourceField()`, `DefineSource(name)`                                                                                                       | Columna jsonb fuente (`_data`)                  |
| `DefinePrimaryKey(cols...)`, `DefinePrimaryKeyField()`                                                                                            | Llave primaria                                  |
| `DefineForeignKey(fks, with, onDelete, onUpdate)`                                                                                                 | Llave foránea                                   |
| `DefineIndex(sort, cols...)`, `DefineUnique(cols...)`                                                                                             | Índices (ver [Índices](#índices))               |
| `DefineRequired(cols...)`, `DefineHidden(cols...)`                                                                                                | Requeridos / ocultos (ver abajo)                |
| `DefineCreatedAtField()`, `DefineUpdatedAtField()`, `DefineStatusField()`, `DefineSystemKeyField()`, `DefineIndexField()`, `DefineProjectField()` | Campos de sistema (se rellenan automáticamente) |
| `DefineRelation(name, with, fks, limit)`                                                                                                          | Relación uno-a-muchos                           |
| `DefineRollup(name, withName, fks, fields).SetWheres(...)`                                                                                        | Agregación desde otra tabla                     |
| `DefineObject(name, with, fks, fields)`                                                                                                           | Objeto embebido uno-a-uno                       |
| `DefineDetail(name, fks, limit)`                                                                                                                  | Detalle                                         |
| `DefineSearch(cols...)`                                                                                                                           | Campos de la búsqueda de texto (ver Búsqueda)   |
| `DefineCalc(name, fn)`, `DefineCalcTx(name, fn)`                                                                                                  | Declara un campo calculado; **hoy no se ejecuta** (ver abajo) |
| `DefineModel()`, `DefineProjectModel()`                                                                                                           | Conjunto estándar de campos de sistema          |
| `DefineIntegrity()`                                                                                                                               | Activa integridad referencial                   |

`NewModel(schema, name, version)` devuelve el modelo ya registrado si existe uno con ese nombre en la base (no lo redefine). `DefineColumn` crea una columna real; `DefineAtribute` guarda el valor dentro de la columna JSON fuente (`_data`). Un nombre ya definido no se vuelve a crear.

Tipos de dato (`TypeData*`) y su tipo en PostgreSQL:

| Tipo | PostgreSQL | Valor por defecto |
| --- | --- | --- |
| `TypeDataText`, `TypeDataSelect`, `TypeDataUser`, `TypeDataEmail`, `TypeDataPhone` | `VARCHAR(250)` | `''` |
| `TypeDataShortText`, `TypeDataKey`, `TypeDataState`, `TypeDataRelation`, `TypeDataRollup` | `VARCHAR(80)` | `''` |
| `TypeDataMemo`, `TypeDataFilesMedia`, `TypeDataUrl`, `TypeDataAddress`, `TypeDataFullText` | `TEXT` | `''` |
| `TypeDataNumber` | `DECIMAL(18,2)` | `0.0` |
| `TypeDataInt` | `BIGINT` | `0` |
| `TypeDataPrecision` | `DOUBLE PRECISION` | `''` |
| `TypeDataDateTime` | `TIMESTAMP` | `''` |
| `TypeDataCheckbox` | `BOOLEAN` | `false` |
| `TypeDataBytes` | `BYTEA` | `''` |
| `TypeDataObject` | `JSONB` | `{}` |
| `TypeDataMultiSelect` | `JSONB` | `[]` |
| `TypeDataGeometry` | `JSONB` | `{"type": "Point", "coordinates": [0, 0]}` |

`TypeDataFullText` es obsoleto: es texto plano; para buscar usa `DefineSearch`. SQLite y Oracle usan sus equivalentes (ver la tabla de dialectos).

Modificadores de `*Column`: `SetDefaultValue`, `SetHidden`, `SetMax`, `SetMin`.

**Ocultos (`DefineHidden`).** Marca columnas como ocultas (`Column.Hidden`). Una consulta sin `Select`/`Data` explícito (también `RunQuery` sin `selects`) no trae las **columnas** ocultas; salen solo si se piden por nombre. Tres matices verificados: un **atributo** oculto (`DefineAtribute`) sigue saliendo, porque viaja dentro de `_data`; `RunQuery` acepta una columna oculta en `selects` y en `wheres` (no la rechaza); y `Ql.Hidden(campos...)` oculta campos solo en esa consulta.

**Calculados (`DefineCalc`, `DefineCalcTx`).** Agregan una columna virtual (`TpCalc`) y guardan la función en `Model.CalcFunction` / `Model.CalcFunctionTx`, pero **ninguna consulta ejecuta esas funciones**: `getDetailsTx` solo resuelve rollups, así que `Detail("campo")` o `Select("campo")` no agregan el valor. Para un dato calculado, complétalo en tu código después de `One()`/`All()`, o usa un rollup (`DefineRollup`/`NewRollup`) si sale de otra tabla.

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

### Índices

| Método | Qué crea en PostgreSQL |
| --- | --- |
| `DefineIndex(sort, cols...)` | Un índice `<modelo>_<col>_idx` por columna: B-tree en columnas, GIN `jsonb_path_ops` en `TypeDataObject` y un índice de expresión `(_data#>>'{attr}')` en atributos (con `DB_VERSION` ≥ 13). `sort` se guarda en el modelo pero no cambia el DDL. |
| `DefineUnique(cols...)` | `CREATE UNIQUE INDEX` por columna (solo columnas reales, no atributos). |
| `DefinePrimaryKey(cols...)` | Restricción `<modelo>_pk`. |
| `DefineForeignKey(fks, with, onDelete, onUpdate)` | `FOREIGN KEY ... REFERENCES` con `ON DELETE/UPDATE CASCADE` opcionales. |
| `DefineSearch(cols...)` | Índice GIN de trigramas de la búsqueda (ver [Búsqueda](#búsqueda-definesearch)). |

Los índices, llaves y columnas nuevas se aplican **solo cuando `Init()` crea la tabla** (o con `db.MutateModel(model)`, que copia la tabla a una nueva). Si la tabla ya existe, `Init()` no agrega columnas ni índices: solo asegura el índice de búsqueda. Un `DefineIndex` agregado a un modelo existente hay que crearlo a mano o con `MutateModel`.

### Consultas (`*Ql`)

Se inicia con `jdb.From(model | "schema.table")`, `db.From(...)`, `Model.Where(...)`, `Model.Select(...)`, `Model.Data(...)` o `Model.Join(...)`. Los métodos mutan el `Ql` y lo devuelven; no reutilices un `Ql` después de ejecutarlo.

| Grupo       | Métodos                                                                                                | Descripción                                                   |
| ----------- | ------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------- |
| Condiciones | `Where`, `And`, `Or`, `WhereConcat`, `WhereSearch`                                                     | Abren una condición sobre un campo (o varios unidos)          |
| Operadores  | `Eq`, `Neg`, `In`, `NotIn`, `Like`, `Search`, `More`, `Less`, `MoreEq`, `LessEq`, `Between`, `IsNull`, `NotNull` | Completan la última condición                       |
| Proyección  | `Select`, `Data`, `Detail`, `Hidden`                                                                   | Campos a devolver                                             |
| Joins       | `Join`, `LeftJoin`, `RightJoin`, `FullJoin`                                                            | Unión con otro modelo                                         |
| Agrupación  | `GroupBy`, `Having`                                                                                    | Agrupa y filtra grupos                                        |
| Orden       | `OrderBy`, `OrderByAsc`, `OrderByDesc`                                                                 | Orden de resultados                                           |
| Paginación  | `Page(p)` + `Rows(n)`, `List(page, rows)`                                                              | `List` devuelve `et.List` con total                           |
| Ejecución   | `All()`, `One()`, `Rows(n)`, `First(n)`, `Last(n)`, `Counted()`, `ItExists()`                          | `First(n)`/`Last(n)` devuelven **el ítem en la posición `n`** |
| Transacción | `AllTx`, `OneTx`, `RowsTx`, `FirstTx`, `LastTx`, `CountedTx`, `ItExistsTx`                             | Igual, dentro de un `*Tx`                                     |
| JSON        | `Query(et.Json)`, `QueryTx`                                                                            | Ejecuta una consulta descrita en JSON                         |
| Otros       | `Debug()`, `Describe()`, `Tx()`                                                                        | Depuración                                                    |

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

// Distinto de, en una lista
items, err = users.Where("status").Neg("deleted").And("role").In("admin", "editor").All()

// Data arma el resultado como objeto JSON; con jdb.SOURCE (_data) entre los campos, los agrega al contenido de la fuente
item, err = users.Data(jdb.SOURCE, "id", "name").Where("id").Eq("u1").One()

// Join (con *Model)
items, err = jdb.From(users).Join(profiles, "id", "=", "user_id").All()

// Ruta JSON: flechas de Postgres o puntos, desde una columna JSON (TypeDataObject) o dentro del source (_data)
items, err = scheduling.Where("customer_id").Eq(customerId).
	And("_data->'sync_external'->>'oss_order_id'").Eq(ossOrderId). // igual a "_data.sync_external.oss_order_id" o "sync_external.oss_order_id"
	OrderByDesc("start_date").All()
```

Resultados: `All()` devuelve `et.Items` (`Ok`, `Count`, `Result`); `One()` devuelve `et.Item` y, si no hay fila, `Ok == false` sin error (no devuelve `ErrNotFound`). `Counted()` hace un `COUNT` con las mismas condiciones y `ItExists()` un `EXISTS`. `Select(campos...)` y `Data(campos...)` devuelven solo esos campos; `Data` los arma como objeto JSON y, si entre ellos va `jdb.SOURCE` (`_data`), los agrega al contenido de la fuente. `Where`, `And` y `Or` abren la condición y el operador siguiente (`Eq`, `Neg`, `In`, …) la completa; un nombre que no es campo se envía como texto literal (`'age'`), no como columna.

Las llaves de una ruta JSON solo admiten letras, números, `_` y `-`; si no, el nombre no se resuelve como campo. En Postgres la ruta se lee como texto y se convierte a `numeric` o `boolean` cuando el valor comparado es un número o un booleano.

`WhereConcat("first_name", "last_name")` abre una sola condición sobre varios campos unidos por un espacio:

```go
items, err := jdb.From(users).WhereConcat("first_name", "last_name").Like("%ana gómez%").All()
```

### Búsqueda (`DefineSearch`)

`DefineSearch(campos...)` declara los campos de texto (columnas o atributos ya definidos; las fechas se ignoran) en los que busca el modelo. `WhereSearch(texto)` agrega una condición por palabra (hasta `SEARCH_MAX_WORDS`, 10): cada palabra debe aparecer en alguno de esos campos, también como parte de una palabra. En `RunQuery` es la llave `"search"`.

```go
customers.DefineSearch("full_name", "name", "last_name")
// ... Init()

items, err := customers.WhereSearch("lui dia").All()                 // Luisa Díaz
items, err = customers.Where("kind").Eq("lead").WhereSearch("gomez").All()
list, err := jdb.RunQuery(customers, et.Json{"search": "ana gomez", "page": 1, "rows": 30})
```

En Postgres la condición es `jdb_unaccent(lower(campos unidos)) LIKE '%' || jdb_unaccent(lower(palabra)) || '%'`: no distingue tildes ni mayúsculas y usa un índice GIN de trigramas (`pg_trgm`). Al iniciar el modelo el driver crea, si no existen, las extensiones `pg_trgm` y `unaccent`, la función `public.jdb_unaccent` (unaccent marcada `IMMUTABLE`) y el índice `<tabla>_search_<hash>` con la misma expresión, también en tablas ya creadas. Requisitos: permiso `CREATE` en la base (desde Postgres 13 las dos extensiones no exigen superusuario); con palabras de menos de 3 letras el índice ayuda poco. Los campos que mandan son los de `DefineSearch` de cada arranque (no los del modelo guardado): si cambian, se crea el índice nuevo y se borran los `<tabla>_search_<hash>` anteriores de esa tabla (sin campos, se borran todos); los demás índices no se tocan. En SQLite y Oracle es un `LIKE` sobre los campos unidos, sin índice ni manejo de tildes. `%` y `_` en el texto actúan como comodines. El operador `search` sobre un solo campo sigue siendo `ILIKE '%v%'` sin índice.

### Consultas en JSON (`RunQuery`)

`RunQuery(model, query, rollups...)` arma la consulta desde un `et.Json` y devuelve una página `et.List` con el total. Es el contrato de los `POST /query` de los servicios (catalog, vcore). El detalle de todos los formatos, con el SQL que generan, está en [jsql.md](jsql.md).

```go
list, err := jdb.RunQuery(users, et.Json{
    "wheres": []et.Json{
        {"age": et.Json{"moreEq": 18}},
        {"or": et.Json{"name": et.Json{"like": "%Ana%"}}},
    },
    "search":   "ana gomez",
    "selects":  []string{"id", "name", "age"},
    "order_by": et.Json{"desc": []string{"age"}},
    "page":     1,
    "rows":     20,
})
```

| Llave | Formato | Comportamiento |
| --- | --- | --- |
| `wheres` | lista de condiciones | Se aplican en orden. Cada elemento es `{campo: {op: valor}}` (une con `AND`), `{"and": {campo: {op: valor}}}`, `{"or": {campo: {op: valor}}}` o `{"concat": {"fields": [...], op: valor}}` (los campos unidos por un espacio). Sin paréntesis: `AND` va antes que `OR`. |
| `search` | **texto** | Busca cada palabra (hasta 10) en los campos de `DefineSearch`; se aplica con `AND` después de `wheres`. Vacío no filtra. Si el modelo no declaró campos, no devuelve filas. |
| `selects` | lista de nombres | Campos a devolver. Sin `selects`, el registro completo sin las columnas ocultas. |
| `order_by` | `{"asc": [...], "desc": [...]}` | Aplica primero `asc` y luego `desc`. Un nombre que no es campo devuelve el error `Columna no válida en el query (x)`. |
| `page` | entero | Desde 1; por defecto 1. |
| `rows` | entero | Filas por página; alias `limit`. Por defecto `QUERY_DEFAULT_LIMIT` (100) y como máximo `QUERY_MAX_LIMIT` (1000). |
| `join`, `left_join`, `right_join`, `full_join` | lista | Se leen, pero hoy no generan el `JOIN`. |

Operadores de `wheres` (distinguen mayúsculas): `eq`, `neg`, `more`, `less`, `moreEq`, `lessEq`, `like` (en Postgres `ILIKE`), `in` (lista), `search` (contiene el texto, sin índice), `isNull` y `notNull` (el valor se ignora). `notIn` y `between` **no** están soportados (ver Limitaciones). Va un operador por condición: dos operadores en el mismo objeto escriben sobre la misma condición.

`search` es un texto, no un objeto. El formato `{"search": {"value": "...", "fields": [...]}}` de los `/query` de catalog lo traduce catalog a condiciones `concat` antes de llamar a `RunQuery`; si se pasa tal cual a jdb, se convierte a texto (`map[...]`) y la consulta no devuelve filas, sin error.

Respuesta (`et.List`):

```json
{ "rows": 20, "all": 57, "count": 20, "page": 1, "start": 1, "end": 20, "result": [ { "id": "u1", "name": "Ana", "age": 31 } ] }
```

`all` es el total que cumple los filtros, `count` lo que trae la página y `start`/`end` las posiciones de la página contando desde 1.

#### `jdb.RunQuery` y `linq.RunQuery` de elvis

`linq.RunQuery(model, query, wheres...)` (`elvis/linq/jsql.go`, aún sin publicar en elvis) replica este contrato para los modelos `linq` de elvis. Las diferencias, leídas en el código de ambos:

| Tema | `jdb.RunQuery` | `linq.RunQuery` (elvis) |
| --- | --- | --- |
| `wheres` con `eq`, `neg`, `more`, `less`, `moreEq`, `lessEq`, `like`, `in`, `and`, `or`, `concat` | Sí | Sí |
| `isNull`, `notNull` | Sí | No (se ignoran) |
| Operador `search` por campo y llave raíz `search` | Sí | No (se ignoran) |
| `notIn`, `between` | No (`notIn` devuelve 0 filas, `between` todas) | No (se ignoran) |
| Varios operadores en una condición | Uno solo (el último que recorre el mapa) | Cada uno es una condición `AND` |
| Precedencia con `or` | Sin paréntesis (`AND` antes que `OR`) | Las condiciones de `wheres` van entre paréntesis si hay un `or` |
| Campo desconocido en `wheres` | Se ignora | Se ignora (también si es oculto) |
| Campo desconocido u oculto en `selects` | Se ignora / se devuelve | Error `Columna no valida en la consulta (x)` |
| `order_by` desconocido | Error | Error (también si es oculto) |
| `page`, `rows`/`limit`, 100 por defecto y máximo 1000, respuesta `et.List` | Sí | Sí |
| Tercer argumento | `rollups ...*Rollup` | `wheres ...*Where`: condiciones fijas del llamador que se aplican con `AND` |
| `join`, `left_join`, … | Se leen y se ignoran | No se leen |
| Valores | Literales en el SQL con `'` duplicada | Literales con `'` duplicada; solo texto, número o booleano |

### Rollups

`DefineRollup(name, withName, fks, fields)` agrega una columna virtual que, cuando se pide en `Select`/`Detail`, consulta el modelo `withName` por cada fila uniendo por `fks` (campo local → campo en `withName`). `SetWheres(et.Json{...})` añade filtros fijos. Para rollups que vienen de datos y no del modelo compartido, crea uno con `NewRollup(name, with, fks, fields, show)` y pásalo por consulta con `Ql.AddRollups(...)` o `RunQuery(model, query, rollup)`.

```go
users.DefineRollup("city_name", cities.Name, map[string]string{"city_id": "id"}, []string{"name"})
items, err := users.Select("id", "name", "city_name").All()
```

### Comandos (`*Command`)

| Inicio (`*Model`) | Descripción                                |
| ----------------- | ------------------------------------------ |
| `Insert(et.Json)` | Inserta una fila                           |
| `Bulk([]et.Json)` | Inserta varias filas                       |
| `Update(et.Json)` | Actualiza las filas que cumplen el `Where` |
| `Delete(field)`   | Elimina; abre la condición sobre `field`   |
| `Upsert(et.Json)` | Busca por la llave primaria de los datos: si existe actualiza, si no inserta |

| Método                                                    | Descripción                                     |
| --------------------------------------------------------- | ----------------------------------------------- |
| `Where`, `And`, `Or` + operadores (`Eq`, `In`, `Like`, …) | Condiciones, igual que en `Ql`                  |
| `Returns(fields...)`                                      | Campos a devolver                               |
| `Exec()`, `One()`                                         | Ejecuta (abre y confirma su propia transacción) |
| `ExecTx(tx)`, `OneTx(tx)`                                 | Ejecuta dentro de una transacción existente     |
| `Debug()`, `Describe()`                                   | Depuración                                      |

```go
item, err := users.Insert(et.Json{"id": "u1", "name": "Ana", "email": "ana@example.com"}).One()

items, err := users.Update(et.Json{"age": 31}).Where("id").Eq("u1").Exec()

items, err = users.Delete("id").Eq("u1").Exec()

items, err = users.Bulk([]et.Json{
    {"id": "u2", "name": "Carlos", "email": "carlos@example.com"},
    {"id": "u3", "name": "María", "email": "maria@example.com"},
}).Exec()

items, err = users.Upsert(et.Json{"id": "u1", "name": "Ana María"}).Exec()
```

`Exec()` devuelve `et.Items` con las filas escritas. `Update` y `Delete` primero leen las filas que cumplen el `Where` (sin `Where` devuelven error) y ejecutan los hooks por cada una. Si no hay filas que escribir, el resultado trae `Ok == false` y **no hay error**. `Bulk` inserta todas las filas, pero `Result` trae solo la última.

### Errores

`jdb` exporta `ErrNotInserted` (`record not inserted`), `ErrNotUpdated`, `ErrNotFound` (`record not found`) y `ErrNotUpserted`, pero **la librería no los devuelve en ningún camino**: un `One()` sin fila o un `Update` sin coincidencias responden `Ok == false` sin error. Son centinelas para que el código del consumidor los devuelva desde un hook y los compare después:

```go
_, err := users.Update(data).
    BeforeUpdate(func(tx *jdb.Tx, data et.Json) error {
        ok, err := users.Where("id").Eq(id).ItExistsTx(tx)
        if err != nil {
            return err
        }
        if !ok {
            return jdb.ErrNotFound
        }
        return nil
    }).
    Where("id").Eq(id).Exec()
if errors.Is(err, jdb.ErrNotFound) {
    // 404
}
```

Los errores de la base dicen en qué modelo ocurrieron: consultas (`All`, `One`, `ItExists`, `Counted`…), escrituras (`Insert`, `Update`, `Delete`, `Upsert`) y la carga del modelo (`Init`) anteponen `schema.tabla` y, en consultas y escrituras, el alias:

```text
projects.components (a): pq: column a.kind does not exist at position 13:9 (42703)
```

El error original queda envuelto, así que `errors.Is` y `errors.As` (por ejemplo, para leer el `*pq.Error`) siguen funcionando. Los errores que devuelve un hook no cambian. `DB.Query` con SQL a mano no tiene modelo y sigue igual.

### Hooks

Disponibles en `*Model` (aplican a todos los comandos) y en `*Command` (solo a ese comando). Por cada fila corren en este orden: `Before*` del modelo, `Before*Trigger` del modelo, `Before*` del comando, `Before*Trigger` del comando, la escritura y luego los `After*` en el mismo orden. Los `Before*` reciben los datos que se van a escribir y pueden modificarlos; un error de cualquier hook detiene el comando; con `Exec()` la transacción propia se revierte y con `ExecTx(tx)` decide quien la abrió. `BeforeInsertOrUpdate` registra la misma función en insert y update. Con `Upsert` corren los hooks de insert o de update según el caso.

| Método                                                                 | Firma de la función                        |
| ---------------------------------------------------------------------- | ------------------------------------------ |
| `BeforeInsert`, `BeforeUpdate`, `BeforeDelete`, `BeforeInsertOrUpdate` | `func(tx *jdb.Tx, data et.Json) error`     |
| `AfterInsert`, `AfterUpdate`, `AfterDelete`, `AfterInsertOrUpdate`     | `func(tx *jdb.Tx, data et.Json) error`     |
| `Before…Trigger`, `After…Trigger` (mismas variantes)                   | `func(tx *jdb.Tx, old, new et.Json) error` |

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

`jdb.Tx` envuelve un `*sql.Tx`. Sin transacción (`Exec()`, `All()`, …), cada comando o consulta abre la suya y la confirma o revierte según el error. Para agrupar varias escrituras, abre una con `db.Begin()`, pásala a las variantes `…Tx` y ciérrala con `tx.End(err)` (revierte si `err` no es nil; si no, confirma y devuelve el error del commit). También están `tx.Commit()` y `tx.Rollback()`. Los hooks reciben el `tx` en curso: úsalo en las consultas que hagan dentro del hook.

```go
tx, err := db.Begin()
if err != nil {
    return err
}

_, err = users.Insert(et.Json{"id": "u4", "name": "Luis", "email": "luis@example.com"}).ExecTx(tx)
if err == nil {
    _, err = profiles.Insert(et.Json{"user_id": "u4"}).ExecTx(tx)
}

return tx.End(err)
```

### Series y utilidades

| Función                                                                                              | Descripción                                                                |
| ---------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `NewSerie(kind, tag, format, last)`, `SetSeries(...)`, `GetSeries(kind, tag)`, `ImportSeries(items)` | Consecutivos en el schema `core` (requiere `UseCore`, activo por defecto con `USE_CORE=true`) |
| `Query(db, sql, args...)`, `QueryTx(db, tx, sql, args...)`                                           | SQL crudo                                                                  |
| `RowsToItems(*sql.Rows)`                                                                             | Convierte filas a `et.Items`                                               |

### Handlers HTTP

| Handler         | Descripción                                |
| --------------- | ------------------------------------------ |
| `ModelDefine`   | Define un modelo desde JSON                |
| `ModelQuery`    | Ejecuta un `Ql` descrito en el body JSON   |
| `ModelCommand`  | Ejecuta comandos descritos en el body JSON |
| `ModelDescribe` | Describe un objeto por tipo y nombre       |

```go
r := chi.NewRouter()
r.Post("/model/query", jdb.ModelQuery)
r.Post("/model/command", jdb.ModelCommand)
r.Post("/model/define", jdb.ModelDefine)
r.Post("/model/describe", jdb.ModelDescribe)
```

---

## Paquete `drivers/postgres`

| Elemento                                                                      | Descripción                                                                   |
| ----------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| `Connection{Database, Host, Port, Username, Password, App, Version, IsDebug}` | Parámetros de conexión (`Chain()`, `Validate()`, `Load(et.Json)`, `ToJson()`) |
| `Params.CreateUser`, `ChangePassword`, `DeleteUser`, `GrantPrivileges`        | Gestión de usuarios de PostgreSQL                                             |
| `Params.ExistDatabase`, `CreateDatabase`, `DropDatabase`, `DropSchema`        | Gestión de bases de datos y schemas                                           |
| `JsonQuote`, `EscapeJSON`, `Normalize`                                        | Utilidades de escape de valores                                               |

Con `DefineSearch`, `Init()` crea si faltan las extensiones `pg_trgm` y `unaccent`, la función `public.jdb_unaccent` y el índice GIN de trigramas de la búsqueda (también en tablas ya creadas): el usuario necesita permiso `CREATE` en la base.

```go
import _ "github.com/celsiainternet/jdb/drivers/postgres" // registra el driver
```

## Paquetes `drivers/sqlite` y `drivers/oracle`

Ambos se comportan como el driver de PostgreSQL, atributos dinámicos incluidos. Diferencias de dialecto:

| Tema          | SQLite                                                          | Oracle 19c+                                                                                     |
| ------------- | --------------------------------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Conexión      | `sqlite.Connection{Database}`: ruta del archivo o `":memory:"`; vacía usa `<DB_NAME>.db` | `oracle.Connection{Host, Port, Username, Password, ServiceName, SSL, SSLVerify, Version}` |
| Schemas       | `schema.table` se guarda como tabla `schema_table`               | Igual, dentro del usuario conectado                                                             |
| Columna JSON  | TEXT; los atributos se actualizan con `json_set`                 | `CLOB CHECK (IS JSON)` con `JSON_MERGEPATCH`: un `null` **borra** la llave y los objetos se fusionan |
| Valores       | INSERT/UPDATE con argumentos enlazados (`?N`)                    | INSERT/UPDATE con argumentos enlazados (`:N`)                                                   |
| Otros         | `Search` y `WhereSearch` son `LIKE '%v%'` sin índice ni tildes (en Postgres, `ILIKE` y el índice de trigramas) | `''` es `NULL`; `LIKE` compara con `UPPER()`; sin `ON UPDATE CASCADE`; el DDL hace commit implícito |

Los dos rechazan `UPDATE`/`DELETE` sin `WHERE`.

```go
import (
    "github.com/celsiainternet/jdb/drivers/sqlite"
    "github.com/celsiainternet/jdb/jdb"
)

db, err := jdb.ConnectTo(jdb.ConnectParams{
    Driver: jdb.SqliteDriver,
    Name:   "myapp",
    Params: &sqlite.Connection{Database: ":memory:"},
})
```

---

## Paquetes de features

Los cuatro siguen el mismo patrón: `Load(db, schema, ...)` crea la tabla una sola vez y guarda un singleton; después se usan las funciones de paquete (o los métodos del objeto que devuelve `Define`).

### `instances`

| Función                                     | Descripción                         |
| ------------------------------------------- | ----------------------------------- |
| `Load(db, schema, name) (*Instance, error)` | Inicializa la tabla                 |
| `Set(id, tag, obj)`                         | Guarda un objeto serializable       |
| `Get(id, &dest) (bool, error)`              | Lee un objeto; `false` si no existe |
| `Delete(id)`                                | Elimina                             |
| `Query(et.Json)`                            | Consulta JSON                       |

```go
instances.Load(db, "core", "instances")
instances.Set("job-1", "jobs", myStruct)
var out MyStruct
found, err := instances.Get("job-1", &out)
```

### `authorization`

| Función                                                          | Descripción                                          |
| ---------------------------------------------------------------- | ---------------------------------------------------- |
| `Load(db, schema)`                                               | Inicializa y registra el store en `elvis/middleware` |
| `SetPath(method, path)`, `RemovePath(method, path)`              | Registra / quita una ruta protegida                  |
| `SetAuthor(project, profile, method, path)`, `RemoveAuthor(...)` | Concede / revoca permiso                             |
| `Author(project, profile, method, path) (bool, error)`           | Verifica permiso                                     |
| `InitEvent(project, profiles)`                                   | Inicializa permisos de un proyecto                   |
| `Query(et.Json)`                                                 | Consulta JSON                                        |

```go
authorization.Load(db, "auth")
authorization.SetAuthor("p1", "admin", "GET", "/users")
ok, err := authorization.Author("p1", "admin", "GET", "/users")
```

### `config`

| Función                            | Descripción              |
| ---------------------------------- | ------------------------ |
| `Load(db, schema)`                 | Inicializa la tabla      |
| `Set(tag, stage, et.Json)`         | Guarda una configuración |
| `Get(tag, stage) (et.Item, error)` | Lee una configuración    |
| `Delete(tag, stage)`               | Elimina                  |
| `Query(et.Json)`                   | Consulta JSON            |

```go
config.Load(db, "core")
config.Set("smtp", "prod", et.Json{"host": "smtp.example.com", "port": 587})
item, err := config.Get("smtp", "prod")
```

### `inbox`

| Función                                                                                                            | Descripción                  |
| ------------------------------------------------------------------------------------------------------------------ | ---------------------------- |
| `Load(db, schema)`                                                                                                 | Inicializa la tabla          |
| `UpsertInboxes(project, id, client, app, kind, data, user)`                                                        | Crea o actualiza un mensaje  |
| `StateInboxes(id, status, user)`                                                                                   | Cambia el estado             |
| `GetInboxesById(id)`, `GetInboxesByCode(kind, code)`                                                               | Lectura puntual              |
| `GetInboxesByUserId(user, app, kind, status, page, rows)`, `GetInboxesByClientId(client, app, status, page, rows)` | Listados paginados           |
| `GenInboxesCode(project)`                                                                                          | Genera un código consecutivo |
| `QueryInboxes(et.Json)`                                                                                            | Consulta JSON                |

```go
inbox.Load(db, "core")
item, err := inbox.UpsertInboxes("p1", "", "c1", "app", "ticket", et.Json{"msg": "Hola"}, "u1")
items, err := inbox.GetInboxesByClientId("c1", "app", "pending", 1, 20)
```

---

## Variables de entorno

| Variable      | Default     | Propósito                              |
| ------------- | ----------- | -------------------------------------- |
| `DB_DRIVER`   | `postgres`  | Driver que carga `jdb.Load()`          |
| `DB_NAME`     | `jdb`       | Base de datos                          |
| `DB_HOST`     | `localhost` | Host                                   |
| `DB_PORT`     | `5432`      | Puerto                                 |
| `DB_USER`     | `admin`     | Usuario                                |
| `DB_PASSWORD` | `admin`     | Contraseña                             |
| `APP_NAME`    | `jdb`       | Nombre de la aplicación en la conexión |
| `DB_VERSION`  | `13`        | Versión de PostgreSQL                  |
| `DEBUG`       | `false`     | Imprime el SQL generado                |
| `USE_CORE`    | `true`      | `Load()` crea los modelos del schema `core` (series) |
| `SQLITE_PATH` | `<DB_NAME>.db` | Archivo SQLite (`:memory:` para memoria) |
| `ORA_DB_HOST`, `ORA_DB_PORT` (`1521`), `ORA_DB_USER`, `ORA_DB_PASSWORD`, `ORA_DB_SERVICE_NAME_ORACLE`, `ORA_DB_SSL_ORACLE`, `ORA_DB_SSL_VERIFY_ORACLE`, `ORA_VERSION` (`19`) | — | Conexión Oracle |

## Limitaciones conocidas

- Los valores del `WHERE` se insertan como literales en el SQL (en Postgres también los de INSERT/UPDATE, salvo la columna JSON): no pases entrada de usuario en `CALC(...)`, en agregaciones ni como nombre de campo.
- `Ql.Join("nombre", …)` con un string que no está en la consulta se ignora; pasa el `*Model`. El valor del lado derecho del join y de `Having` se trata como literal, no como campo.
- `RunQuery` descarta en silencio una llave de `wheres` que no es campo del modelo (un error de tipeo devuelve todas las filas) y no soporta `between` ni `notIn`.
- `Ql.Query`/`Model.Query`/`ModelQuery`: con `limit > 1` devuelve el ítem en esa posición (o entra en pánico); usa `limit: 0` para todas las filas. `having` no funciona.
- `ModelCommand`: solo `insert` y `bulk` funcionan; `update`/`delete` no cambian nada.
- `DefineCalc`/`DefineCalcTx` no se ejecutan en ninguna consulta (ver [Definición de columnas](#definición-de-columnas-define)).
- `DefineHidden` oculta columnas, no atributos: un atributo oculto sale dentro de `_data`. `RunQuery` no rechaza campos ocultos en `selects` ni en `wheres`.
- `Init()` no agrega columnas, índices ni llaves a una tabla que ya existe (solo el índice de búsqueda).
- `RunQuery` ignora en silencio los nombres desconocidos de `selects` y recibe `search` solo como texto.
- `ErrNotInserted`, `ErrNotFound` y los demás `ErrNot*` no los devuelve la librería.
- `Bulk` devuelve solo la última fila insertada.
- En SQLite, `GetSeries` no incrementa el consecutivo.
- Búsqueda (`DefineSearch`): `%` y `_` en el texto actúan como comodines; no ordena por parecido ni tolera errores de digitación.

## Herramientas de `cmd/`

```bash
go run ./cmd/test                                       # sandbox contra una base real
go run github.com/celsiainternet/jdb/cmd/install        # go mod tidy + guía de agentes de jdb
go run github.com/celsiainternet/jdb/cmd/create go      # genera proyectos, microservicios, modelos y RPC
```

## Desarrollo

```bash
gofmt -w . && go build ./... && go vet ./...
go test ./drivers/...                          # SQLite en memoria + SQL de postgres
go test ./drivers/sqlite -run TestName -v      # un solo test
go test ./cmd/create/...                       # plantillas del generador
```

Los tests de `drivers/oracle` corren solo si `ORA_TEST_HOST` (+ `ORA_TEST_PORT`, `ORA_TEST_USER`, `ORA_TEST_PASSWORD`, `ORA_TEST_SERVICE`) apunta a un servidor; si no, se saltan.

## Cambios sin publicar

Cambios en el árbol de trabajo posteriores al último tag publicado (el de la insignia de versión), sin commit al 2026-11-17. Al publicar la versión, mueve esta lista a las notas del tag:

- Búsqueda: al cambiar los campos de `DefineSearch`, `Init()` borra los índices `<tabla>_search_<hash>` anteriores de la tabla (sin campos, todos). Los campos que mandan son los de `DefineSearch` del arranque actual, no los guardados.
- Modelos: `Model.Load()` deja de ser público (pasa a `load`, lo usa `Init()`); `Model.Save()` ya no depende de `Model.UseCore` y guarda la definición cuando el modelo `core` está iniciado; al cargar una definición guardada se recuperan también los rollups que el código no define. `Model` serializa columnas, llaves, índices, únicos y requeridos (antes se omitían), y `Index` serializa `column` y `sorted`.
- Comentarios de documentación en formato Elvis (`@param` en una línea) en `jdb`, `drivers/postgres` y las plantillas de `cmd/create`.

## Versionado

```bash
git add . && git commit -m 'Update version' && ./version.sh --r   # parche  X.Y.Z+1
./version.sh --n                                                  # menor    X.Y+1.0
./version.sh --m                                                  # mayor    X+1.0.0
```

`version.sh` actualiza la versión en este README y publica el tag.
