<!-- Generado por elvis devel con cmd/install. No lo edites: se sobrescribe al reinstalar. -->

## Catálogo de patrones de Elvis

Firmas verificadas contra el código de `github.com/celsiainternet/elvis` (v1.1.321). Si el `go.mod` del proyecto fija otra versión, comprueba el símbolo antes de usarlo:

```bash
grep -rn "func <Nombre>" "$(go env GOMODCACHE)/github.com/celsiainternet/elvis@$(go list -m -f '{{.Version}}' github.com/celsiainternet/elvis)"
```

### 1. Para X, usa Y (no Z)

| Para… | Usa | No uses |
|---|---|---|
| Responder HTTP | `response.ITEM(w, r, status, et.Item)`, `ITEMS(…, et.Items)`, `LIST(…, et.List)`, `JSON(…, any)`, `HTTPError(w, r, status, msg)`, `HTTPAlert(w, r, msg)`, `Unauthorized(w, r)`, `Forbidden(w, r)`, `InternalServerError(w, r, err)`, `Stream(…)` | `json.NewEncoder(w)`, `w.Write` |
| Leer la petición | `response.GetBody(r)` → `et.Json`, `GetArray(r)`, `GetQuery(r)`, `GetParam(r, key)`, `chi.URLParam(r, "id")` | `json.NewDecoder(r.Body)` a un struct ad hoc |
| Registrar rutas | `router.PublicRoute(r, router.Get, "/x", h, PackageName, PackagePath, host)`, `ProtectRoute`, `AuthorizationRoute`, `With(r, method, path, middlewares, h, …)` | `r.Get`, `r.Post`, … directo en `chi` |
| Leer configuración | `envar.GetStr(def, "VAR")`, `GetInt`, `GetInt64`, `GetFloat64`, `GetBool` | `os.Getenv`; las deprecadas `EnvarStr`, `EnvarInt`, `EnvarInt64`, `EnvarNumber`, `EnvarBool` |
| Configuración que también es flag | `envar.SetStr(nombre, def, uso, "VAR")`, `SetInt`, `SetInt64`, `SetBool` | `flag` directo; las deprecadas `SetvarStr`, `SetvarInt`, `SetvarBool` |
| Logs de aplicación | `console.LogK(kind, …)`, `console.LogKF(kind, formato, …)`, `console.Info`, `console.Debug`, `console.AlertE(err)`, `console.Error(err)`, `console.Panic(err)` (arranque) | `fmt.Println`, `log.Print*`; `logs.*` (es interno de las librerías y no publica eventos) |
| Crear un error | `fmt.Errorf(msg.MSG_X, args...)` | Texto en línea; `errors.New("…")` con texto de negocio |
| Mensajes compartidos | constantes de `elvis/msg` (`MSG_*`, `ERR_*`, `RECORD_*`); si no existe, el `msg.go` del paquete | Duplicar en el paquete una constante que ya está en `elvis/msg` |
| Datos dinámicos | `et.Json` (`.Str`, `.Int`, `.Int64`, `.Bool`, `.Json`, `.ArrayJson`, `.Key`), `et.Item`, `et.Items`, `et.List` | `map[string]interface{}`; structs DTO de paso |
| Ids y fechas | `utility.UUID()`, `utility.GenKey(id)` (usa el id si es válido, si no genera uno), `utility.Now()`; validar con `utility.ValidId`, `utility.ValidStr(val, min, notIn)` | `uuid` u otra librería directa; `time.Now().Format(…)` para fechas de registro |
| Caché compartido | `cache.Set(key, val, duración)`, `cache.Get(key, def)`, `cache.Delete(key)`, `cache.DeleteByPrefix(prefijo)`, claves con `cache.GenKey(…)` | Mapas globales en memoria para datos compartidos entre réplicas |
| Caché local del proceso | `mem` | Mapas globales sin mutex |
| Evento asíncrono | `event.Publish(canal, et.Json)`; consumir con `event.Subscribe(canal, f)`, `event.Queue(canal, cola, f)` (reparte entre réplicas) o `event.Stack(canal, f)` (solo suscripciones de reset; se re-registra al reconectar) | Llamadas HTTP en cadena para notificar |
| Evento local (mismo proceso) | `event.On(canal, handler)` / `event.Emit(canal, data)` | Canales Go ad hoc entre paquetes |
| Llamada síncrona a otro servicio | `jrpc.CallJson(método, args)`, `CallItem`, `CallItems`, `CallList`; exponer con `jrpc.Mount(servicio)` | HTTP interno entre servicios; leer la base de otro servicio |
| HTTP saliente | `request.Get/Post/Put/Delete/Patch`, `<Verbo>WithTimeout(…, timeout, defaultValue)`, `<Verbo>WithTls`, `<Verbo>WithTlsTimeout` → `(*request.Body, request.Status)` | `http.Client` directo |
| Reintentos | `resilience` (`RESILIENCE_TOTAL_ATTEMPTS`, `RESILIENCE_TIME_ATTEMPTS`) | Bucles de reintento a mano |
| JWT | `claim.NewToken` / `claim.ValidToken` + `middleware.Autentication` (+ `middleware.Authorization`); datos del usuario con `claim.*(r)` | Parsear el token a mano |
| Tareas programadas | `crontab` | `time.Ticker` en goroutines sueltas |
| Lotes en proceso | `queue.Queue[T]` (despacha por tamaño o timeout) | Buffers a mano |
| Excel | `xls.ReadXls*`, `xls.NewXls(…).ToFile/ToWriter/ToHttp` | `excelize` directo |
| Consultas JSON → SQL | `jquery` | Construir SQL concatenando strings |

### 2. Modelos `linq` (proyectos que no importan jdb)

```go
var Groups *linq.Model

/**
* DefineGroups: define la tabla de grupos.
* @param db *jdb.DB
* @return error
**/
func DefineGroups(db *jdb.DB) error {
	if Groups != nil {
		return nil
	}

	Groups = linq.NewModel(schema, "GROUPS", "Grupos", 1)
	Groups.DefineColum(jdb.KEY, "", "VARCHAR(80)", "-1")
	Groups.DefineColum("name", "", "VARCHAR(250)", "")
	Groups.DefineColum("_data", "", "JSONB", "{}")
	Groups.DefinePrimaryKey([]string{jdb.KEY})
	Groups.DefineIndex([]string{"name"})

	return Groups.Init()
}
```

Dos familias de métodos, no intercambiables:

- **Escritura:** `model.Insert(data)`, `.Update(data)`, `.Upsert(data)`, `.Delete()` → encadenan `.Command()` (`et.Items`) o `.CommandOne()` (`et.Item`).
- **Lectura:** `model.Select(…)` / `model.Data(…)` → `.Where(model.Column("x").Eq(v))`, `.And(…)`, `.Or(…)`, `.OrderBy(col, asc)` → terminan en `.First()` (`et.Item`), `.All()` (`et.Items`), `.Page(page, rows)` (`et.Items`, sin total) o `.List(page, rows)` (`et.List` con total).

No existe `.One()` en `linq`, y `.Page(…)` ya ejecuta la consulta: no se le encadena `.List()`. SQL crudo (`model.Query(sql, args...)`, `model.QueryOne`) solo para lo que `linq` no modela, siempre con argumentos enlazados, nunca concatenando valores.

### 3. Middleware y claim

- CORS: `middleware.AllowAll(origenes []string).Handler`. No existe `middleware.Cors`.
- Request ID: `middleware.RequestID`. No existe `middleware.RequestId`.
- Existen tal cual: `middleware.Autentication`, `middleware.Authorization`, `middleware.Ephemeral`, `middleware.Logger`, `middleware.Recoverer`.
- Telemetría: helpers (`middleware.NewMetric`, `middleware.PushTelemetry`), no un middleware de chi.
- El perfil del JWT es `claim.Claim.ProfileId` / `claim.ProfileId(r)`. No existe `ProfileTp`.

### 4. Cliente HTTP (`request`)

En error el `*Body` es `nil`: comprueba `status.Ok` antes de `body.ToJson()`. Un timeout devuelve `Status.Code == 408` con el `defaultValue` como body. El header debe llevar `Content-Type`. No existe `request.Do`. Evita `multipart/form-data`: hoy provoca un panic.

### 5. Proyecto nuevo

1. Generador según el ORM (ver `MARCO.md` §2): `go run github.com/celsiainternet/elvis/cmd/create go` (linq) o `go run github.com/celsiainternet/jdb/cmd/create go` (jdb). Con el de jdb, agrega `github.com/celsiainternet/jdb` al `go.mod`: `cmd/install` de elvis no lo hace. Ejecútalo donde ya exista el `go.mod`: el generador toma de ahí el módulo. Menú del generador: **Project** (proyecto completo), **Microservice** (servicio dentro de un repo existente), **Modelo** (modelo + handler CRUD en un paquete existente), **Rpc** (stub RPC en un paquete existente).
2. `go run github.com/celsiainternet/elvis/cmd/install` para instalar esta guía y, si el proyecto ya tiene código, ejecutar `go mod tidy`. No fija versiones: usa las que exige elvis. No bajes dependencias con `go get <módulo>@<versión vieja>`: Go bajaría también elvis para cumplirla.
3. Las plantillas traen valores de Celsia (`APP`, `COMPANY`, `URL_API_FAILURE`): son placeholders, reemplázalos. `scripts/*.http` trae `@token=` vacío: usa un JWT local y no lo subas al repositorio.
4. Los `$ROLE`, `$NS`, `$PORT`, `$IMAGE`… de `oke-*.yml` los resuelve el pipeline; no los rellenes a mano.
5. Con `cmd/create`, la opción **Modelo** agrega `Define<Modelo>(db)` a `initModels`, registra las rutas en `Routes()` de `router.go` y ejecuta `go mod tidy`. Si `router.go` fue editado y no conserva la línea `ctx := context.Background()`, avisa y hay que copiar las rutas del final de `h<Modelo>.go`. Un modelo con schema en un paquete creado sin schema se rechaza con la lista de cambios a hacer antes (campo `Db`, `jdb.Load` en `api.go`, `initModels(c.Db)` en `Controller.Init`).

Layout que genera `cmd/create` (linq):

```
cmd/<servicio>/                  main.go + Dockerfile
internal/service/<servicio>/     service.go (arma el router, cache/db/event.Load)
internal/service/<servicio>/v1/  api.go (monta pkg.Router en pkg.PackagePath)
pkg/<servicio>/                  controller.go, router.go, event.go, msg.go, config.go, rpc.go
                                 (+ model.go, schema.go, h<Modelo>.go si hay schema)
deployments/<servicio>/          local.yml, oke-template.yml, oke-statefulset-template.yml
scripts/<servicio>.http          peticiones de ejemplo
```

El generador de jdb genera en cambio `internal/models/<dominio>/` (modelos jdb), `internal/services/<servicio>/` y `pkg/<servicio>/router-<modelo>.go`.

### 6. Variables de entorno

| Variable | Paquete | Default |
|---|---|---|
| `DB_DRIVER`, `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` | `elvis/jdb` | — |
| `DB_APPLICATION_NAME` | `elvis/jdb` | `elvis` |
| `USE_CORE` | `elvis/jdb` | `true` (tablas series/records/recycling) |
| `REDIS_HOST`, `REDIS_PASSWORD`, `REDIS_DB` | cache | — |
| `NATS_HOST`, `NATS_USER`, `NATS_PASSWORD` | event | — |
| `CONSOLE_LOG_EVENTS` | console | `true` (publica cada log en NATS) |
| `SECRET` | claim | `1977` (cámbiala en todo entorno real) |
| `HOST`, `RPC_HOST`, `RPC_PORT` | jrpc / router | `localhost`, `4200` |
| `PIPE_HOST`, `PIPE_PORT` | jrpc | —, `4200` |
| `AUTHORIZATION_METHOD` | router / middleware | — |
| `RESILIENCE_TOTAL_ATTEMPTS`, `RESILIENCE_TIME_ATTEMPTS` | resilience | `3`, `30` s |
| `PRODUCTION` | dt | `true` |

`elvis/jdb` (paquete de Elvis, drivers postgres/mysql/oracle) no es la librería `github.com/celsiainternet/jdb`. Para esta última, ver `JDB.md`.
