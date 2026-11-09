package create

import (
	"github.com/celsiainternet/elvis/file"
	"github.com/celsiainternet/elvis/strs"
)

/**
* MakeCmd: Genera cmd/<name> (Dockerfile y main.go).
* @param packageName, name string
* @return error
**/
func MakeCmd(packageName, name string) error {
	path, err := file.MakeFolder("cmd", name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "Dockerfile", modelDockerfile, name)
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "main.go", modelMain, packageName, name)
	if err != nil {
		return err
	}

	return nil
}

/**
* DeleteCmd: Borra lo que genera makeMicroservice para name.
* @param name string
* @return error
**/
func DeleteCmd(name string) error {
	paths := []string{
		strs.Format(`./cmd/%s`, name),
		strs.Format(`./deployments/%s`, name),
		strs.Format(`./internal/services/%s`, name),
		strs.Format(`./internal/models/%s`, name),
		strs.Format(`./pkg/%s`, name),
		strs.Format(`./scripts/%s.http`, name),
		strs.Format(`./www/%s`, name),
	}
	for _, path := range paths {
		if !file.ExistPath(path) {
			continue
		}

		_, err := file.RemoveFile(path)
		if err != nil {
			return err
		}
	}

	return nil
}
