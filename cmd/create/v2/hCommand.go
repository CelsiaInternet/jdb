package create

import (
	"fmt"

	"github.com/celsiainternet/elvis/strs"
	"github.com/celsiainternet/elvis/utility"
	"github.com/spf13/cobra"
)

var CmdProject = &cobra.Command{
	Use:   "project [name author schema]",
	Short: "Create a new project with a base microservice.",
	Long:  "Template project including README, .env, .gitignore, agents rules and guides, and a base microservice with folders cmd, deployments, internal, pkg, scripts and test.",
	Run: func(cmd *cobra.Command, args []string) {
		packageName, err := moduleName()
		if err != nil {
			fmt.Println(err)
			return
		}

		name, err := PrompStr("Name", true)
		if err != nil {
			fmt.Printf("\tPrompt failed: %v\n", err)
			return
		}

		author, err := PrompStr("Author", true)
		if err != nil {
			fmt.Printf("\tPrompt failed: %v\n", err)
			return
		}

		schema, err := PrompStr("Schema", false)
		if err != nil {
			fmt.Printf("\tPrompt failed: %v\n", err)
			return
		}

		err = MkProject(packageName, name, author, schema)
		if err != nil {
			fmt.Printf("\tCommand failed: %v\n", err)
			return
		}
	},
}

var CmdMicro = &cobra.Command{
	Use:   "micro [name schema]",
	Short: "Create a microservice inside the current project.",
	Long:  "Template microservice including folders cmd, deployments, internal, pkg, scripts, test and www, with the .go files required for making a microservice.",
	Run: func(cmd *cobra.Command, args []string) {
		packageName, err := moduleName()
		if err != nil {
			fmt.Println(err)
			return
		}

		name, err := PrompStr("Name", true)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		schema, err := PrompStr("Schema", false)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		err = MkMicroservice(packageName, name, schema)
		if err != nil {
			fmt.Printf("Command failed %v\n", err)
			return
		}
	},
}

var CmdModelo = &cobra.Command{
	Use:   "modelo [package modelo schema]",
	Short: "Create model to microservice.",
	Long:  "Template model to microservice: with schema, a jdb model in internal/models/<package> with its CRUD handlers; without schema, an empty use case. Routes are registered in router.go.",
	Run: func(cmd *cobra.Command, args []string) {
		module, err := moduleName()
		if err != nil {
			fmt.Println(err)
			return
		}

		packageName, err := PrompStr("Package", true)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		modelo, err := PrompStr("Model", true)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		schema, err := PrompStr("Schema", false)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		routed, err := MkMolue(module, packageName, modelo, schema)
		if err != nil {
			fmt.Printf("Command failed %v\n", err)
			return
		}

		if !routed {
			fmt.Println(strs.Format(`router.go fue editado y no se pudieron registrar las rutas: copia las que estan al final de pkg/%s/router-%s.go en Routes().`, strs.Lowcase(packageName), strs.Lowcase(modelo)))
		}
	},
}

var CmdRpc = &cobra.Command{
	Use:   "rpc [package]",
	Short: "Create rpc model to microservice.",
	Long:  "Adds pkg/<package>/rpc.go (Services + StartRpcServer) when it is missing.",
	Run: func(cmd *cobra.Command, args []string) {
		name, err := PrompStr("Package", true)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		err = MkRpc(name)
		if err != nil {
			fmt.Printf("Command failed %v\n", err)
			return
		}
	},
}

/**
* moduleName: Lee el modulo del go.mod del directorio actual; el generador
* lo usa para los imports internos del proyecto.
* @return string, error
**/
func moduleName() (string, error) {
	result, err := utility.GoMod("module")
	if err != nil || result == "" {
		return "", fmt.Errorf("no se encontro go.mod con module en el directorio actual; ejecuta primero: go mod init <modulo>")
	}

	return result, nil
}
