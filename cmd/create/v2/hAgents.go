package create

import (
	_ "embed"
	"os"
	"strings"

	"github.com/celsiainternet/elvis/file"
	"github.com/celsiainternet/elvis/strs"
)

//go:embed template/agents.md
var modelAgents string

//go:embed template/settings.json
var modelClaudeSettings string

/**
* rulesMarker: Marca que identifica las reglas del proyecto en CLAUDE.md y
* AGENTS.md; evita agregarlas dos veces.
**/
const rulesMarker = "<!-- jdb:project-rules -->"

/**
* MakeAgents: Deja las reglas del proyecto para los agentes en CLAUDE.md y
* AGENTS.md, y crea .claude/settings.json, que impide a Claude Code hacer
* commit o push. Si los .md ya existen (p. ej. los creo cmd/install con el
* bloque de guias), antepone las reglas y conserva su contenido.
* @param name string
* @return error
**/
func MakeAgents(name string) error {
	err := upsertRules("CLAUDE.md", name)
	if err != nil {
		return err
	}

	err = upsertRules("AGENTS.md", name)
	if err != nil {
		return err
	}

	path, err := file.MakeFolder(".claude")
	if err != nil {
		return err
	}

	_, err = file.MakeFile(path, "settings.json", modelClaudeSettings)
	return err
}

/**
* upsertRules: Crea fileName con las reglas o, si ya existe sin ellas, las
* antepone al contenido actual.
* @param fileName, name string
* @return error
**/
func upsertRules(fileName, name string) error {
	rules := strs.Format("# %s\n\n%s", fileName, fillTemplate(modelAgents, name))
	path := "./" + fileName
	if !file.ExistPath(path) {
		return os.WriteFile(path, []byte(rules), 0644)
	}

	current, err := file.ReadFile(path)
	if err != nil {
		return err
	}

	if strings.Contains(current, rulesMarker) {
		return nil
	}

	return os.WriteFile(path, []byte(rules+"\n"+current), 0644)
}
