package agentsguide

import (
	"embed"

	elvisguide "github.com/celsiainternet/elvis/agentsguide"
)

//go:embed JDB.md
var files embed.FS

/**
* Install: Installs the jdb agents guide (JDB.md) into the current project:
* writes it to .claude/guides/jdb/ and a managed block in CLAUDE.md and
* AGENTS.md that points to it. Running it again replaces both with the ones of
* the installed jdb version.
* @return error
**/
func Install() error {
	return elvisguide.InstallGuide("jdb", "github.com/celsiainternet/jdb", files, []string{"JDB.md"})
}
