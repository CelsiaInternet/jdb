package create

import "github.com/celsiainternet/elvis/file"

func MakeReadme(packageName, author string) error {
	_, _ = file.MakeFile(".", "README.md", modelReadme, packageName, author)

	return nil
}
