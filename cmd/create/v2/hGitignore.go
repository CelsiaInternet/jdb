package create

import "github.com/celsiainternet/elvis/file"

/**
* MakeGitignore: Escribe .gitignore si no existe.
* @param name string
* @return error
**/
func MakeGitignore(name string) error {
	_, err := file.MakeFile(".", ".gitignore", modelGitignore)
	return err
}
