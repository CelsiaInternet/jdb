package main

import (
	"fmt"
	"os/exec"
)

var dependencies = []string{
	"github.com/joho/godotenv@v1.5.1",
	"github.com/bwmarrin/snowflake@v0.3.0",
	"github.com/google/uuid@v1.6.0",
	"github.com/matoous/go-nanoid/v2@v2.1.0",
	"github.com/oklog/ulid@v1.3.1",
	"golang.org/x/crypto/bcrypt@v0.37.0",
	"github.com/manifoldco/promptui@v0.9.0",
	"github.com/schollz/progressbar/v3@v3.18.0",
	"github.com/spf13/cobra@v1.9.1",
}

func main() {
	total := len(dependencies)
	for i, dep := range dependencies {
		p := (i + 1) * 100 / total
		fmt.Printf("\r[%-50s] %d%% Installing %s", progressBar(i+1, total, 50), p, dep)
		err := installLibrary(dep)
		if err != nil {
			fmt.Printf("\r[%-50s] %d%% Error installing %s - %v", progressBar(i+1, total, 50), p, dep, err)
			return
		}
	}

	fmt.Printf("\r[%-50s] %d%% ¡Completado!", progressBar(total, total, 50), total)
}

func installLibrary(library string) error {
	cmd := exec.Command("go", "get", library)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	return nil
}

func progressBar(current, total, width int) string {
	progress := int(float64(current) / float64(total) * float64(width))
	return fmt.Sprintf("%s%s", string(repeatRune('=', progress)), string(repeatRune(' ', width-progress)))
}

func repeatRune(char rune, count int) []rune {
	r := make([]rune, count)
	for i := 0; i < count; i++ {
		r[i] = char
	}
	return r
}
