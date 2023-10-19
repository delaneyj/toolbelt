package datastar

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/delaneyj/toolbelt/gomps"
)

//go:embed latest.js
var latestJS string

func CDN() gomps.NODE {
	return gomps.SCRIPT(
		gomps.TYPE("module"),
		gomps.RAW(`
import  { runDatastarWithAllPlugins }  from 'https://cdn.jsdelivr.net/npm/@sudodevnull/datastar/+esm'
runDatastarWithAllPlugins()
		`),
	)
}

func Latest() gomps.NODE {
	return gomps.SCRIPT(
		gomps.TYPE("module"),
		gomps.RAW(latestJS),
	)
}

func LatestRunAllPlugins() gomps.NODE {
	return gomps.GRP(
		Latest(),
		gomps.SCRIPT(
			gomps.TYPE("module"),
			gomps.RAW(`runDatastarWithAllPlugins()`),
		),
	)
}

func AsyncThunk(asyncFuncLines string) string {
	lines := strings.Split(asyncFuncLines, "\n")
	for i, line := range lines {
		lines[i] = fmt.Sprintf("\t%s", line)
	}
	return fmt.Sprintf(
		"(async()=> {\n%s})()",
		strings.Join(lines, "\n"),
	)
}
