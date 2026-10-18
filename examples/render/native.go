//go:build !(js && wasm)

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Fprintln(os.Stderr, "render is the browser playground's renderer; build it with GOOS=js GOARCH=wasm (make wasm-site)")
	os.Exit(2)
}
