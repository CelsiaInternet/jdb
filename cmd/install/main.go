package main

import (
	"fmt"
	"os/exec"
)

var dependencies = []string{
	"github.com/joho/godotenv/autoload",
	"github.com/bwmarrin/snowflake",
	"github.com/google/uuid",
	"github.com/matoous/go-nanoid/v2",
	"github.com/oklog/ulid",
	"golang.org/x/crypto/bcrypt",
	"golang.org/x/exp/slices",
	"github.com/manifoldco/promptui",
	"github.com/schollz/progressbar/v3",
	"github.com/spf13/cobra",
	"github.com/dimiro1/banner",
	"github.com/mattn/go-colorable",
	"github.com/celsiainternet/elvis@v1.1.202",
}

func main() {
	total := len(dependencies)
	for i, dep := range dependencies {
		p := (i + 1) * 100 / total
		fmt.Printf("\r[%-50s] %d%% Installing %s", progressBar(p, 100, 50), p, dep)
		err := installLibrary(dep)
		if err != nil {
			return
		}
	}

	fmt.Printf("\r[%-50s] %d%% ¡Completado!", progressBar(total, 100, 50), total)
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
