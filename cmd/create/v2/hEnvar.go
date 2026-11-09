package create

import "github.com/celsiainternet/elvis/file"

/**
* MakeEnv: Escribe .env si no existe.
* @param name string
* @return error
**/
func MakeEnv(name string) error {
	_, err := file.MakeFile(".", ".env", modelEnvar, name)
	return err
}
