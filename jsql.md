# JSQL: consultas y comandos en JSON con jdb

jdb interpreta definiciones en JSON y las convierte en SQL. Esta guía muestra cada formato con un ejemplo completo, el SQL que genera y lo que devuelve.

Todos los ejemplos se ejecutaron el 2026-11-04 contra SQLite en memoria, y el SQL de Postgres se generó con el driver `postgres` sin base de datos. Las limitaciones de cada formato están marcadas con ⚠️ y resumidas al final.

## Los tres formatos

| Formato | Entrada en Go | Para qué |
| --- | --- | --- |
| **Query paginado** (recomendado) | `jdb.RunQuery(model, query, rollups...)` | Listados con filtros, orden y paginación. Devuelve `et.List` con el total. Es el que usa vcore (`octopus/api/internal/models/vcore/query.go`). |
| **Query libre** | `model.Query(query)`, `ql.Query(query)`, `db.JQuery(query)` y el handler HTTP `jdb.ModelQuery` | Consultas con `select`, `group_by` y un `where` en forma de objeto. Tiene varias limitaciones (ver abajo). |
| **Comandos** | handler HTTP `jdb.ModelCommand` | Insertar registros desde JSON. `update`, `delete` y `upsert` no funcionan hoy. |

Además, `jdb.ModelDefine` y `jdb.ModelDescribe` describen modelos en JSON. El handler `jdb.JSQL` solo devuelve el mismo body que recibe.

## Modelo de ejemplo

Todos los ejemplos usan este modelo:

```go
schema := jdb.NewSchema(db, "crm")

customers := jdb.NewModel(schema, "customers", 1)
customers.DefineModel()                            // _id, _data, created_at, updated_at, _status, idx, jdbid
customers.DefineColumn("name", jdb.TypeDataText)
customers.DefineColumn("kind", jdb.TypeDataText)
customers.DefineColumn("age", jdb.TypeDataInt)
customers.DefineColumn("email", jdb.TypeDataText)
customers.DefineAtribute("city", jdb.TypeDataText) // atributo: vive dentro de _data
if err := customers.Init(); err != nil {
	return err
}
```

Datos:

| `_id` | `name` | `kind` | `age` | `city` | `email` |
| --- | --- | --- | --- | --- | --- |
| c1 | Ana | lead | 30 | Cali | ana@x.co |
| c2 | Beto | client | 17 | Palmira | |
| c3 | Carla | lead | 45 | Cali | |
| c4 | Dario | client | 52 | Buga | d@x.co |
| c5 | Elena | prospect | 22 | Cali | |

## Condiciones

Una condición tiene siempre la forma `{ "campo": { "operador": valor } }`.

```json
{ "kind": { "eq": "lead" } }
{ "age":  { "moreEq": 18 } }
{ "kind": { "in": ["lead", "prospect"] } }
{ "city": { "eq": "Cali" } }
```

### Operadores

| Operador | SQL | Valor | `RunQuery` | `Query` |
| --- | --- | --- | --- | --- |
| `eq` | `=` | valor | ✅ | ✅ |
| `neg` | `!=` | valor | ✅ | ✅ |
| `more` | `>` | valor | ✅ | ✅ |
| `less` | `<` | valor | ✅ | ✅ |
| `moreEq` | `>=` | valor | ✅ | ✅ |
| `lessEq` | `<=` | valor | ✅ | ✅ |
| `like` | `LIKE` (Postgres: `ILIKE`, sin distinguir mayúsculas) | patrón con `%` | ✅ | ✅ |
| `in` | `IN (...)` | lista | ✅ | ✅ |
| `notIn` | `NOT IN (...)` | lista | ⚠️ devuelve 0 filas | ✅ |
| `between` | `BETWEEN a AND b` | lista de 2 | ⚠️ se ignora y devuelve todo | ✅ |
| `isNull` | `IS NULL` | cualquiera (`true`) | ✅ | ✅ |
| `notNull` | `IS NOT NULL` | cualquiera (`true`) | ✅ | ✅ |
| `search` | contiene el valor: Postgres `ILIKE '%' \|\| v \|\| '%'`, SQLite `LIKE`, Oracle `UPPER` + `LIKE` | texto | ✅ | ✅ |

Los nombres distinguen mayúsculas: `moreEq`, no `moreeq`.

### Qué se puede usar como campo

- **Una columna:** `"kind"`.
- **Un atributo** declarado con `DefineAtribute`: `"city"`. jdb lo busca dentro de `_data` y genera `COALESCE(A._data#>>'{city}', '')` en Postgres.
- **Una ruta JSON** dentro de una columna objeto o de `_data`: `"_data.city"`, `"_data.sync.order_id"` o la forma Postgres `"_data->'sync'->>'order_id'"`. Solo se aceptan claves con `[A-Za-z0-9_-]`. Solo funciona en el **query libre** (`Query`): `RunQuery` ignora el campo porque no es una columna del modelo.

> ⚠️ **`isNull` y los textos vacíos.** Las columnas de texto se guardan con `''` por defecto, no con `NULL`. `{ "email": { "isNull": true } }` no encuentra a Beto. Para "sin email" usa `{ "email": { "eq": "" } }`.

> ⚠️ **Un operador por condición.** `{ "age": { "moreEq": 18, "lessEq": 40 } }` no genera un rango: los dos operadores escriben sobre la misma condición y queda uno al azar. Usa dos condiciones o `between` en el query libre.

---

## 1. Query paginado: `RunQuery` (recomendado)

### Ejemplo completo

```json
{
  "wheres": [
    { "kind": { "in": ["lead", "prospect"] } },
    { "and":  { "age": { "moreEq": 18 } } },
    { "and":  { "name": { "like": "%a%" } } },
    { "or":   { "city": { "eq": "Buga" } } }
  ],
  "selects": ["_id", "name", "age", "city"],
  "order_by": { "desc": ["age"], "asc": ["name"] },
  "page": 1,
  "rows": 20
}
```

```go
list, err := jdb.RunQuery(customers, query)
```

### Llaves

| Llave | Tipo | Comportamiento |
| --- | --- | --- |
| `wheres` | lista de condiciones | Se aplican **en orden**. El primer elemento abre el `WHERE` y los siguientes sin conector se unen con `AND`. |
| `{ "and": {...} }` / `{ "or": {...} }` | elemento de `wheres` | Une la condición con `AND` u `OR`. Va **un objeto** con una condición, no una lista. |
| `search` | texto | Búsqueda en los campos de `DefineSearch` del modelo: una condición `AND` por palabra (hasta 10; vacío no filtra). En Postgres sin tildes ni mayúsculas y con índice de trigramas. Si el modelo no declaró campos, no devuelve nada. Ver README › Búsqueda. |
| `selects` | lista de nombres | Columnas a devolver. Si se omite, devuelve el registro completo con los atributos de `_data`. |
| `order_by` | `{ "asc": [...], "desc": [...] }` | Cada columna se valida: un nombre que no existe devuelve el error `Columna no válida en el query (x)`. Aplica primero `asc` y luego `desc`. |
| `page` | entero | Página, desde 1. Por defecto 1. |
| `rows` | entero | Filas por página. Por defecto **100** y máximo **1000**. `limit` se acepta como alias. |
| `join`, `left_join`, `right_join`, `full_join` | lista | ⚠️ Se leen pero hoy no generan el `JOIN` (ver limitaciones). |

### Precedencia

No hay paréntesis: aplica la precedencia de SQL, donde `AND` va antes que `OR`. El ejemplo anterior genera:

```sql
WHERE A.kind IN ('lead', 'prospect')
  AND A.age >= 18
  AND A.name ILIKE '%a%'
   OR COALESCE(A._data#>>'{city}', '') = 'Buga'
```

y se lee como `(kind IN (...) AND age >= 18 AND name LIKE '%a%') OR city = 'Buga'`.

### Respuesta (`et.List`)

Con `{ "order_by": { "asc": ["name"] }, "page": 2, "rows": 2 }`:

```json
{
  "rows": 2,
  "all": 5,
  "count": 2,
  "page": 2,
  "start": 3,
  "end": 4,
  "result": [
    { "_id": "c3", "name": "Carla", "kind": "lead",   "age": 45, "city": "Cali", "email": "",       "...": "..." },
    { "_id": "c4", "name": "Dario", "kind": "client", "age": 52, "city": "Buga", "email": "d@x.co", "...": "..." }
  ]
}
```

- `all` es el total que cumple los filtros, sin paginar.
- `count` es lo que trae esta página.
- `start` y `end` son las posiciones de la página, contando desde 1.

### Rollups por consulta

Los rollups definidos en datos se pasan como argumento extra y se resuelven en cada fila:

```go
rollup := jdb.NewRollup("team_name", teams, map[string]string{"team_id": jdb.KEY}, []string{"name"}, jdb.StrToShowRollup("attribs")) // teams es *jdb.Model; "objects" devuelve un objeto
list, err := jdb.RunQuery(roles, query, rollup)
```

> ⚠️ **Un campo que no existe se ignora sin error.** `{ "wheres": [{ "nope": { "eq": "x" } }] }` devuelve **todos** los registros. Valida los nombres de campo antes de llamar a `RunQuery` si vienen del usuario.

---

## 2. Query libre: `Query`, `JQuery` y `ModelQuery`

### Ejemplo completo

```json
{
  "from": "crm.customers",
  "select": ["name", "age"],
  "where": {
    "kind": { "eq": "lead" },
    "and": [
      { "age":  { "moreEq": 18 } },
      { "city": { "like": "%Cal%" } }
    ]
  },
  "order_by": { "desc": ["age"] },
  "limit": 0,
  "debug": false
}
```

```go
result, err := customers.Query(query) // o jdb.From(customers).Query(query)
result, err := db.JQuery(query)       // usa "from" para buscar el modelo
// HTTP: r.Post("/query", jdb.ModelQuery)  (body = el JSON de arriba)
```

SQL generado en Postgres:

```sql
SELECT
jsonb_build_object(
'name', A.name,
'age', A.age) AS result
FROM crm.customers AS A
WHERE A.kind='lead'
AND A.age>=18
AND COALESCE(A._data#>>'{city}', '') ILIKE '%Cal%'
ORDER BY
A.age DESC;
```

Respuesta:

```json
{ "ok": true, "count": 2, "result": [ { "age": 45, "name": "Carla" }, { "age": 30, "name": "Ana" } ] }
```

### Llaves

| Llave | Tipo | Comportamiento |
| --- | --- | --- |
| `from` | `"schema.modelo"` o `"modelo"` | Solo lo leen `db.JQuery` y `ModelQuery`. `model.Query` usa su propio modelo. |
| `select` | lista de nombres | Columnas a devolver. Si se omite, devuelve el registro completo con los atributos de `_data`. |
| `where` | **objeto** | Llaves sueltas `{ "campo": {...} }` y grupos `"and": [...]` / `"or": [...]` (listas de condiciones, sin importar mayúsculas). Las llaves se resuelven a columnas, atributos o rutas JSON. |
| `group_by` | lista de nombres | `GROUP BY`. Úsalo con `select` de las mismas columnas. |
| `order_by` | `{ "asc": [...], "desc": [...] }` | Igual que en `RunQuery`, pero sin validar nombres. |
| `limit` | entero | `0`: todas las filas, con respuesta `{ ok, count, result: [...] }`. `1`: una fila, con respuesta `{ ok, result: {...} }`. Si se omite vale **1000**, que cae en el caso `limit > 1` (ver abajo). |
| `debug` | bool | Imprime el JSON recibido. |
| `page`, `having`, `join` | — | ⚠️ No funcionan (ver abajo). |

### `group_by`

```json
{ "select": ["kind"], "group_by": ["kind"], "limit": 0 }
```

```sql
SELECT jsonb_build_object('kind', A.kind) AS result
FROM crm.customers AS A
GROUP BY A.kind;
```

Devuelve `[{ "kind": "client" }, { "kind": "lead" }, { "kind": "prospect" }]`. El JSON no admite agregaciones (`COUNT`, `SUM`, …): para eso usa la API fluida con `jdb.COUNT(...)`.

### Limitaciones del query libre

> ⚠️ **`limit` mayor que 1 no sirve.** Con `"limit": N` (o sin `limit`, que vale 1000), jdb consulta con `LIMIT N` y devuelve solo la fila en la posición `N`, que no existe. Si hay menos de `N` filas, la respuesta llega vacía. Si hay `N` o más, **el proceso cae** con `panic: index out of range [N] with length N` (verificado con `"limit": 2`). Envía siempre `"limit": 0` o `"limit": 1`. Para paginar usa `RunQuery`.

> ⚠️ **`page` sin `limit` devuelve vacío.** `{ "page": 2, "limit": 0 }` genera `LIMIT 0 OFFSET 0` y no trae filas.

> ⚠️ **No mezcles `or` con otras llaves en el mismo `where`.** El `where` es un objeto y Go lo recorre en orden aleatorio. `{ "kind": {...}, "and": [...], "or": [...] }` generó en ejecuciones distintas `… AND … OR …` y `… AND … AND …`: el `OR` se pierde cuando queda de primero. `and` con llaves sueltas es seguro, porque `AND` es conmutativo. Si necesitas `OR`, usa `RunQuery`, que recibe una lista ordenada.

> ⚠️ **`having` genera SQL inválido** (`HAVING '{"kind":{"eq":"lead"}}'`) y no trae filas.

> ⚠️ **`join` se ignora.** `[{ "with": "crm.contacts", "field": "customer_id", "operator": "eq", "value": "_id" }]` produce un `SELECT` sin `JOIN`.

---

## 3. Comandos: `ModelCommand`

El body es un objeto cuyas llaves son nombres de modelo (`"schema.modelo"`); cada una lleva una acción. Responde `{ ok, count, result }` con los registros afectados.

### Insertar

```json
{
  "crm.customers": {
    "insert": { "_id": "c9", "name": "Zoe", "kind": "lead", "age": 28, "city": "Tulua" }
  }
}
```

```json
{ "ok": true, "count": 1, "result": [ { "_id": "c9", "name": "Zoe", "kind": "lead", "age": 28, "city": "Tulua", "email": "", "...": "..." } ] }
```

`city` no es columna, así que se guarda en `_data`. Si el modelo no es `Integrity`, cualquier llave no declarada también va a `_data`.

### Insertar varios

```json
{
  "crm.customers": {
    "bulk": [
      { "_id": "b1", "name": "B1" },
      { "_id": "b2", "name": "B2" }
    ]
  }
}
```

Inserta todos los registros, pero `result` solo trae el último.

### `update`, `delete` y `upsert`: no usar hoy

Estos formatos los acepta el handler, pero **no hacen lo que dicen**:

```json
{ "crm.customers": { "update": { "age": 29 }, "where": { "_id": { "eq": "c9" } } } }
{ "crm.customers": { "delete": true,          "where": { "_id": { "eq": "b1" } } } }
{ "crm.customers": { "upsert": { "_id": "c9", "name": "Zoe2" }, "where": { "_id": { "eq": "c9" } } } }
```

| Acción | Resultado verificado |
| --- | --- |
| `update` | HTTP 200 con `{ "ok": false, "count": 0 }` y no cambia nada. Las llaves del `where` de un comando no se resuelven a columnas y llegan al SQL como texto (`'_id' = 'c9'`). |
| `delete` | Igual que `update`: no borra nada. |
| `upsert` | Intenta insertar y falla con `UNIQUE constraint failed` cuando el registro ya existe. |

Para actualizar o borrar usa la API fluida:

```go
customers.Update(et.Json{"age": 29}).Where("_id").Eq("c9").Exec()
customers.Delete("_id").Eq("b1").Exec()
```

---

## 4. Describir modelos

```json
// jdb.ModelDefine
{ "params": { "type": "model", "name": "crm.customers" } }   // type: model | schema | db

// jdb.ModelDescribe
{ "kind": "field", "name": "customers.city" }                  // kind: db | schema | model | field
```

Devuelven el `Describe()` del objeto: columnas con `name`, `type_column`, `type_data`, `hidden`, etc.

---

## Resumen de limitaciones (verificado 2026-11-04)

| Formato | Limitación |
| --- | --- |
| `RunQuery` | Un campo inexistente en `wheres` se ignora y la consulta devuelve todo. |
| `RunQuery` | `between` se ignora (devuelve todo) y `notIn` devuelve 0 filas. |
| `RunQuery` | Las rutas JSON (`_data.x`) no se aceptan como campo. |
| `RunQuery` / `Query` | `join` no genera el `JOIN`. |
| `Query` | `limit` mayor que 1, o sin `limit`: respuesta vacía, o caída del proceso si hay al menos `limit` filas. |
| `Query` | `page` con `limit: 0` devuelve vacío. |
| `Query` | `or` mezclado con otras llaves genera SQL distinto en cada ejecución. |
| `Query` | `having` genera SQL inválido. |
| `ModelCommand` | `update` y `delete` no afectan filas y `upsert` falla con un registro existente. `bulk` solo devuelve el último registro. |
| Todos | `isNull` no encuentra textos vacíos (`''`). |
| Todos | Los valores del `WHERE` se escriben dentro del SQL, no como parámetros. No uses nombres de campo que vengan del usuario sin validarlos. |
