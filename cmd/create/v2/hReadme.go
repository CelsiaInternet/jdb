package create

import "github.com/celsiainternet/elvis/file"

/**
* MakeReadme: Escribe README.md si no existe.
* @param name, author string
* @return error
**/
func MakeReadme(name, author string) error {
	_, err := file.MakeFile(".", "README.md", modelReadme, name, author)
	return err
}
