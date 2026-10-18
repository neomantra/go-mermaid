# go-mermaid demos

Two demos, both live at the project's demo site. This is a separate Go module
so go-booba and the UI dependencies stay out of the library's `go.mod`. It
needs Go 1.26.8 or newer.

## Terminal editor

A Bubble Tea program: edit Mermaid on the left, and the diagram redraws on the
right as you type. Run it from the repository root:

```sh
GOWORK=off go -C examples run ./editor
```

In Kitty, Ghostty, or WezTerm it draws with Kitty graphics; anywhere else, with
colored half blocks.

| Key | Action |
| --- | --- |
| typing | edit the source; the diagram redraws after a short pause |
| Ctrl+N | next example |
| Ctrl+T | next theme |
| Ctrl+G | switch between Kitty graphics and half blocks |
| Ctrl+C | quit |

A diagram that does not parse keeps the last good picture on screen, and the
status line shows the error with its line and column.

In the browser, the same program runs in a terminal through
[booba](https://github.com/NimbleMarkets/go-booba). The program advertises its
keys, and booba turns them into buttons above the terminal: Next example,
Theme, and Graphics, with More for the full list (Quit) and Keyboard to focus
the terminal. Each button sends the key you would have typed.

## SVG playground

`render` is the library compiled to WebAssembly. It exports one function,
`goMermaidRender(source, theme)`, which returns `{svg}` or `{error}`;
`web/playground.html` provides the editor, examples, themes, and download.
It only builds for `GOOS=js GOARCH=wasm`.

## Building the site

From the repository root (the site also needs Python 3 to serve, and Node to
check):

```sh
make serve-site    # build both demos into web/ and serve http://localhost:8000/
make check-site    # drive both WebAssembly demos under Node
make examples-test # test and vet the demos natively
make notices       # regenerate THIRD_PARTY_NOTICES.md after dependency changes
```

Browser builds use `wasm.work` at the root, which joins this module with the
library and selects the Bubble Tea and clipboard forks that js/wasm needs.
