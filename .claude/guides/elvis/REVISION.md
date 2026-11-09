<!-- Generado por elvis devel con cmd/install. No lo edites: se sobrescribe al reinstalar. -->

## Checklist de revisión

Úsalo antes de dar un cambio por terminado, o cuando el usuario pida revisar código de un proyecto que usa Elvis o jdb. Revisa solo el código que el cambio agrega o modifica; lo existente se menciona, no se corrige sin que se pida.

### Compila y pasa las verificaciones

- [ ] `gofmt -l .` no lista archivos.
- [ ] `go build ./... && go vet ./...` sin errores.
- [ ] `go test ./...` si el proyecto tiene tests.

### Usa la librería en lugar de reimplementarla

Cada hallazgo de esta búsqueda en el diff es un candidato a corregir (ver `ELVIS.md` §1):

```bash
git diff -U0 | grep -E '^\+' | grep -nE \
  'os\.Getenv|fmt\.Print|log\.Print|json\.NewEncoder\(w\)|w\.Write\(|\br\.(Get|Post|Put|Delete|Patch)\(|envar\.(Envar|Setvar)|errors\.New\("|http\.Client\{|map\[string\]interface\{\}'
```

- [ ] Rutas registradas con `router.*`.
- [ ] Respuestas con `response.*`.
- [ ] Configuración con `envar.Get*` / `envar.Set*`, sin funciones deprecadas.
- [ ] Logs con `console.*`.
- [ ] Datos dinámicos en `et.Json`; ningún struct de conveniencia nuevo.
- [ ] Llamadas a otros servicios por `jrpc` o `event`, nunca a su base.

### Errores

- [ ] Errores nuevos con `fmt.Errorf(msg.MSG_X, args...)` y la constante en `elvis/msg` o en el `msg.go` del paquete (sin duplicar una que ya existe en `elvis/msg`).
- [ ] Ningún formato variable o con datos del usuario pasado a `console.Alert*`, `console.Error*`, `logs.*f` o `utility.NewError*`.
- [ ] El error se registra una sola vez, en el handler o worker; las capas intermedias solo lo devuelven.
- [ ] No se descartan errores con `_` salvo que el comentario explique por qué.

### Datos

- [ ] El ORM es el que corresponde (`MARCO.md` §2): jdb si el `go.mod` lo requiere, `linq` si no.
- [ ] Si usa jdb, el cambio no cae en ninguno de los defectos de `JDB.md` §4.
- [ ] Toda consulta que devuelve listas está paginada.
- [ ] Ningún valor del usuario se concatena en SQL, ni se pasa a `CALC`, a una agregación o como nombre de campo.
- [ ] La clave de caché se invalida donde se escribe el dato.

### Legibilidad

- [ ] Toda función, método y tipo nuevo tiene comentario Elvis (`/** … @param … @return … **/`).
- [ ] Identificadores en inglés.
- [ ] Sin llamadas anidadas en los argumentos cuando una variable con nombre lo aclara.
- [ ] Sin abstracciones, interfaces o capas que el cambio no necesita.
- [ ] El cambio se limita a lo pedido.
