# JDB - Go Database Library

[![Go Version](https://img.shields.io/badge/Go-1.23.0+-blue.svg)](https://golang.org)
[![Version](https://img.shields.io/badge/Version-v1.0.95-orange.svg)](https://github.com/celsiainternet/jdb/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![GitHub](https://img.shields.io/badge/GitHub-celsiainternet%2Fjdb-black.svg)](https://github.com/celsiainternet/jdb)

JDB (`github.com/celsiainternet/jdb`) es una librería de Go que proporciona una interfaz unificada sobre PostgreSQL, MySQL y SQLite: definición declarativa de modelos, un lenguaje de consulta (`Ql`) y de comandos (`Command`) fluido, transacciones, y un pequeño conjunto de paquetes de features (autorización, configuración, inbox) construidos sobre el mismo patrón.

No es una aplicación: `cmd/` no tiene un `main.go` de nivel superior, sino tres binarios independientes (ver [Herramientas de `cmd/`](#herramientas-de-cmd)). Depende de [`github.com/celsiainternet/elvis`](https://github.com/celsiainternet/elvis) para utilidades, logging, tipos JSON (`et.Json`) y eventos.

## Características

### Multi-driver

- **PostgreSQL** (`drivers/postgres`) — driver nativo sobre `github.com/lib/pq`.
- **MySQL** (`drivers/mysql`) — sobre `github.com/go-sql-driver/mysql`.
- **SQLite** (`drivers/sqlite`) — sobre `modernc.org/sqlite`, para uso embebido.

Cada driver se registra a sí mismo en su propio `init()` vía `jdb.Register(...)` y activa sus valores por defecto desde variables de entorno; se habilita con un import en blanco (`import _ "github.com/celsiainternet/jdb/drivers/postgres"`).

### API declarativa y fluida

- **Modelos declarativos**: columnas, llaves, índices, relaciones, rollups y campos especiales definidos con métodos `Define*` sobre `*Model`.
- **`Ql`** (`jdb/ql.go`): consultas de lectura inmutables construidas encadenando métodos (`Where`, `And`, `Or`, `Eq`, `Like`, `In`, `Between`, joins, orden, límites) y ejecutadas con `All()`, `One()`, `First(n)`, `Counted()`, `ItExists()` (o sus variantes `*Tx` dentro de una transacción).
- **`Command`** (`jdb/command.go`): operaciones de escritura inmutables (`Insert`, `Update`, `Delete`, `Upsert`, `Bulk`) construidas igual que `Ql` y ejecutadas con `Exec()` / `One()` (o `ExecTx(tx)` dentro de una transacción).
- **Hooks de ciclo de vida**: `BeforeInsert`, `BeforeUpdate`, `BeforeDelete`, `BeforeInsertOrUpdate`, `AfterInsert`, `AfterUpdate`, `AfterDelete`, `AfterInsertOrUpdate` — cada uno recibe `func(tx *jdb.Tx, data et.Json) error`.
- **Eventos por modelo**: `model.On(channel, handler)` / `model.Emit(channel, data)`, integrados con `elvis/event`.

### Transacciones

Soporte para transacciones vía `jdb.NewTx()` + `tx.Begin(db.Db)`, con `Commit()` / `Rollback()`, y las variantes `*Tx` de `Ql`/`Command` para ejecutar consultas y comandos dentro de la misma transacción.

### HTTP handlers

`jdb.go` expone cuatro `http.HandlerFunc` listos para montar en cualquier router (Chi u otro):

- `jdb.ModelDefine` — describe un modelo/schema/DB.
- `jdb.ModelQuery` — ejecuta un `Ql` a partir de un body JSON.
- `jdb.ModelCommand` — ejecuta uno o más `Command` a partir de un body JSON.
- `jdb.ModelDescribe` — describe un objeto por tipo + nombre.

### Paquetes de features (`instances`, `authorization`, `config`, `inbox`)

Cuatro paquetes que siguen el mismo patrón: un singleton de paquete (no exportado), poblado una única vez por una función `Load(db, schema, ...)` que no hace nada si ya fue cargado, y construido por una función `Define(...)` que define el schema/modelo contra `jdb` de forma idempotente. Cada uno posee exactamente una tabla/modelo y agrega comportamiento propio (CRUD, handlers HTTP, eventos) encima — por ejemplo, `authorization.Load` también se conecta a `elvis/middleware.SetAuthorizationStore`. Ver [`authorization/authorization.go`](authorization/authorization.go) como referencia de implementación.

### Generación de datos y utilidades

- `model.New(fields ...string)` — genera un `et.Json` con los valores por defecto de las columnas del modelo (útil para plantillas/formularios).
- `jdb.GetSeries(model, field)` — obtiene el siguiente valor de una secuencia interna del `core`.

## Instalación

```bash
go get github.com/celsiainternet/jdb@v1.0.95
```

`jdb` depende de `elvis`, así que normalmente también necesitarás:

```bash
go get github.com/celsiainternet/elvis@v1.1.309
```

### Workspace local (`elvis` + `jdb` en desarrollo conjunto)

Si estás desarrollando `jdb` junto con `elvis` en este mismo workspace (ver `../CLAUDE.md`), enlázalos con un `go.work` en la raíz del workspace en lugar de depender de la versión publicada:

```bash
go work init ./elvis
go work use ./elvis
go work use ./jdb
```

## Configuración

### Variables de entorno

| Variable      | Default     | Propósito                                                     |
| ------------- | ----------- | ------------------------------------------------------------- |
| `DB_NAME`     | `jdb`       | Nombre de la base de datos                                    |
| `DB_DRIVER`   | —           | `postgres`, `mysql` o `sqlite`                                |
| `DB_HOST`     | `localhost` | Host de la base de datos                                      |
| `DB_PORT`     | `5432`      | Puerto de la base de datos                                    |
| `DB_USER`     | `admin`     | Usuario de la base de datos                                   |
| `DB_PASSWORD` | `admin`     | Contraseña de la base de datos                                |
| `APP_NAME`    | `jdb`       | Nombre de la aplicación (usado en el connection string de PG) |
| `NODE_ID`     | `0`         | ID de nodo para generación de IDs distribuidos                |
| `DEBUG`       | `false`     | Habilita logging de debug                                     |
| `DB_VERSION`  | `13`        | Versión del servidor PostgreSQL                               |

## Uso básico

### Conexión a la base de datos

```go
package main

import (
    "fmt"

    jdb "github.com/celsiainternet/jdb/jdb"
    "github.com/celsiainternet/jdb/drivers/postgres" // importarlo ya registra el driver en su init()
)

func main() {
    params := jdb.ConnectParams{
        Driver:   "postgres",
        Name:     "myapp",
        UserCore: true, // crea el schema "core" con metadatos internos
        NodeId:   1,
        IsDebug:  true,
        Params: &postgres.Connection{
            Host:     "localhost",
            Port:     5432,
            Username: "postgres",
            Password: "password",
            Database: "myapp",
            App:      "myapp",
        },
    }

    db, err := jdb.ConnectTo(params)
    if err != nil {
        panic(err)
    }
    defer db.Disconected()

    fmt.Println("Conectado a:", db.Name)
}
```

También puedes conectar directamente desde las variables de entorno (usa los defaults que registra el driver en su `init()`):

```go
db, err := jdb.Load()
```

### Definición de modelos

```go
schema := jdb.NewSchema(db, "public")

user := jdb.NewModel(schema, "users", 1)
user.DefineColumn("name", jdb.TypeDataText)
user.DefineColumn("email", jdb.TypeDataText)
user.DefineColumn("age", jdb.TypeDataInt)
user.DefineRequired("name", "email")
user.DefineUnique("email")

// Campos especiales del sistema
user.DefineCreatedAtField() // fecha de creación
user.DefineUpdatedAtField() // fecha de actualización
user.DefineStatusField()    // estado (activo/archivado/etc.)
user.DefineSystemKeyField() // clave del sistema
user.DefineIndexField()     // índice
user.DefineSourceField()    // origen

// Crea/registra el modelo en la base de datos
if err := user.Init(); err != nil {
    panic(err)
}
```

### Operaciones CRUD

```go
import "github.com/celsiainternet/elvis/et"

// Insertar
item, err := user.Insert(et.Json{
    "name":  "Juan Pérez",
    "email": "juan@example.com",
    "age":   30,
}).One()

// Consultar
items, err := user.
    Where("active").Eq(true).
    All()

// Actualizar
result, err := user.
    Update(et.Json{"age": 31}).
    Where("id").Eq("user123").
    Exec()

// Eliminar
result, err := user.
    Delete("id").Eq("user123").
    Exec()
```

### Bulk insert

```go
result, err := user.Bulk([]et.Json{
    {"name": "Ana García", "email": "ana@example.com", "age": 25},
    {"name": "Carlos López", "email": "carlos@example.com", "age": 35},
    {"name": "María Rodríguez", "email": "maria@example.com", "age": 28},
}).Exec()
```

### Transacciones

```go
tx := jdb.NewTx()
if err := tx.Begin(db.Db); err != nil {
    panic(err)
}
defer tx.Rollback()

_, err := user.
    Insert(et.Json{"name": "Usuario Transaccional", "email": "tx@example.com"}).
    ExecTx(tx)
if err != nil {
    panic(err)
}

if err := tx.Commit(); err != nil {
    panic(err)
}
```

### Consultas con JOIN, orden y paginación

```go
items, err := user.
    Join(profile, "id", "=", "user_id").
    Where("users.active").Eq(true).
    And("profiles.verified").Eq(true).
    OrderByDesc("users.created_at").
    First(10)

// Paginación (página, filas por página)
list, err := user.Where("active").Eq(true).List(1, 20)
```

### Hooks (before/after)

```go
user.BeforeInsert(func(tx *jdb.Tx, data et.Json) error {
    fmt.Println("Insertando usuario:", data)
    return nil
})

user.AfterUpdate(func(tx *jdb.Tx, data et.Json) error {
    fmt.Println("Usuario actualizado:", data)
    return nil
})
```

### Eventos por modelo

```go
user.On("custom_event", func(msg event.EvenMessage) {
    fmt.Println("Evento personalizado:", msg)
})

user.Emit("custom_event", et.Json{"user_id": "123"})
```

### Campos especiales

```go
// Texto completo
user.DefineFullText("spanish", []string{"name", "description"})

// Relación (uno-a-muchos hacia el modelo actual)
user.DefineRelation("profile", "profiles", map[string]string{"user_id": "id"}, 1)

// Rollup (agregación desde otra tabla)
user.DefineRollup("total_orders", "orders", map[string]string{"user_id": "id"}, []string{"amount"})

// Objeto embebido (uno-a-uno)
user.DefineObject("address", "addresses", map[string]string{"user_id": "id"}, []string{"street", "city", "country"})
```

### Generación de datos de prueba

```go
// Valores por defecto para todas las columnas
data := user.New()

// Solo para las columnas indicadas
data := user.New("name", "email", "age")
```

### HTTP handlers

```go
import (
    "github.com/go-chi/chi/v5"
    jdb "github.com/celsiainternet/jdb/jdb"
)

r := chi.NewRouter()
r.Post("/model/query", jdb.ModelQuery)
r.Post("/model/command", jdb.ModelCommand)
r.Get("/model/define", jdb.ModelDefine)
r.Get("/model/describe", jdb.ModelDescribe)
```

## Estructura del proyecto

```
jdb/
├── jdb/                 # Paquete principal
│   ├── database.go      # *DB: conexión, schemas, modelos
│   ├── schema.go        # *Schema: namespace dentro de una DB
│   ├── model.go         # *Model: definición de tabla
│   ├── model-define.go  # Métodos Define* (columnas, llaves, relaciones...)
│   ├── column.go        # *Column y tipos de dato (TypeData*)
│   ├── command*.go      # *Command: Insert/Update/Delete/Upsert/Bulk
│   ├── ql*.go           # *Ql: consultas (where, joins, orden, límites)
│   ├── tx.go            # *Tx: transacciones
│   ├── drivers.go        # interfaz Driver + registro de drivers
│   └── jdb.go            # singleton global + HTTP handlers
├── drivers/
│   ├── postgres/
│   ├── mysql/
│   └── sqlite/
├── instances/           # paquete de feature: singleton CRUD genérico
├── authorization/       # paquete de feature: modelo de sesión/autorización
├── config/              # paquete de feature: configuración en runtime
├── inbox/               # paquete de feature: bandeja de mensajes
└── cmd/
    ├── test/            # sandbox manual contra una DB real (no es un ejemplo estable)
    ├── install/         # instala dependencias de terceros vía `go get`
    └── create/          # CLI (Cobra) que genera proyectos/modelos desde plantillas
```

## Drivers soportados

### PostgreSQL

```go
import _ "github.com/celsiainternet/jdb/drivers/postgres"

params := jdb.ConnectParams{
    Driver: "postgres",
    Params: &postgres.Connection{
        Host:     "localhost",
        Port:     5432,
        Username: "postgres",
        Password: "password",
        Database: "myapp",
        App:      "myapp",
    },
}
```

### MySQL

```go
import _ "github.com/celsiainternet/jdb/drivers/mysql"

params := jdb.ConnectParams{
    Driver: "mysql",
    Params: &mysql.Connection{
        Host:     "localhost",
        Port:     3306,
        Username: "root",
        Password: "password",
        Database: "myapp",
    },
}
```

### SQLite

```go
import _ "github.com/celsiainternet/jdb/drivers/sqlite"

params := jdb.ConnectParams{
    Driver: "sqlite",
    Params: &sqlite.Connection{
        Database: "./data.db",
    },
}
```

## Herramientas de `cmd/`

`cmd/` no contiene la librería en sí, sino tres binarios independientes:

```bash
# Sandbox de desarrollo: conecta a una DB real vía variables de entorno
gofmt -w . && go run --race ./cmd/test

# Instala un conjunto fijo de dependencias de terceros (bootstrap de un nuevo consumidor)
go run github.com/celsiainternet/jdb/cmd/install

# CLI para scaffolding de nuevos proyectos/modelos de microservicio
go run github.com/celsiainternet/jdb/cmd/create go
```

No hay archivos de test (`*_test.go`) en este repositorio.

## Gestión de versiones (`version.sh`)

```bash
# Incrementar versión de parche (X.Y.Z+1), etiquetar y hacer push de tags
git add . && git commit -m 'Update version' && ./version.sh --r

# Incrementar versión menor (X.Y+1.0)
./version.sh --n

# Incrementar versión mayor (X+1.0.0)
./version.sh --m

# Ayuda
./version.sh --h
```

`version.sh` reescribe la versión anterior en este `README.md` (badge de versión) y crea/empuja el tag de Git correspondiente — sigue el estándar semántico (SemVer).
