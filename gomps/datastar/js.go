package datastar

import (
	_ "embed"

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
