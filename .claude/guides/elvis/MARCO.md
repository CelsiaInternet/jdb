<!-- Generado por elvis devel con cmd/install. No lo edites: se sobrescribe al reinstalar. -->

## Marco de trabajo para agentes: código sobre Elvis y jdb

Este marco aplica a cualquier proyecto cuyo `go.mod` requiera `github.com/celsiainternet/elvis` o `github.com/celsiainternet/jdb`. El objetivo es que el código que escribe un agente sea **entendible, mantenible y escalable por un humano**: que se parezca al resto del servicio, use lo que ya resuelven las librerías y no agregue piezas que alguien tenga que descifrar después.

Se complementa con:
- `.claude/guides/elvis/ELVIS.md`: catálogo de patrones de Elvis y firmas verificadas.
- `.claude/guides/jdb/JDB.md`: catálogo de jdb y defectos conocidos (solo si el proyecto importa jdb; se instala con `go run github.com/celsiainternet/jdb/cmd/install`).
- `.claude/guides/elvis/REVISION.md`: checklist para revisar un cambio antes de darlo por terminado.

Si una regla de este marco choca con el `CLAUDE.md` propio del proyecto, gana el del proyecto: es una decisión local tomada a propósito.

### 1. Orden de prioridad

Antes de escribir algo nuevo, busca la solución en este orden y usa la primera que sirva:

1. **Un patrón que ya existe en el mismo servicio.** Copia la forma de un handler, modelo o router vecino antes que inventar una.
2. **Un paquete de Elvis o jdb.** Ver el catálogo en `ELVIS.md` / `JDB.md`.
3. **La librería estándar de Go.**
4. **Una dependencia nueva**, solo si las anteriores no alcanzan, y explicando el motivo en la respuesta al usuario.

No reimplementes lo que ya da la librería (respuestas HTTP, lectura de variables de entorno, caché, eventos, RPC, JWT, generación de ids). Si la librería tiene un defecto que te impide usarla, dilo; no lo rodees en silencio con código propio.

### 2. Reglas de decisión

| Necesidad | Regla |
|---|---|
| **ORM** | Si el `go.mod` requiere `github.com/celsiainternet/jdb`, los modelos nuevos usan **jdb** (`jdb.NewModel`). Si solo requiere elvis, usan **`linq`** (`linq.NewModel`). Los modelos existentes no se migran salvo que se pida. |
| **Layout y generador** | Proyecto con jdb: layout `internal/models/<dominio>/` + `internal/services/<servicio>/`, generador `go run github.com/celsiainternet/jdb/cmd/create go`. Proyecto solo con elvis: layout `internal/service/<servicio>/` + `pkg/<servicio>/`, generador `go run github.com/celsiainternet/elvis/cmd/create go`. Nunca mezcles los dos layouts en un repo; `find internal -maxdepth 2 -type d` dice cuál está en uso. |
| **Logs** | `console` en código de aplicación (`console.LogK`, `console.AlertE`, `console.Error`): además de imprimir, publica el log como evento `logs` en NATS. `logs` es para el interior de las librerías (no publica eventos). Nunca `fmt.Println` ni `log.Print`. |
| **Errores** | Ver sección 3. |
| **Configuración** | `envar.GetStr/GetInt/GetInt64/GetFloat64/GetBool(default, "VAR")` para leer; `envar.SetStr/SetInt/SetBool(nombre, default, uso, "VAR")` cuando también debe ser flag. No uses `os.Getenv` ni las funciones marcadas como deprecadas (`EnvarStr`, `EnvarInt`, `SetvarStr`, …). Revisa la tabla de variables de `ELVIS.md` antes de inventar una nueva. |
| **Datos entre funciones** | `et.Json` con claves en snake_case para entradas, salidas y DTOs internos; `et.Item` / `et.Items` / `et.List` para resultados. Los `struct` se reservan para entidades de dominio, modelos, interfaces de repositorio y tipos del framework. No crees structs de conveniencia solo para agrupar campos que viajan como JSON. |
| **Comunicación entre servicios** | Síncrona: `jrpc.CallJson/CallItem/CallItems/CallList`. Asíncrona: `event.Publish` + `event.Subscribe` / `event.Queue`. Nunca leas ni escribas en la base o el schema de otro servicio. |
| **Rutas HTTP** | Siempre `router.PublicRoute` / `ProtectRoute` / `AuthorizationRoute` / `With`. Registrar con `r.Get` / `r.Post` deja la ruta fuera del API Gateway. |
| **Respuestas HTTP** | `response.ITEM` / `ITEMS` / `LIST` / `JSON` / `HTTPError` / `Unauthorized` / `Forbidden`. Nunca `json.NewEncoder(w)` ni `w.Write`. |
| **Caché** | `cache` (Redis) para datos compartidos entre réplicas, con claves de `cache.GenKey(...)` y prefijo del servicio; `mem` solo para datos locales del proceso. Invalida la clave en el mismo lugar donde se escribe el dato. |

### 3. Errores

| Situación | Patrón |
|---|---|
| Error nuevo (validación, negocio) | `fmt.Errorf(msg.MSG_X, args...)`. El texto sale de una constante: primero `elvis/msg`; si no existe, el `msg.go` del paquete. No escribas el texto en línea. |
| Propagar un error de una capa inferior | `return err` tal cual; si hace falta contexto, `fmt.Errorf("%s: %w", contexto, err)`. Nunca conviertas el error a texto (`err.Error()`) para volver a crearlo: se pierde `errors.Is`. |
| Registrar en el log | Una sola vez, en el borde: el handler HTTP responde con `response.HTTPError(w, r, status, err.Error())`; un worker o suscriptor de eventos usa `console.AlertE(err)`. Las capas intermedias no registran, solo devuelven. |

Usa siempre un formato constante (una constante de `msg`) en las funciones con `f` (`fmt.Errorf`, `console.AlertF`, `console.ErrorF`, `logs.Alertf`, `utility.NewErrorf`), nunca datos del usuario como formato. En versiones de Elvis anteriores a la que trae esta guía, esas funciones y las variantes con `M`/`m` reformateaban el mensaje: un `%` dentro de los datos se imprimía mal y `%w` no envolvía. Si el servicio fija una versión anterior, no les pases texto con `%` ni dependas de `%w` en ellas.

### 4. Entendible

- **Simplicidad como regla:** nada de abstracciones, capas o interfaces que el problema no pida, ni pasos redundantes. Tres líneas parecidas son mejores que una abstracción prematura.
- **Variables con nombre antes de pasarlas a una función**, en lugar de anidar llamadas en los argumentos. Excepción: un método que aporta significado a la llamada.

  ```go
  // No
  return s.RunFlowFor(id, data.Str("trigger_id"), "", data, userId)

  // Sí
  triggerId := data.Str("trigger_id")
  return s.RunFlowFor(id, triggerId, "", data, userId)
  ```
- **Identificadores Go en inglés** (funciones, variables, parámetros), aunque el dominio sea en español. Los comentarios de documentación, los mensajes (`msg.go`) y las rutas que reflejan un contrato externo pueden ir en español. No renombres código existente solo por esto.
- **Comentario Elvis en toda función, método y tipo:**

  ```go
  /**
  * FunctionName: descripción breve.
  * @param paramName type
  * @return type
  **/
  ```

  Si la función devuelve un `et.Json` con varias claves, documenta las claves en el comentario.
- **Una responsabilidad por archivo**, siguiendo el layout del generador: el router solo registra rutas; el handler lee la petición, llama a la lógica y responde; la lógica y la persistencia viven en el modelo o controlador.

### 5. Mantenible

- **Un modelo por tabla**, definido una sola vez en una función `Define<Modelo>(db)` idempotente (`if Modelo != nil { return nil }`) y llamada desde el arranque del servicio.
- **Un dominio no importa el paquete de otro dominio.** Si necesita sus datos, los pide por RPC o reacciona a sus eventos.
- **Toda la configuración entra por `envar`**, con un default sensato y el nombre de la variable en mayúsculas.
- **Sin texto de negocio en línea:** mensajes y errores en `msg.go` o `elvis/msg`.
- **Cambia solo lo que la tarea pide.** No reformatees, renombres ni "mejores" código vecino sin que se pida; si ves un problema, menciónalo.

### 6. Escalable

- **Los servicios no guardan estado propio en memoria** que deba ser igual entre réplicas: usa `cache` o la base.
- **Toda consulta que devuelve listas se pagina** (`List(page, rows)`), y las exportaciones grandes usan `response.Stream`.
- **El trabajo lento o reintentable va a eventos** (`event.Publish` + `event.Queue` para repartir entre réplicas) en lugar de hacerse dentro de la petición HTTP.
- **Las llamadas a sistemas externos** pasan por `request` (con timeout) y, si deben reintentarse, por `resilience`.
- **Las rutas se registran con `router.*`** para que el API Gateway las conozca y las balancee.

### 7. Actualizar la versión de elvis o jdb en un servicio

1. `go get github.com/celsiainternet/elvis@vX.Y.Z` (y/o jdb) y `go mod tidy`.
2. `go build ./... && go vet ./...`. Corrige cada error contra el código de la versión nueva, no de memoria: `ELVIS.md` / `JDB.md` de esa versión y, si no basta, la firma real en `$(go env GOMODCACHE)/github.com/celsiainternet/<lib>@vX.Y.Z`.
3. Reemplaza las funciones que la versión nueva marque como deprecadas solo en los archivos que ya estás tocando.
4. Vuelve a ejecutar `go run github.com/celsiainternet/elvis/cmd/install` (y `go run github.com/celsiainternet/jdb/cmd/install` si el proyecto usa jdb) para actualizar las guías a la versión nueva.
5. Limita el cambio a la actualización: los refactors van en otro cambio.

### 8. Fuera de alcance

- `library/old/`: copias archivadas de versiones viejas de Elvis. No son fuente de patrones ni de firmas.
- `vendor/`, `bin/`, `dist/` y código generado.
