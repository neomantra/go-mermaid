//go:build js && wasm

// The browser playground's renderer: it exports goMermaidRender(source,
// theme), which returns {svg} or {error}. The page owns everything else.
package main

import (
	"syscall/js"

	mermaid "github.com/zkrebbekx/go-mermaid"
)

func render(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return map[string]any{"error": "goMermaidRender needs Mermaid source"}
	}
	opts := []mermaid.Option{}
	if len(args) > 1 && args[1].Type() == js.TypeString {
		opts = append(opts, mermaid.WithTheme(mermaid.Theme(args[1].String())))
	}
	out, err := mermaid.Render(args[0].String(), opts...)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"svg": string(out)}
}

func main() {
	js.Global().Set("goMermaidRender", js.FuncOf(render))
	select {}
}
