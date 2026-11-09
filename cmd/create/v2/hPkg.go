package create

import (
	"fmt"
	"os"
	"strings"

	"github.com/celsiainternet/elvis/file"
	"github.com/celsiainternet/elvis/strs"
)

/**
* routesAnchor: Linea que el generador escribe en Routes() justo despues de
* las rutas; las rutas de los modelos nuevos se insertan antes de ella.
**/
const routesAnchor = "\n\tctx := context.Background()\n"

/**
* upsertModelInit: Escribe pkg/<packageName>/model.go la primera vez que se
* agrega un modelo y, en las siguientes, agrega la llamada a
* models.Define<modelo>(db) en initModels (file.MakeFile nunca sobrescribe).
* @param path, packageName, module, modelo string
* @return error
**/
func upsertModelInit(path, packageName, module, modelo string) error {
	modelPath := strs.Format(`%s/model.go`, path)
	if !file.ExistPath(modelPath) {
		_, err := file.MakeFile(path, "model.go", modelModel, packageName, module, modelo)
		return err
	}

	content, err := file.ReadFile(modelPath)
	if err != nil {
		return err
	}

	call := strs.Format(`models.Define%s(db)`, modelo)
	if strings.Contains(content, call) {
		return nil
	}

	marker := "\n\treturn nil\n}"
	idx := strings.LastIndex(content, marker)
	if idx == -1 {
		return fmt.Errorf("no se pudo actualizar initModels en %s", modelPath)
	}

	block := strs.Format("\n\tif err := %s; err != nil {\n\t\treturn err\n\t}\n", call)
	content = content[:idx] + block + content[idx:]

	return os.WriteFile(modelPath, []byte(content), 0644)
}

/**
* upsertRoutes: Agrega a Routes() de pkg/<name>/router.go las rutas de un
* modelo nuevo. No hace nada si el handler ya esta enrutado. Devuelve false
* si router.go fue editado y ya no tiene la linea de anclaje: en ese caso
* las rutas hay que copiarlas a mano.
* @param path, handler, routes string
* @return bool, error
**/
func upsertRoutes(path, handler, routes string) (bool, error) {
	routerPath := strs.Format(`%s/router.go`, path)
	content, err := file.ReadFile(routerPath)
	if err != nil {
		return false, err
	}

	if strings.Contains(content, strs.Format(`rt.%s,`, handler)) {
		return true, nil
	}

	idx := strings.Index(content, routesAnchor)
	if idx == -1 {
		return false, nil
	}

	content = content[:idx] + "\n" + routes + content[idx:]
	return true, os.WriteFile(routerPath, []byte(content), 0644)
}

/**
* hasModels: Indica si el paquete ya inicializa modelos (Controller.Init
* llama a initModels), es decir, si se creo con schema y tiene base de datos.
* @param path string
* @return bool
**/
func hasModels(path string) bool {
	content, err := file.ReadFile(strs.Format(`%s/controller.go`, path))
	if err != nil {
		return false
	}

	return strings.Contains(content, "initModels(c.Db)")
}

/**
* MakePkg: Genera pkg/<name>: config, eventos, mensajes, RPC, controller,
* router y el handler del primer modelo (CRUD jdb si hay schema, un caso de
* uso vacio si no).
* @param module, name, schema string
* @return error
**/
func MakePkg(module, name, schema string) error {
	name = strs.Lowcase(name)
	modelo := strs.Titlecase(name)
	path, err := file.MakeFolder("pkg", name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "event.go", modelEvent, name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "msg.go", modelMsg, name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "config.go", modelConfig, name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "rpc.go", modelhRpc, name)
	if err != nil {
		return err
	}

	handlerFile := strs.Format(`router-%s.go`, name)
	if len(schema) == 0 {
		_, err = file.MakeFile(path, "controller.go", modelController, name)
		if err != nil {
			return err
		}

		_, err = file.MakeFile(path, handlerFile, modelHandler, name, modelo, name)
		if err != nil {
			return err
		}

		_, err = file.MakeFile(path, "router.go", modelRouter, name, name)
		return err
	}

	_, err = file.MakeFile(path, "controller.go", modelDbController, name)
	if err != nil {
		return err
	}

	err = upsertModelInit(path, name, module, modelo)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, handlerFile, modelDbHandler, name, modelo, name, module)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "router.go", modelDbRouter, name, modelo)
	return err
}

/**
* MakeModel: Agrega un modelo a un paquete existente en pkg/<packageName>:
* con schema, el modelo jdb en internal/models/<packageName> y su CRUD; sin
* schema, un caso de uso vacio. Registra las rutas en router.go y devuelve
* false si no se pudieron registrar solas.
* @param module, packageName, modelo, schema string
* @return bool, error
**/
func MakeModel(module, packageName, modelo, schema string) (bool, error) {
	packageName = strs.Lowcase(packageName)
	modelo = strs.Lowcase(modelo)
	title := strs.Titlecase(modelo)
	path := strs.Format(`./pkg/%s`, packageName)
	if !file.ExistPath(path) {
		return false, fmt.Errorf("no existe el paquete %s; crea primero el microservicio", path)
	}

	handlerFile := strs.Format(`router-%s.go`, modelo)
	if len(schema) == 0 {
		_, err := file.MakeFile(path, handlerFile, modelHandler, packageName, title, modelo)
		if err != nil {
			return false, err
		}

		routes := fillTemplate(modelRoutes, title, modelo)
		return upsertRoutes(path, modelo, routes)
	}

	if !hasModels(path) {
		return false, fmt.Errorf(`el paquete %s se creo sin schema y no tiene base de datos; antes de agregar un modelo con schema:
  1. controller.go: agrega el campo Db *jdb.DB a Controller y, en Init, llama a initModels(c.Db) y devuelve su error
  2. internal/services/<servicio>/v1/api.go: importa github.com/celsiainternet/jdb/jdb y el driver (_ ".../jdb/drivers/postgres"), carga la base con jdb.Load() y pasala como pkg.Controller{Db: db}
  3. vuelve a ejecutar la opcion Modelo`, path)
	}

	// El handler expone un metodo RPC en Services, definido en rpc.go
	_, err := file.MakeFile(path, "rpc.go", modelhRpc, packageName)
	if err != nil {
		return false, err
	}

	err = makeModelData(packageName, modelo, schema)
	if err != nil {
		return false, err
	}

	err = upsertModelInit(path, packageName, module, title)
	if err != nil {
		return false, err
	}

	_, err = file.MakeFile(path, handlerFile, modelDbHandler, packageName, title, modelo, module)
	if err != nil {
		return false, err
	}

	routes := fillTemplate(modelDbRoutes, title, modelo)
	return upsertRoutes(path, strs.Format(`get%sById`, title), routes)
}

/**
* MakeRpc: Genera pkg/<name>/rpc.go (Services + StartRpcServer) si no existe.
* @param name string
* @return error
**/
func MakeRpc(name string) error {
	path := strs.Format(`./pkg/%s`, strs.Lowcase(name))
	if !file.ExistPath(path) {
		return fmt.Errorf("no existe el paquete %s; crea primero el microservicio", path)
	}

	_, err := file.MakeFile(path, "rpc.go", modelhRpc, strs.Lowcase(name))
	return err
}

/**
* fillTemplate: Sustituye $1, $2, ... en model, igual que file.MakeFile.
* @param model string, args ...string
* @return string
**/
func fillTemplate(model string, args ...string) string {
	result := model
	for i, v := range args {
		result = strings.ReplaceAll(result, strs.Format(`$%d`, i+1), v)
	}

	return result
}
