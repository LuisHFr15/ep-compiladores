package mips

import (
	"embed"
	"strings"
)

//go:embed runtime/*.s
var runtimeFS embed.FS

var runtimeOrder = []string{"core", "str", "num", "edit", "inspect", "file"}

func runtimeText(need map[string]bool) string {
	var b strings.Builder
	for _, name := range runtimeOrder {
		if !need[name] {
			continue
		}
		data, err := runtimeFS.ReadFile("runtime/" + name + ".s")
		if err != nil {
			continue
		}
		b.Write(data)
		b.WriteString("\n")
	}
	return b.String()
}
