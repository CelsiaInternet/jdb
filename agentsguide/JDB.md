## Catálogo de patrones de jdb

Aplica a proyectos cuyo `go.mod` requiere `github.com/celsiainternet/jdb`: en ellos, jdb es el ORM estándar para modelos nuevos (ver `MARCO.md` de Elvis). Firmas verificadas contra jdb v1.0.112. Para otra versión, comprueba el símbolo:

```bash
grep -rn "func (s \*Ql) <Nombre>" "$(go env GOMODCACHE)/github.com/celsiainternet/jdb@$(go list -m -f '{{.Version}}' github.com/celsiainternet/jdb)/jdb"
```

No confundas este paquete (`github.com/celsiainternet/jdb/jdb`) con `github.com/celsiainternet/elvis/jdb`, que es otro paquete con otra API.

### 1. Conexión

- `jdb.Load()`: driver de `DB_DRIVER` (default `postgres`) y parámetros de las variables de entorno.
- Importa el driver en blanco: `_ "github.com/celsiainternet/jdb/drivers/postgres"` (o `sqlite`, `oracle`). No hay driver MySQL.
- Para las tablas del core (series, registros) usa `jdb.ConnectTo(jdb.ConnectParams{UseCore: true, …})`: `jdb.Load()` no las activa.

### 2. Modelos

```go
var Blacklist *jdb.Model

/**
* DefineBlacklist: define la tabla de direcciones bloqueadas.
* @param db *jdb.DB
* @return error
**/
func DefineBlacklist(db *jdb.DB) error {
	if Blacklist != nil {
		return nil
	}

	schema := jdb.NewSchema(db, "octopus")
	Blacklist = jdb.NewModel(schema, "blacklist_address", 1)
	Blacklist.DefineModel()
	Blacklist.DefineColumn("address", jdb.TypeDataText)
	Blacklist.DefineColumn("date_unlocked", jdb.TypeDataDateTime)
	Blacklist.DefineIndex(true, "address")

	return Blacklist.Init()
}
```

- `DefineModel()` agrega los campos estándar: fechas de creación y actualización, estado, llave primaria, fuente JSON, llave de sistema e índice. `DefineProjectModel()` agrega además el proyecto. Úsalos en lugar de declarar esos campos a mano.
- Otros: `DefineAtribute` (clave dentro de la fuente JSON), `DefineRequired`, `DefineUnique`, `DefineForeignKey`, `DefineRelation`, `DefineRollup`, `DefineObject`, `DefineDetail`, `DefineCalc`, `DefineHidden`, `DefineIntegrity` (rechaza claves no declaradas).
- Hooks del modelo: `BeforeInsert…` / `AfterInsert…` y equivalentes de update y delete. Revisa §4 antes de depender de ellos.

### 3. Consultas y comandos

```go
items, err := Blacklist.Where("address").Eq(address).And("date_unlocked").IsNull().All()
list, err  := Blacklist.Where(jdb.STATUS_ID).Eq(utility.ACTIVE).List(page, rows)
item, err  := Blacklist.Where(jdb.KEY).Eq(id).One()
items, err := Blacklist.Insert(data).Exec()
items, err := Blacklist.Update(data).Where(jdb.KEY).Eq(id).Exec()
items, err := Blacklist.Delete(jdb.KEY).Eq(id).Exec()
```

- `Where(campo)` abre una condición y el siguiente operador (`Eq`, `Neg`, `In`, `NotIn`, `Like`, `More`, `Less`, `MoreEq`, `LessEq`, `Between`, `IsNull`, `NotNull`) la completa.
- Terminales de lectura: `All()`, `One()`, `Rows(n)` (primeras n, respeta `Page(p)`), `List(page, rows)` (con total, `et.List`). Usa `List` para todo lo que devuelva listas a un cliente.
- **`First(n)` devuelve el elemento en la posición n**, no las primeras n filas. No existen `Limit(n)` ni `Offset(n)`.
- Transacciones: no hay un constructor público. El `*jdb.Tx` llega a los hooks (`BeforeInsert(func(tx *jdb.Tx, data et.Json) error)`, …) o a funciones `…Tx(tx *jdb.Tx, …)` del repositorio, y se propaga con `ExecTx(tx)`, `AllTx(tx)`, `OneTx(tx)`. Dentro de un hook, usa siempre el `tx` recibido para que la escritura quede en la misma transacción.
- Un nombre de campo que no resuelve se convierte en **literal de texto** (`'age'`), no en columna: escribe los nombres exactos. Rutas JSON: `_data.a.b` o `_data->'a'->>'b'`.
- `Ql` y `Command` son builders mutables: crea uno nuevo por consulta y no lo reutilices después de ejecutarlo.
- Los valores de WHERE se incrustan en el SQL (no son argumentos enlazados): nunca pases datos del usuario a `CALC(…)`, a una agregación ni como nombre de campo.

### 4. Defectos conocidos que debes esquivar

Verificados en jdb v1.0.112. Antes de depender de uno, o de corregirlo, revisa la sección "Known defects" del `CLAUDE.md` del repo jdb: cambia con frecuencia.

| Defecto | Cómo esquivarlo |
|---|---|
| `Select("a", "b")` genera `SELECT *` | Selecciona todo y toma los campos del resultado. |
| `Update` / `Delete` con `Where` generan SQL inválido en postgres | Verifica el comportamiento en el driver del proyecto antes de usarlos; si falla, usa la variante que el servicio ya tenga probada. |
| `Upsert` no genera SQL (afecta también `GetSeries` / `SetSeries`) | Consulta con `One()` y luego `Insert` o `Update`. |
| `Ql.Join("nombre", …)` con un string se ignora | Pasa el `*Model`, no el nombre. |
| El lado derecho de un `Join` y `Having` son literales, no campos | No dependas de ellos para comparar columnas. |
| `Model.BeforeInsert*` y el `BeforeUpdate` del comando no se ejecutan | Haz la validación antes de llamar a `Insert` / `Update`. |
| `ExecTx(nil)` (y `Exec()`) hace commit aunque falle un hook | No confíes en que un error de hook revierta la escritura: valida antes de escribir y revisa el error devuelto. |
| `DB_RECORD_LIMIT` no limita nada | Pagina siempre con `List` o `Rows`. |
| `GetSchema("sinPunto")`, `From(<nombre desconocido>)` y `Detail(<campo desconocido>)` hacen panic | Valida el nombre antes de llamarlos. |
| En la consulta JSON (`Ql.Query`), `limit > 1` devuelve una sola fila | Pasa `limit: 0` para todas las filas. |

### 5. Paquetes de funcionalidad

`instances`, `authorization`, `config` e `inbox` siguen la misma forma: un singleton por paquete cargado con `Load(db, schema, …)` idempotente, un modelo por tabla y helpers en `handler.go`. Si creas un paquete parecido, sigue esa forma.

### 6. Variables de entorno

| Variable | Default |
|---|---|
| `DB_DRIVER` | `postgres` (`postgres`, `sqlite` u `oracle`) |
| `DB_NAME`, `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD` | `jdb`, `localhost`, `5432`, `admin`, `admin` |
| `APP_NAME` | `jdb` |
| `DEBUG` | `false` (imprime el SQL generado) |
| `SQLITE_PATH` | `<DB_NAME>.db` (`:memory:` en memoria) |
| `ORA_DB_HOST`, `ORA_DB_PORT`, `ORA_DB_USER`, `ORA_DB_PASSWORD`, `ORA_DB_SERVICE_NAME_ORACLE` | Oracle |
