package create

import "github.com/celsiainternet/elvis/file"

func MakeTest(name string) error {
	_, err := file.MakeFolder("test")
	if err != nil {
		return err
	}

	return nil
}
