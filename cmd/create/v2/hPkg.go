package create

import (
	"github.com/celsiainternet/elvis/file"
	"github.com/celsiainternet/elvis/strs"
)

func MakePkg(name, schema string) error {
	name = strs.Lowcase(name)
	modelo := strs.Titlecase(name)
	pkgPath, err := file.MakeFolder("pkg", name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(pkgPath, "event.go", modelEvent, name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(pkgPath, "config.go", modelConfig, name)
	if err != nil {
		return err
	}

	if len(schema) > 0 {
		_, err = file.MakeFile(pkgPath, "controller.go", modelDbController, name)
		if err != nil {
			return err
		}

		_, err = file.MakeFile(pkgPath, "model.go", modelModel, name, modelo)
		if err != nil {
			return err
		}

		title := strs.Titlecase(name)
		_, err = file.MakeFile(pkgPath, "router.go", modelDbRouter, name, title)
		if err != nil {
			return err
		}

		_, err = file.MakeFile(pkgPath, "rpc.go", modelhRpc, name, modelo)
		if err != nil {
			return err
		}
	} else {
		_, err = file.MakeFile(pkgPath, "controller.go", modelController, name)
		if err != nil {
			return err
		}

		_, err = file.MakeFile(pkgPath, "router.go", modelRouter, name, strs.Lowcase(name))
		if err != nil {
			return err
		}
	}

	routerFileName := strs.Format(`router-%s.go`, name)
	_, err = file.MakeFile(pkgPath, routerFileName, modelDbModelRouter, name, modelo, strs.Uppcase(modelo), strs.Lowcase(modelo))
	if err != nil {
		return err
	}

	return nil
}

func MakeModel(packageName, modelo, schema string) error {
	modelo = strs.Lowcase(modelo)
	modelPath := strs.Format(`./internal/models/%s`, packageName)

	if len(schema) > 0 {
		_, _ = file.MakeFile(modelPath, "schema.go", modelSchema, packageName, "schema", schema)

		modelFileName := strs.Format(`%s.go`, modelo)
		_, _ = file.MakeFile(modelPath, modelFileName, modelData, packageName, strs.Titlecase(modelo), strs.Lowcase(modelo))
	}

	pkgPath := strs.Format(`./pkg/%s`, packageName)

	routerFileName := strs.Format(`router-%s.go`, modelo)
	_, err := file.MakeFile(pkgPath, routerFileName, modelDbModelRouter, packageName, modelo, strs.Titlecase(modelo), strs.Lowcase(modelo))
	if err != nil {
		return err
	}

	return nil
}

func MakeRpc(name, modelo string) error {
	path, err := file.MakeFolder("pkg", name)
	if err != nil {
		return err
	}

	modelo = strs.Titlecase(modelo)
	_, err = file.MakeFile(path, "rpc.go", modelhRpc, name, modelo)
	if err != nil {
		return err
	}

	return nil
}
