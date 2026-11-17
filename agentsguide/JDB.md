## Catálogo de patrones de jdb

Aplica a proyectos cuyo `go.mod` requiere `github.com/celsiainternet/jdb`: en ellos, jdb es el ORM estándar para modelos nuevos (ver `MARCO.md` de Elvis). Firmas verificadas contra jdb v1.0.115 (revisado el 2026-11-17). Para otra versión, comprueba el símbolo:

```bash
grep -rn "func (s \*Ql) <Nombre>" "$(go env GOMODCACHE)/github.com/celsiainternet/jdb@$(go list -m -f '{{.Version}}' github.com/celsiainternet/jdb)/jdb"
```

No confundas este paquete (`github.com/celsiainternet/jdb/jdb`) con `github.com/celsiainternet/elvis/jdb`, que es otro paquete con otra API.

### 1. Conexión

- `jdb.Load()`: driver de `DB_DRIVER` (default `postgres`) y parámetros de las variables de entorno.
- Importa el driver en blanco: `_ "github.com/celsiainternet/jdb/drivers/postgres"` (o `sqlite`, `oracle`). No hay driver MySQL.
- `jdb.Load()` crea las tablas del core (series, modelos) por defecto; `USE_CORE=false` lo evita. Con `jdb.ConnectTo` se activan con `ConnectParams{UseCore: true, …}`.

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
- Otros: `DefineAtribute` (clave dentro de la fuente JSON), `DefineRequired`, `DefineUnique`, `DefineForeignKey`, `DefineRelation`, `DefineRollup`, `DefineObject`, `DefineDetail`, `DefineSearch`, `DefineHidden`, `DefineIntegrity` (rechaza claves no declaradas).
- `DefineHidden` oculta **columnas** en las consultas sin select explícito; un atributo oculto sigue saliendo dentro de `_data`.
- `DefineCalc` / `DefineCalcTx` **no se ejecutan**: ninguna consulta llama a la función. Calcula el valor en tu código después de leer, o usa un rollup.
- `Init()` crea tabla, índices y llaves solo si la tabla no existe: en una tabla existente no agrega columnas ni índices nuevos.
- Hooks del modelo y del comando: `BeforeInsert…` / `AfterInsert…` y equivalentes de update y delete. Corren por fila: los del modelo, luego los del comando, la escritura y los `After*`. Un error en un hook detiene el comando.

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
- Transacciones: `tx, err := db.Begin()`, las variantes `ExecTx(tx)`, `AllTx(tx)`, `OneTx(tx)`, `ItExistsTx(tx)` y al final `return tx.End(err)` (revierte con error, confirma sin él). El `*jdb.Tx` también llega a los hooks: dentro de un hook, usa siempre el `tx` recibido para que la escritura quede en la misma transacción.
- Sin filas, `One()` y los comandos devuelven `Ok == false` **sin error**. `jdb.ErrNotFound`, `jdb.ErrNotInserted` y los demás `ErrNot*` no los devuelve jdb: devuélvelos tú desde un hook y compáralos después.
- Listados desde JSON: `jdb.RunQuery(model, query)` con `wheres`, `search` (texto), `selects`, `order_by`, `page` y `rows` (100 por defecto, máximo 1000). Ver `jsql.md` del repo jdb.
- Un nombre de campo que no resuelve se convierte en **literal de texto** (`'age'`), no en columna: escribe los nombres exactos. Rutas JSON: `_data.a.b` o `_data->'a'->>'b'`.
- `Ql` y `Command` son builders mutables: crea uno nuevo por consulta y no lo reutilices después de ejecutarlo.
- Los valores de WHERE se incrustan en el SQL (no son argumentos enlazados): nunca pases datos del usuario a `CALC(…)`, a una agregación ni como nombre de campo.

### 4. Defectos conocidos que debes esquivar

Verificados en jdb v1.0.115 (2026-11-17). Antes de depender de uno, o de corregirlo, revisa la sección "Known defects" del `CLAUDE.md` del repo jdb: cambia con frecuencia.

| Defecto | Cómo esquivarlo |
|---|---|
| `Ql.Join("nombre", …)` con un string se ignora | Pasa el `*Model`, no el nombre. |
| El lado derecho de un `Join` y `Having` son literales, no campos | No dependas de ellos para comparar columnas. |
| `DefineCalc` no se ejecuta | Completa el valor en el código después de la consulta. |
| `RunQuery` ignora en silencio un campo desconocido en `wheres` o `selects`; `notIn` devuelve 0 filas y `between` todas; `search` como objeto devuelve 0 filas | Valida los nombres antes de llamar; usa `search` como texto; para rangos, dos condiciones (`moreEq` y `lessEq`). |
| `Bulk` devuelve solo la última fila | Si necesitas todas, inserta una por una dentro de una transacción. |
| `DB_RECORD_LIMIT` no limita nada | Pagina siempre con `List` o `Rows`. |
| `GetSchema("sinPunto")`, `From(<nombre desconocido>)` y `Detail(<campo desconocido>)` hacen panic | Valida el nombre antes de llamarlos. |
| En la consulta JSON (`Ql.Query`), `limit > 1` devuelve una sola fila o hace panic | Pasa `limit: 0` para todas las filas, o usa `RunQuery`. |
| `ModelCommand` (HTTP): `update` y `delete` no cambian nada | Usa la API fluida (`Update(...).Where(...)`, `Delete(...)`). |

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
