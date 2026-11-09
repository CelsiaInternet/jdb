package create

import (
	"fmt"
	"os/exec"
	"strings"
)

/**
* Command: Ejecuta cada comando de coms y devuelve su salida estandar.
* @param coms []string
* @return [][]byte, error
**/
func Command(coms []string) ([][]byte, error) {
	var result [][]byte
	for _, com := range coms {
		args := strings.Fields(com)
		if len(args) == 0 {
			continue
		}

		out, err := exec.Command(args[0], args[1:]...).Output()
		if err != nil {
			return result, err
		}
		result = append(result, out)
	}

	return result, nil
}

/**
* GoModTidy: Ejecuta go mod tidy en el directorio actual para que el go.mod
* incluya los imports de lo recien generado. No es fatal: si falla (p. ej.
* sin red), avisa y lo generado se conserva.
**/
func GoModTidy() {
	fmt.Println("\nEjecutando go mod tidy (la primera vez descarga dependencias y puede tardar unos minutos)...")
	out, err := exec.Command("go", "mod", "tidy").CombinedOutput()
	if err != nil {
		fmt.Printf("\nAviso: go mod tidy fallo (%v); ejecutalo a mano:\n%s\n", err, out)
		return
	}

	fmt.Println("go.mod actualizado")
}
