package create

import (
	"encoding/json"
	"go/format"
	"strings"
	"testing"
)

/**
* TestGoTemplatesAreGofmtClean: Cada plantilla Go, con los argumentos que le
* pasa el generador, debe ser codigo valido y ya formateado con gofmt.
**/
func TestGoTemplatesAreGofmtClean(t *testing.T) {
	cases := []struct {
		name  string
		model string
		args  []string
	}{
		{"main.go", modelMain, []string{"github.com/acme/demo", "demo"}},
		{"service.go", modelService, []string{"github.com/acme/demo", "demo"}},
		{"api.go (schema)", modelDbApi, []string{"github.com/acme/demo", "demo"}},
		{"api.go", modelApi, []string{"github.com/acme/demo", "demo"}},
		{"event.go", modelEvent, []string{"demo"}},
		{"msg.go", modelMsg, []string{"demo"}},
		{"config.go", modelConfig, []string{"demo"}},
		{"rpc.go", modelhRpc, []string{"demo"}},
		{"controller.go (schema)", modelDbController, []string{"demo"}},
		{"controller.go", modelController, []string{"demo"}},
		{"model.go", modelModel, []string{"demo", "github.com/acme/demo", "Demo"}},
		{"schema_ventas.go", modelSchema, []string{"demo", "Ventas", "ventas"}},
		{"models/demo.go", modelData, []string{"demo", "Demo", "demo", "Ventas"}},
		{"router-demo.go (schema)", modelDbHandler, []string{"demo", "Demo", "demo", "github.com/acme/demo"}},
		{"router-demo.go", modelHandler, []string{"demo", "Demo", "demo"}},
		{"router.go (schema)", modelDbRouter, []string{"demo", "Demo"}},
		{"router.go", modelRouter, []string{"demo", "demo"}},
	}

	for _, c := range cases {
		src := fillTemplate(c.model, c.args...)
		got, err := format.Source([]byte(src))
		if err != nil {
			t.Errorf("%s: no es Go valido: %v", c.name, err)
			continue
		}

		if string(got) != src {
			t.Errorf("%s: no esta formateado con gofmt", c.name)
		}

		if strings.Contains(src, "envar.Envar") {
			t.Errorf("%s: usa envar.Envar* (obsoleto)", c.name)
		}
	}
}

/**
* TestRoutesMatchHandlers: Las rutas que el generador inserta en router.go
* apuntan a handlers que existen en la plantilla del modelo.
**/
func TestRoutesMatchHandlers(t *testing.T) {
	handler := fillTemplate(modelDbHandler, "demo", "Cliente", "cliente", "github.com/acme/demo")
	routes := fillTemplate(modelDbRoutes, "Cliente", "cliente")
	for _, line := range strings.Split(strings.TrimSpace(routes), "\n")[1:] {
		start := strings.Index(line, "rt.") + 3
		name := line[start : start+strings.Index(line[start:], ",")]
		if !strings.Contains(handler, "func (rt *Router) "+name+"(") {
			t.Errorf("la ruta usa rt.%s, que no esta en la plantilla del handler", name)
		}
	}
}

/**
* TestTemplatesHaveNoTokens: Las plantillas no deben traer JWT reales.
**/
func TestTemplatesHaveNoTokens(t *testing.T) {
	for name, model := range map[string]string{
		"restHttp":    restHttp,
		"modelDeploy": modelDeploy,
		"modelEnvar":  modelEnvar,
	} {
		if strings.Contains(model, "eyJhbGci") {
			t.Errorf("%s contiene un JWT", name)
		}
	}
}

/**
* TestEnvTemplate: El .env usa el nombre del servicio y trae las variables
* que el servicio generado y jdb leen al arrancar.
**/
func TestEnvTemplate(t *testing.T) {
	env := fillTemplate(modelEnvar, "demo")
	for _, want := range []string{
		"APP=demo\n", "PORT=3300\n", "PATH_URL=/api/demo\n", "RPC_PORT=", "DB_DRIVER=postgres\n",
		"DB_HOST=", "DB_NAME=", "APP_NAME=demo\n", "REDIS_HOST=", "NATS_HOST=", "SECRET=",
	} {
		if !strings.Contains(env, want) {
			t.Errorf(".env no contiene %q", want)
		}
	}

	if strings.Contains(env, "$") {
		t.Errorf(".env tiene un placeholder sin reemplazar")
	}
}

/**
* TestAgentsTemplates: Las reglas para agentes estan en la plantilla y el
* settings.json de Claude Code es JSON valido que bloquea commit y push.
**/
func TestAgentsTemplates(t *testing.T) {
	rules := fillTemplate(modelAgents, "demo")
	for _, want := range []string{
		rulesMarker,
		"entendible, mantenible y escalable por un humano",
		"Los commits los hace el usuario",
		"go run ./cmd/demo",
		"jdb/cmd/create go",
	} {
		if !strings.Contains(rules, want) {
			t.Errorf("agents.md no contiene %q", want)
		}
	}

	var settings struct {
		Permissions struct {
			Deny []string `json:"deny"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(modelClaudeSettings), &settings); err != nil {
		t.Fatalf("settings.json no es JSON valido: %v", err)
	}

	deny := strings.Join(settings.Permissions.Deny, " ")
	for _, want := range []string{"Bash(git commit:*)", "Bash(git push:*)"} {
		if !strings.Contains(deny, want) {
			t.Errorf("settings.json no bloquea %s", want)
		}
	}
}
