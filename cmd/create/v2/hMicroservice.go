package create

/**
* MkProject
* @param packageName, name, author, schema string
* @return error
**/
func MkProject(packageName, name, author, schema string) error {
	ProgressNext(20)
	err := MkMicroservice(packageName, name, schema)
	if err != nil {
		return err
	}

	ProgressNext(20)
	err = MakeReadme(name)
	if err != nil {
		return err
	}

	ProgressNext(20)
	err = MakeEnv(name)
	if err != nil {
		return err
	}

	ProgressNext(20)
	err = MakeGitignore(name)
	if err != nil {
		return err
	}

	ProgressNext(20)

	return nil
}

/**
* MkMicroservice
* @param packageName, name, schema string
* @return error
**/
func MkMicroservice(packageName, name, schema string) error {
	ProgressNext(10)
	err := MakeCmd(packageName, name)
	if err != nil {
		return err
	}

	ProgressNext(10)
	err = MakeDeployments(name)
	if err != nil {
		return err
	}

	ProgressNext(10)
	err = MakeInternal(packageName, name, schema)
	if err != nil {
		return err
	}

	ProgressNext(10)
	err = MakePkg(name, schema)
	if err != nil {
		return err
	}

	ProgressNext(10)
	err = MakeScripts(name)
	if err != nil {
		return err
	}

	ProgressNext(40)
	err = MakeTest(name)
	if err != nil {
		return err
	}

	ProgressNext(10)

	return nil
}

/**
* MkMolue
* @param packageName, modelo, schema string
* @return error
**/
func MkMolue(packageName, modelo, schema string) error {
	ProgressNext(10)
	err := MakeModel(packageName, modelo, schema)
	if err != nil {
		return err
	}

	ProgressNext(90)

	return nil
}

/**
* MkRpc
* @param name, modelo string
* @return error
**/
func MkRpc(name, modelo string) error {
	ProgressNext(10)
	err := MakeRpc(name, modelo)
	if err != nil {
		return err
	}

	ProgressNext(90)

	return nil
}

/**
* DeleteMicroservice
* @param packageName string
* @return error
**/
func DeleteMicroservice(packageName string) error {
	ProgressNext(10)
	err := DeleteCmd(packageName)
	if err != nil {
		return err
	}

	ProgressNext(80)

	ProgressNext(10)

	return nil
}
