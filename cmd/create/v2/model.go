package create

const modelDockerfile = `# Versión de Go como argumento
ARG GO_VERSION=1.23

# Stage 1: Compilación (builder)
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

# Argumentos para el sistema operativo y la arquitectura
ARG TARGETOS
ARG TARGETARCH

# Instalación de dependencias necesarias
RUN apk update && apk add --no-cache ca-certificates openssl git tzdata \
    && update-ca-certificates

# Configuración de las variables de entorno para la build
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH}

# Directorio de trabajo
WORKDIR /src

# Descargar dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el código fuente
COPY . .

# Formatear el código Go
RUN gofmt -w .

# Compilar el binario
RUN go build -v -o /$1 ./cmd/$1

# Cambiar permisos del binario
RUN chmod +x /$1

# Stage 2: Imagen final mínima
FROM scratch

# Copiar certificados y binario
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /$1 /$1

# Establecer el binario como punto de entrada
ENTRYPOINT ["/$1"]
`

const modelMain = `package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"

	serv "$1/internal/services/$2"
)

func main() {
	envar.SetInt("port", 3300, "Port server", "PORT")
	envar.SetInt("rpc", 4200, "Port rpc server", "RPC_PORT")
	envar.SetStr("dbhost", "localhost", "Database host", "DB_HOST")
	envar.SetInt("dbport", 5432, "Database port", "DB_PORT")
	envar.SetStr("dbname", "", "Database name", "DB_NAME")
	envar.SetStr("dbuser", "", "Database user", "DB_USER")
	envar.SetStr("dbpass", "", "Database password", "DB_PASSWORD")

	srv, err := serv.New()
	if err != nil {
		console.Fatal(err)
	}

	go srv.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	srv.Close()
}
`

const modelService = `package $2

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/middleware"
	"github.com/celsiainternet/elvis/response"
	"github.com/celsiainternet/elvis/strs"
	"github.com/go-chi/chi/v5"

	v1 "$1/internal/services/$2/v1"
)

type Server struct {
	http *http.Server
}

/**
* New: Arma el router HTTP del servicio con los middleware de elvis.
* @return *Server, error
**/
func New() (*Server, error) {
	port := envar.GetInt(3300, "PORT")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	latest := v1.New()

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.HTTPError(w, r, http.StatusNotFound, "404 Not Found")
	})

	r.Mount("/", latest)
	r.Mount("/v1", latest)

	handler := middleware.AllowAll([]string{}).Handler(r)
	addr := strs.Format(":%d", port)
	result := &Server{
		http: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}

	return result, nil
}

/**
* Close: Detiene el servidor HTTP y libera las conexiones del servicio.
**/
func (serv *Server) Close() {
	if serv.http != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := serv.http.Shutdown(ctx); err != nil {
			console.Error(err)
		}
	}

	v1.Close()

	console.LogK("Http", "Shutting down server...")
}

/**
* StartHttpServer: Atiende peticiones HTTP hasta que se llama Close.
**/
func (serv *Server) StartHttpServer() {
	if serv.http == nil {
		return
	}

	console.LogKF("Http", "Running on http://localhost%s", serv.http.Addr)
	err := serv.http.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		console.Fatal(err)
	}
}

/**
* Start: Arranca el servidor HTTP y muestra el banner.
**/
func (serv *Server) Start() {
	go serv.StartHttpServer()

	v1.Banner()

	<-make(chan struct{})
}
`

const modelDbApi = `package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/celsiainternet/elvis/cache"
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/event"
	"github.com/celsiainternet/elvis/jrpc"
	"github.com/celsiainternet/elvis/utility"
	_ "github.com/celsiainternet/jdb/drivers/postgres"
	"github.com/celsiainternet/jdb/jdb"
	"github.com/dimiro1/banner"
	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-colorable"

	pkg "$1/pkg/$2"
)

/**
* New: Conecta cache, eventos, RPC y base de datos, y monta el router del paquete.
* La base se elige con DB_DRIVER (postgres por defecto); para sqlite u oracle
* importa tambien github.com/celsiainternet/jdb/drivers/sqlite o /oracle.
* @return http.Handler
**/
func New() http.Handler {
	r := chi.NewRouter()

	err := pkg.LoadConfig()
	if err != nil {
		console.Panic(err)
	}

	_, err = cache.Load()
	if err != nil {
		console.Panic(err)
	}

	_, err = event.Load()
	if err != nil {
		console.Panic(err)
	}

	err = pkg.StartRpcServer()
	if err != nil {
		console.Panic(err)
	}

	db, err := jdb.Load()
	if err != nil {
		console.Panic(err)
	}

	_pkg := &pkg.Router{
		Repository: &pkg.Controller{
			Db: db,
		},
	}

	r.Mount(pkg.PackagePath, _pkg.Routes())

	return r
}

/**
* Close: Cierra RPC, cache y eventos.
**/
func Close() {
	jrpc.Close()
	cache.Close()
	event.Close()
}

/**
* Banner: Imprime el nombre del servicio al arrancar.
**/
func Banner() {
	time.Sleep(3 * time.Second)
	templ := utility.BannerTitle(pkg.PackageName, 4)
	banner.InitString(colorable.NewColorableStdout(), true, true, templ)
	fmt.Println()
}
`

const modelApi = `package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/celsiainternet/elvis/cache"
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/event"
	"github.com/celsiainternet/elvis/jrpc"
	"github.com/celsiainternet/elvis/utility"
	"github.com/dimiro1/banner"
	"github.com/go-chi/chi/v5"
	"github.com/mattn/go-colorable"

	pkg "$1/pkg/$2"
)

/**
* New: Conecta cache, eventos y RPC, y monta el router del paquete.
* @return http.Handler
**/
func New() http.Handler {
	r := chi.NewRouter()

	err := pkg.LoadConfig()
	if err != nil {
		console.Panic(err)
	}

	_, err = cache.Load()
	if err != nil {
		console.Panic(err)
	}

	_, err = event.Load()
	if err != nil {
		console.Panic(err)
	}

	err = pkg.StartRpcServer()
	if err != nil {
		console.Panic(err)
	}

	_pkg := &pkg.Router{
		Repository: &pkg.Controller{},
	}

	r.Mount(pkg.PackagePath, _pkg.Routes())

	return r
}

/**
* Close: Cierra RPC, cache y eventos.
**/
func Close() {
	jrpc.Close()
	cache.Close()
	event.Close()
}

/**
* Banner: Imprime el nombre del servicio al arrancar.
**/
func Banner() {
	time.Sleep(3 * time.Second)
	templ := utility.BannerTitle(pkg.PackageName, 4)
	banner.InitString(colorable.NewColorableStdout(), true, true, templ)
	fmt.Println()
}
`

const modelEvent = `package $1

import (
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/event"
)

/**
* initEvents: Suscribe los manejadores de eventos del paquete.
**/
func initEvents() {
	// TODO: cambia el canal por el evento real que consume el servicio
	err := event.Subscribe(PackageName+"/event", eventAction)
	if err != nil {
		console.Error(err)
	}
}

/**
* eventAction: Maneja un mensaje recibido por NATS.
* @param m event.EvenMessage
**/
func eventAction(m event.EvenMessage) {
	console.LogK("eventAction", m.Data.ToString())
}
`

const modelMsg = `package $1

/**
* Mensajes propios del paquete. Usa primero las constantes de
* github.com/celsiainternet/elvis/msg y agrega aqui solo las que no existan.
**/
const ()
`

const modelConfig = `package $1

import (
	"github.com/celsiainternet/elvis/config"
	"github.com/celsiainternet/elvis/envar"
)

/**
* LoadConfig: Carga la configuracion del paquete para el STAGE actual.
* @return error
**/
func LoadConfig() error {
	stage := envar.GetStr("local", "STAGE")
	return config.Load(stage, PackageName)
}
`

const modelhRpc = `package $1

import (
	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/jrpc"
)

type Services struct{}

/**
* StartRpcServer: Registra los metodos de Services como RPC ($1.Services.<Metodo>).
* @return error
**/
func StartRpcServer() error {
	pkg, err := jrpc.Load(PackageName)
	if err != nil {
		return err
	}

	services := new(Services)
	err = jrpc.Mount(services)
	if err != nil {
		return err
	}

	go pkg.Start()

	return nil
}

/**
* Version: Describe el servicio por RPC.
* @param require et.Json
* @param response *et.Item
* @return error
**/
func (c *Services) Version(require et.Json, response *et.Item) error {
	response.Ok = true
	response.Result = et.Json{
		"method":  "RPC",
		"version": PackageVersion,
		"service": PackageName,
		"host":    HostName,
		"company": envar.GetStr("", "COMPANY"),
		"web":     envar.GetStr("", "WEB"),
		"help":    envar.GetStr("", "RPC_HELP"),
	}

	return console.Rpc(response)
}
`

const modelDbController = `package $1

import (
	"context"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/jdb/jdb"
)

type Controller struct {
	Db *jdb.DB
}

/**
* Version: Describe el servicio.
* @param ctx context.Context
* @return et.Json, error
**/
func (c *Controller) Version(ctx context.Context) (et.Json, error) {
	service := et.Json{
		"version": PackageVersion,
		"service": PackageName,
		"host":    HostName,
		"company": envar.GetStr("", "COMPANY"),
		"web":     envar.GetStr("", "WEB"),
		"help":    "",
	}

	return service, nil
}

/**
* Init: Define los modelos y suscribe los eventos del paquete.
* @param ctx context.Context
* @return error
**/
func (c *Controller) Init(ctx context.Context) error {
	if err := initModels(c.Db); err != nil {
		return err
	}

	initEvents()

	return nil
}

type Repository interface {
	Version(ctx context.Context) (et.Json, error)
	Init(ctx context.Context) error
}
`

const modelController = `package $1

import (
	"context"

	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
)

type Controller struct{}

/**
* Version: Describe el servicio.
* @param ctx context.Context
* @return et.Json, error
**/
func (c *Controller) Version(ctx context.Context) (et.Json, error) {
	service := et.Json{
		"version": PackageVersion,
		"service": PackageName,
		"host":    HostName,
		"company": envar.GetStr("", "COMPANY"),
		"web":     envar.GetStr("", "WEB"),
		"help":    "",
	}

	return service, nil
}

/**
* Init: Suscribe los eventos del paquete.
* @param ctx context.Context
* @return error
**/
func (c *Controller) Init(ctx context.Context) error {
	initEvents()

	return nil
}

type Repository interface {
	Version(ctx context.Context) (et.Json, error)
	Init(ctx context.Context) error
}
`

const modelModel = `package $1

import (
	"github.com/celsiainternet/jdb/jdb"

	models "$2/internal/models/$1"
)

/**
* initModels: Define e inicializa los modelos del paquete.
* @param db *jdb.DB
* @return error
**/
func initModels(db *jdb.DB) error {
	if err := models.Define$3(db); err != nil {
		return err
	}

	return nil
}
`

const modelSchema = `package $1

import (
	"github.com/celsiainternet/jdb/jdb"
)

var schema$2 *jdb.Schema

/**
* defineSchema$2: Registra el schema $3 (postgres lo crea si no existe;
* sqlite y oracle lo usan como prefijo de las tablas: $3_<tabla>).
* @param db *jdb.DB
* @return error
**/
func defineSchema$2(db *jdb.DB) error {
	if schema$2 == nil {
		schema$2 = jdb.NewSchema(db, "$3")
	}

	return nil
}
`

const modelData = `package $1

import (
	"fmt"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/msg"
	"github.com/celsiainternet/elvis/utility"
	"github.com/celsiainternet/jdb/jdb"
)

var $2 *jdb.Model

/**
* Define$2: Define la tabla $3. DefineProjectModel agrega las columnas que
* jdb mantiene solo (_id, project_id, _status, created_at, updated_at, _data,
* jdbid, idx); los atributos que no son columnas se guardan en _data.
* @param db *jdb.DB
* @return error
**/
func Define$2(db *jdb.DB) error {
	if err := defineSchema$4(db); err != nil {
		return err
	}

	if $2 != nil {
		return nil
	}

	model := jdb.NewModel(schema$4, "$3", 1)
	model.DefineProjectModel()
	model.DefineAtribute("name", jdb.TypeDataText)
	model.DefineAtribute("description", jdb.TypeDataText)
	if err := model.Init(); err != nil {
		return err
	}

	$2 = model
	return nil
}

/**
* Get$2ById: Busca un registro por id; Ok=false si no existe.
* @param id string
* @return et.Item, error
**/
func Get$2ById(id string) (et.Item, error) {
	return $2.
		Where(jdb.KEY).Eq(id).
		One()
}

/**
* Insert$2: Crea el registro; si id viene vacio se genera uno. Si el id ya
* existe devuelve Ok=false con el registro actual.
* @param projectId, id, name, description string
* @param data et.Json
* @return et.Item, error
**/
func Insert$2(projectId, id, name, description string, data et.Json) (et.Item, error) {
	if !utility.ValidStr(projectId, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, jdb.PROJECT_ID)
	}

	if !utility.ValidStr(name, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "name")
	}

	id = $2.GetId(id)
	current, err := Get$2ById(id)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok {
		return et.Item{Ok: false, Result: current.Result}, nil
	}

	data[jdb.PROJECT_ID] = projectId
	data[jdb.KEY] = id
	data[jdb.STATUS_ID] = utility.ACTIVE
	data["name"] = name
	data["description"] = description

	return $2.
		Insert(data).
		One()
}

/**
* Upsert$2: Crea el registro o, si ya existe y esta activo, lo actualiza.
* @param projectId, id, name, description string
* @param data et.Json
* @return et.Item, error
**/
func Upsert$2(projectId, id, name, description string, data et.Json) (et.Item, error) {
	current, err := Insert$2(projectId, id, name, description, data)
	if err != nil {
		return et.Item{}, err
	}

	if current.Ok {
		return current, nil
	}

	if current.Str(jdb.STATUS_ID) != utility.ACTIVE {
		return et.Item{}, fmt.Errorf(msg.RECORD_NOT_UPDATE)
	}

	id = current.Str(jdb.KEY)
	delete(data, jdb.STATUS_ID)
	data[jdb.PROJECT_ID] = projectId
	data[jdb.KEY] = id
	data["name"] = name
	data["description"] = description

	return $2.
		Update(data).
		Where(jdb.KEY).Eq(id).
		One()
}

/**
* State$2: Cambia el estado (_status) de un registro.
* @param id, state string
* @return et.Item, error
**/
func State$2(id, state string) (et.Item, error) {
	if !utility.ValidStr(state, 0, []string{""}) {
		return et.Item{}, fmt.Errorf(msg.MSG_ATRIB_REQUIRED, "state")
	}

	current, err := Get$2ById(id)
	if err != nil {
		return et.Item{}, err
	}

	if !current.Ok {
		return et.Item{}, fmt.Errorf(msg.RECORD_NOT_FOUND)
	}

	if current.Str(jdb.STATUS_ID) == state {
		return et.Item{}, fmt.Errorf(msg.RECORD_NOT_CHANGE)
	}

	return $2.
		Update(et.Json{
			jdb.STATUS_ID: state,
		}).
		Where(jdb.KEY).Eq(id).
		One()
}

/**
* Delete$2: Marca el registro para borrar (borrado logico).
* @param id string
* @return et.Item, error
**/
func Delete$2(id string) (et.Item, error) {
	return State$2(id, utility.FOR_DELETE)
}

/**
* All$2: Lista paginada del proyecto. state "" = activos, "*" = todos
* menos los marcados para borrar; search filtra por name (LIKE).
* @param projectId, state, search string
* @param page, rows int
* @return et.List, error
**/
func All$2(projectId, state, search string, page, rows int) (et.List, error) {
	if state == "" {
		state = utility.ACTIVE
	}

	ql := $2.
		Where(jdb.PROJECT_ID).Eq(projectId)
	if state == "*" {
		ql.And(jdb.STATUS_ID).Neg(utility.FOR_DELETE)
	} else {
		ql.And(jdb.STATUS_ID).Eq(state)
	}

	if search != "" {
		ql.And("name").Like("%" + search + "%")
	}

	return ql.
		OrderByAsc("name").
		List(page, rows)
}

/**
* Query$2: Consulta en formato JSON (wheres, selects, order_by, page, rows),
* ver jsql.md en el repo de jdb.
* @param query et.Json
* @return et.List, error
**/
func Query$2(query et.Json) (et.List, error) {
	return jdb.RunQuery($2, query)
}
`

const modelDbHandler = `package $1

import (
	"net/http"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/response"
	"github.com/celsiainternet/jdb/jdb"
	"github.com/go-chi/chi/v5"

	models "$4/internal/models/$1"
)

/**
* Get$2ById: Expone models.Get$2ById por RPC ($1.Services.Get$2ById).
* @param require et.Json
* @param response *et.Item
* @return error
**/
func (c *Services) Get$2ById(require et.Json, response *et.Item) error {
	result, err := models.Get$2ById(require.Str("id"))
	if err != nil {
		return err
	}

	*response = result

	return console.Rpc(response)
}

/**
* upsert$2: POST, crea o actualiza.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) upsert$2(w http.ResponseWriter, r *http.Request) {
	body, err := response.GetBody(r)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	projectId := body.Str(jdb.PROJECT_ID)
	id := body.Str(jdb.KEY)
	name := body.Str("name")
	description := body.Str("description")

	result, err := models.Upsert$2(projectId, id, name, description, body)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* get$2ById: GET /{id}.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) get$2ById(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := models.Get$2ById(id)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* state$2: PUT /state/{id}, body {"state": "..."}.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) state$2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	body, err := response.GetBody(r)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	result, err := models.State$2(id, body.Str("state"))
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* delete$2: DELETE /{id}.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) delete$2(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	result, err := models.Delete$2(id)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/**
* all$2: GET /?project_id=&state=&search=&page=&rows=.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) all$2(w http.ResponseWriter, r *http.Request) {
	query := response.GetQuery(r)
	projectId := query.Str(jdb.PROJECT_ID)
	state := query.Str("state")
	search := query.Str("search")
	page := query.ValInt(1, "page")
	rows := query.ValInt(30, "rows")

	result, err := models.All$2(projectId, state, search, page, rows)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.LIST(w, r, http.StatusOK, result)
}

/**
* query$2: POST /query, body con wheres, selects, order_by, page y rows.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) query$2(w http.ResponseWriter, r *http.Request) {
	body, err := response.GetBody(r)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	result, err := models.Query$2(body)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.LIST(w, r, http.StatusOK, result)
}

/** El generador agrega estas rutas a Routes() de router.go; si no pudo (router.go editado), copialas:
	// $2
	er.ProtectRoute(r, er.Get, "/$3/{id}", rt.get$2ById, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/$3", rt.upsert$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Put, "/$3/state/{id}", rt.state$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Delete, "/$3/{id}", rt.delete$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/$3", rt.all$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/$3/query", rt.query$2, PackageName, PackagePath, host)
**/
`

const modelHandler = `package $1

import (
	"net/http"

	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/response"
)

/**
* $2: Logica del caso de uso.
* @param projectId, id string
* @param params et.Json
* @return et.Item, error
**/
func $2(projectId, id string, params et.Json) (et.Item, error) {
	return et.Item{Ok: true, Result: params}, nil
}

/**
* $3: Handler HTTP de $2.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) $3(w http.ResponseWriter, r *http.Request) {
	body, err := response.GetBody(r)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	projectId := body.Str("project_id")
	id := body.Str("id")

	result, err := $2(projectId, id, body)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.ITEM(w, r, http.StatusOK, result)
}

/** El generador agrega esta ruta a Routes() de router.go; si no pudo (router.go editado), copiala:
	// $2
	er.ProtectRoute(r, er.Post, "/$3", rt.$3, PackageName, PackagePath, host)
**/
`

const modelDbRouter = `package $1

import (
	"context"
	"net/http"
	"os"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/middleware"
	"github.com/celsiainternet/elvis/response"
	er "github.com/celsiainternet/elvis/router"
	"github.com/celsiainternet/elvis/strs"
	"github.com/go-chi/chi/v5"
)

var PackageName = "$1"
var PackageTitle = "$1"
var PackagePath = envar.GetStr("/api/$1", "PATH_URL")
var PackageVersion = envar.GetStr("0.0.1", "VERSION")
var HostName, _ = os.Hostname()

type Router struct {
	Repository Repository
}

/**
* Routes: Registra las rutas del paquete e inicializa el repositorio.
* @return http.Handler
**/
func (rt *Router) Routes() http.Handler {
	defaultHost := strs.Format("http://%s", HostName)
	var host = strs.Format("%s:%d", envar.GetStr(defaultHost, "HOST"), envar.GetInt(3300, "PORT"))

	r := chi.NewRouter()

	er.PublicRoute(r, er.Get, "/version", rt.version, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/routes", rt.routes, PackageName, PackagePath, host)
	// $2
	er.ProtectRoute(r, er.Get, "/{id}", rt.get$2ById, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/", rt.upsert$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Put, "/state/{id}", rt.state$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Delete, "/{id}", rt.delete$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/", rt.all$2, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/query", rt.query$2, PackageName, PackagePath, host)

	ctx := context.Background()
	if err := rt.Repository.Init(ctx); err != nil {
		console.Panic(err)
	}
	middleware.SetServiceName(PackageName)

	console.LogKF(PackageName, "Router version:%s", PackageVersion)
	return r
}

/**
* version: Responde la version del servicio.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) version(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	result, err := rt.Repository.Version(ctx)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}

/**
* routes: Lista las rutas registradas.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) routes(w http.ResponseWriter, r *http.Request) {
	_routes := er.GetRoutes()
	routes := []et.Json{}
	for _, route := range _routes {
		routes = append(routes, et.Json{
			"method": route.Str("method"),
			"path":   route.Str("path"),
		})
	}

	result := et.Items{
		Ok:     true,
		Count:  len(routes),
		Result: routes,
	}

	response.ITEMS(w, r, http.StatusOK, result)
}
`

const modelRouter = `package $1

import (
	"context"
	"net/http"
	"os"

	"github.com/celsiainternet/elvis/console"
	"github.com/celsiainternet/elvis/envar"
	"github.com/celsiainternet/elvis/et"
	"github.com/celsiainternet/elvis/middleware"
	"github.com/celsiainternet/elvis/response"
	er "github.com/celsiainternet/elvis/router"
	"github.com/celsiainternet/elvis/strs"
	"github.com/go-chi/chi/v5"
)

var PackageName = "$1"
var PackageTitle = "$1"
var PackagePath = envar.GetStr("/api/$1", "PATH_URL")
var PackageVersion = envar.GetStr("0.0.1", "VERSION")
var HostName, _ = os.Hostname()

type Router struct {
	Repository Repository
}

/**
* Routes: Registra las rutas del paquete e inicializa el repositorio.
* @return http.Handler
**/
func (rt *Router) Routes() http.Handler {
	defaultHost := strs.Format("http://%s", HostName)
	var host = strs.Format("%s:%d", envar.GetStr(defaultHost, "HOST"), envar.GetInt(3300, "PORT"))

	r := chi.NewRouter()

	er.PublicRoute(r, er.Get, "/version", rt.version, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/routes", rt.routes, PackageName, PackagePath, host)
	// $2
	er.ProtectRoute(r, er.Post, "/", rt.$2, PackageName, PackagePath, host)

	ctx := context.Background()
	if err := rt.Repository.Init(ctx); err != nil {
		console.Panic(err)
	}
	middleware.SetServiceName(PackageName)

	console.LogKF(PackageName, "Router version:%s", PackageVersion)
	return r
}

/**
* version: Responde la version del servicio.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) version(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	result, err := rt.Repository.Version(ctx)
	if err != nil {
		response.HTTPError(w, r, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, r, http.StatusOK, result)
}

/**
* routes: Lista las rutas registradas.
* @param w http.ResponseWriter
* @param r *http.Request
**/
func (rt *Router) routes(w http.ResponseWriter, r *http.Request) {
	_routes := er.GetRoutes()
	routes := []et.Json{}
	for _, route := range _routes {
		routes = append(routes, et.Json{
			"method": route.Str("method"),
			"path":   route.Str("path"),
		})
	}

	result := et.Items{
		Ok:     true,
		Count:  len(routes),
		Result: routes,
	}

	response.ITEMS(w, r, http.StatusOK, result)
}
`

const modelDbRoutes = `	// $1
	er.ProtectRoute(r, er.Get, "/$2/{id}", rt.get$1ById, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/$2", rt.upsert$1, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Put, "/$2/state/{id}", rt.state$1, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Delete, "/$2/{id}", rt.delete$1, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Get, "/$2", rt.all$1, PackageName, PackagePath, host)
	er.ProtectRoute(r, er.Post, "/$2/query", rt.query$1, PackageName, PackagePath, host)
`

const modelRoutes = `	// $1
	er.ProtectRoute(r, er.Post, "/$2", rt.$2, PackageName, PackagePath, host)
`

const restHttp = `@host=localhost:3300
# JWT firmado con el SECRET del servicio (claim.NewToken); no lo subas al repositorio
@token=

### Version (publica)
GET http://{{host}}/api/$1/version

### Rutas registradas
GET http://{{host}}/api/$1/routes
Authorization: Bearer {{token}}

### Crear o actualizar
POST http://{{host}}/api/$1
Content-Type: application/json
Authorization: Bearer {{token}}

{
  "project_id": "-1",
  "name": "Ejemplo"
}
`

const modelReadme = `# $1

Autor: $2

Microservicio construido sobre [elvis](https://github.com/celsiainternet/elvis) y [jdb](https://github.com/celsiainternet/jdb).

## Actualizar elvis y jdb

` + "`" + `` + "`" + `` + "`" + `bash
go get github.com/celsiainternet/elvis@latest github.com/celsiainternet/jdb@latest
go mod tidy
go run github.com/celsiainternet/elvis/cmd/install
go run github.com/celsiainternet/jdb/cmd/install
` + "`" + `` + "`" + `` + "`" + `

## Generar microservicios, modelos y RPC

` + "`" + `` + "`" + `` + "`" + `bash
go run github.com/celsiainternet/jdb/cmd/create go
` + "`" + `` + "`" + `` + "`" + `

## Ejecutar

Requiere la base de datos (` + "`" + `DB_DRIVER` + "`" + `, PostgreSQL por defecto), Redis y NATS; la configuracion esta en ` + "`" + `.env` + "`" + `.

` + "`" + `` + "`" + `` + "`" + `bash
gofmt -w . && go run ./cmd/$1
` + "`" + `` + "`" + `` + "`" + `

## Verificar

` + "`" + `` + "`" + `` + "`" + `bash
go build ./... && go vet ./...
` + "`" + `` + "`" + `` + "`" + `
`

const modelEnvar = `APP=$1
PORT=3300
VERSION=0.0.0
COMPANY=Company
WEB=https://www.company.com
# HOST=http://localhost
HOST=http://host.docker.internal
PATH_URL=/api/$1
PRODUCTION=false

# RPC
RPC_PORT=4200
RPC_HOST=localhost

# DB (jdb)
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=test
DB_USER=test
DB_PASSWORD=test
APP_NAME=$1
USE_CORE=true
DEBUG=false

# REDIS
REDIS_HOST=localhost:6379
REDIS_PASSWORD=test
REDIS_DB=0

# NATS
NATS_HOST=localhost:4222

# CALM
SECRET=test
`

const modelDeploy = `version: "3"

networks:
  $3:
    external: true

services:
  $1:
    image: $1:latest
    logging:
      driver: "json-file"
      options:
        max-size: "1m"
        max-file: "2"
    networks:
      - $3
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.$1.rule=PathPrefix($2)"
      - "traefik.http.services.$1.loadbalancer.server.port=3300"
    deploy:
      replicas: 1
    environment:
      - "APP=$1"
      - "PORT=3300"
      - "VERSION=1.0.1"
      - "COMPANY=Celsia Internet"
      - "WEB=https://www.app.com"
      - "PATH_URL=/api/$1"
      - "PRODUCTION=true"
      - "HOST=stack"
      # DB (jdb)
      - "DB_DRIVER=postgres"
      - "DB_HOST="
      - "DB_PORT=5432"
      - "DB_NAME=internet"
      - "DB_USER=internet"
      - "DB_PASSWORD="
      - "APP_NAME=$1"
      # REDIS
      - "REDIS_HOST="
      - "REDIS_PASSWORD="
      - "REDIS_DB=0"
      # NATS
      - "NATS_HOST=nats:4222"
      # CALM
      - "SECRET="
      # RPC
      - "RPC_PORT=4200"
`

const modelGitignore = `# Created by https://www.toptal.com/developers/gitignore/api/go
# Edit at https://www.toptal.com/developers/gitignore?templates=go

### Go ###
# If you prefer the allow list template instead of the deny list, see community template:
# https://github.com/github/gitignore/blob/main/community/Golang/Go.AllowList.gitignore
#
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib
.env
data
build
sql
.vscode
deployments/oke.yml

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Credencial acces token to AWS server
*.pem`
