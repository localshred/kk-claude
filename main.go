// kk-claude/main.go
package main

import (
	"os"

	"github.com/localshred/kk-claude/cmd"
	"github.com/localshred/koolkit/pkg/plugin"
)

func pluginMeta() plugin.Meta {
	return plugin.Meta{
		Describe: "Pick/resume a saved Claude session, or start a named new one",
		Requires: []string{"claude"},
	}
}

func main() {
	if handled, code := plugin.Handle(os.Args[1:], pluginMeta(), os.Stdout); handled {
		os.Exit(code)
	}
	os.Exit(cmd.Execute())
}
