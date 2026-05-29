package main

import (
	"fmt"
	"os/exec"
)

var dependencies = []string{
	"github.com/joho/godotenv@v1.5.1",
	"github.com/bwmarrin/snowflake@v0.3.0",
	"github.com/oklog/ulid/v2@v2.1.1",
	"github.com/matoous/go-nanoid/v2@v2.1.0",
	"github.com/oklog/ulid@v1.3.1",
	"github.com/lib/pq@v1.10.9",
	"github.com/go-chi/chi@v1.5.5",
	"github.com/go-chi/chi/v5@v5.2.2",
	"github.com/google/uuid@v1.6.0",
	"golang.org/x/crypto@v0.37.0",
	"golang.org/x/exp@v0.0.0-20250408133849-7e4ce0ab07d0",
	"github.com/manifoldco/promptui@v0.9.0",
	"github.com/schollz/progressbar/v3@v3.18.0",
	"github.com/redis/go-redis/v9@v9.12.1",
	"github.com/spf13/cobra@v1.9.1",
	"github.com/nats-io/nats.go@v1.41.2",
	"github.com/golang-jwt/jwt/v4@v4.5.2",
	"github.com/robfig/cron/v3@v3.0.1",
	"github.com/rs/cors@v1.11.1",
	"github.com/rs/xid@v1.6.0",
	"github.com/go-sql-driver/mysql@v1.9.3",
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

	fmt.Printf("\r[%-50s] %d%% ¡Completado!", progressBar(total, total, 50), 100)
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
