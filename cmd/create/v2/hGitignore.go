package create

import "github.com/celsiainternet/elvis/file"

func MakeGitignore(packageName string) error {
	_, _ = file.MakeFile(".", ".gitignore", modelGitignore)

	return nil
}
