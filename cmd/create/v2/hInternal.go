package create

import (
	"github.com/celsiainternet/elvis/file"
	"github.com/celsiainternet/elvis/strs"
)

/**
* MakeInternal: Genera internal/services/<name> (servidor HTTP y arranque) y,
* si hay schema, internal/models/<name> con el primer modelo jdb.
* @param packageName, name, schema string
* @return error
**/
func MakeInternal(packageName, name, schema string) error {
	name = strs.Lowcase(name)
	servicePath, err := file.MakeFolder("internal", "services", name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(servicePath, "service.go", modelService, packageName, name)
	if err != nil {
		return err
	}

	v1Path, err := file.MakeFolder("internal", "services", name, "v1")
	if err != nil {
		return err
	}

	if len(schema) == 0 {
		_, err = file.MakeFile(v1Path, "api.go", modelApi, packageName, name)
		return err
	}

	_, err = file.MakeFile(v1Path, "api.go", modelDbApi, packageName, name)
	if err != nil {
		return err
	}

	return makeModelData(name, name, schema)
}

/**
* makeModelData: Escribe en internal/models/<packageName> el schema y el
* modelo jdb con su logica de datos (Define, Get, Insert, Upsert, State,
* Delete, All, Query).
* @param packageName, modelo, schema string
* @return error
**/
func makeModelData(packageName, modelo, schema string) error {
	path, err := file.MakeFolder("internal", "models", packageName)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "msg.go", modelMsg, packageName)
	if err != nil {
		return err
	}

	schema = strs.Lowcase(schema)
	schemaTitle := strs.Titlecase(schema)
	_, err = file.MakeFile(path, strs.Format(`schema_%s.go`, schema), modelSchema, packageName, schemaTitle, schema)
	if err != nil {
		return err
	}

	modelo = strs.Lowcase(modelo)
	_, err = file.MakeFile(path, strs.Format(`%s.go`, modelo), modelData, packageName, strs.Titlecase(modelo), modelo, schemaTitle)
	return err
}
