package create

import "github.com/celsiainternet/elvis/file"

func MakeEnv(packageName string) error {
	_, _ = file.MakeFile(".", ".env", modelEnvar, packageName)

	return nil
}
