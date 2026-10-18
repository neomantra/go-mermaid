// A live Mermaid editor for the terminal: edit the source on the left and the
// diagram redraws on the right. The same program runs natively and, through
// booba, in the browser.
package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	mermaid "github.com/zkrebbekx/go-mermaid"
	"github.com/zkrebbekx/go-mermaid/examples/internal/demorun"
	"github.com/zkrebbekx/go-mermaid/raster"
	picture "nimbleterminal.dev/ntpicture"
)

// example is one sample diagram.
type example struct{ name, source string }

var examples = []example{
	{"flowchart", "flowchart LR\n    A[Edit Mermaid] --> B{Parses?}\n    B -->|yes| C[Lay out]\n    B -->|no| D[Show the error]\n    C --> E[Draw]"},
	{"sequence", "sequenceDiagram\n    participant E as Editor\n    participant M as go-mermaid\n    E->>M: Render(source)\n    M-->>E: SVG\n    E->>E: rasterize and show"},
	{"class", "classDiagram\n    class Diagram {\n      +Render(src) []byte\n    }\n    class Theme\n    Diagram --> Theme"},
	{"state", "stateDiagram-v2\n    [*] --> Typing\n    Typing --> Drawn: valid\n    Typing --> Error: invalid\n    Error --> Typing\n    Drawn --> [*]"},
	{"pie", "pie title Diagram types\n    \"Flowchart\" : 45\n    \"Sequence\" : 30\n    \"Other\" : 25"},
	{"gantt", "gantt\n    title Release\n    dateFormat YYYY-MM-DD\n    section Build\n    Parser :a1, 2026-10-01, 7d\n    Layout :after a1, 5d\n    section Ship\n    Demo :2026-10-14, 3d"},
}

var themes = []mermaid.Theme{mermaid.Dark, mermaid.Default, mermaid.Neutral, mermaid.Forest, mermaid.Base}

// Edits redraw after a short pause, so typing stays responsive.
const debounce = 150 * time.Millisecond

// Rows above and below the panes: title, blank, status, help.
const chromeRows = 4

type (
	renderTick struct{ generation int }
	rendered   struct {
		generation int
		image      image.Image
		elapsed    time.Duration
		err        error
	}
)

type model struct {
	editor        textarea.Model
	pic           *picture.Picture
	example       int
	theme         int
	generation    int
	width, height int
	status        string
	failed        bool
	closed        bool
}

func newModel() *model {
	editor := textarea.New()
	editor.ShowLineNumbers = false
	editor.Prompt = ""
	editor.CharLimit = 16 << 10
	editor.MaxHeight = 0
	m := &model{
		editor: editor,
		pic: picture.New(picture.Options{
			Render:   picture.Config{Fit: picture.FitContain},
			AutoMode: true,
		}),
	}
	m.editor.SetValue(examples[0].source)
	return m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.pic.Init(), m.editor.Focus(), m.schedule())
}

// schedule asks for a redraw of the current source after the debounce.
func (m *model) schedule() tea.Cmd {
	m.generation++
	g := m.generation
	return tea.Tick(debounce, func(time.Time) tea.Msg { return renderTick{g} })
}

// render draws the source off the event loop.
func (m *model) render() tea.Cmd {
	g, source, theme := m.generation, m.editor.Value(), themes[m.theme]
	return func() tea.Msg {
		start := time.Now()
		data, err := raster.PNG(source, 2, mermaid.WithTheme(theme))
		if err != nil {
			return rendered{generation: g, err: err}
		}
		img, err := png.Decode(bytes.NewReader(data))
		return rendered{generation: g, image: img, elapsed: time.Since(start), err: err}
	}
}

func (m *model) layout() tea.Cmd {
	w, h := m.panes()
	m.editor.SetWidth(w.editorWidth)
	m.editor.SetHeight(w.editorHeight)
	m.unscroll()
	return m.pic.SetSize(h.width, h.height)
}

// unscroll shows the source from its first line, keeping the cursor in view.
// A text area that grows keeps the scroll offset it needed when it was small;
// browser terminals start small and then fit the page, which would otherwise
// leave the top of the source scrolled away.
func (m *model) unscroll() {
	row, col := m.editor.Line(), m.editor.Column()
	m.editor.MoveToBegin()
	// CursorDown moves one wrapped row at a time; stop if it cannot advance.
	for range 1 << 16 {
		if m.editor.Line() >= row {
			break
		}
		before := m.editor.LineInfo()
		line := m.editor.Line()
		m.editor.CursorDown()
		if m.editor.Line() == line && m.editor.LineInfo() == before {
			break
		}
	}
	m.editor.SetCursorColumn(col)
}

type editorSize struct{ editorWidth, editorHeight int }
type pictureSize struct{ width, height int }

// panes splits the screen side by side when it is wide, stacked when narrow.
func (m *model) panes() (editorSize, pictureSize) {
	body := max(1, m.height-chromeRows)
	if m.width >= 90 {
		left := max(28, m.width*2/5)
		return editorSize{left - 2, body}, pictureSize{max(1, m.width-left-1), body}
	}
	top := min(10, max(3, body/3))
	return editorSize{max(1, m.width-1), top}, pictureSize{m.width, max(1, body-top-1)}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.closed {
		return m, nil
	}
	cmds := []tea.Cmd{m.pic.Update(msg)}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		cmds = append(cmds, m.layout())
		return m, tea.Batch(cmds...)
	case renderTick:
		if msg.generation == m.generation {
			m.status = "Drawing…"
			cmds = append(cmds, m.render())
		}
		return m, tea.Batch(cmds...)
	case rendered:
		if msg.generation != m.generation {
			return m, tea.Batch(cmds...) // A later edit replaced this one.
		}
		if msg.err != nil {
			// Keep the last good diagram; say what is wrong with this one.
			m.status, m.failed = msg.err.Error(), true
			return m, tea.Batch(cmds...)
		}
		m.status, m.failed = fmt.Sprintf("Drew %d×%d in %s · theme %s", msg.image.Bounds().Dx(), msg.image.Bounds().Dy(), msg.elapsed.Round(time.Millisecond), themes[m.theme]), false
		cmds = append(cmds, m.pic.SetImageImmutable(msg.image))
		return m, tea.Batch(cmds...)
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			m.closed = true
			return m, tea.Sequence(m.pic.Close(), tea.Quit)
		case "ctrl+n":
			m.example = (m.example + 1) % len(examples)
			m.editor.SetValue(examples[m.example].source)
			return m, tea.Batch(append(cmds, m.schedule())...)
		case "ctrl+t":
			m.theme = (m.theme + 1) % len(themes)
			return m, tea.Batch(append(cmds, m.schedule())...)
		case "ctrl+g":
			return m, tea.Batch(append(cmds, m.pic.ToggleMode())...)
		}
	}
	before := m.editor.Value()
	var cmd tea.Cmd
	m.editor, cmd = m.editor.Update(msg)
	cmds = append(cmds, cmd)
	if m.editor.Value() != before {
		cmds = append(cmds, m.schedule())
	}
	return m, tea.Batch(cmds...)
}

func (m *model) View() tea.View {
	if m.closed {
		return tea.NewView("")
	}
	v := tea.NewView("")
	v.AltScreen = true
	if m.width < 40 || m.height < 12 {
		v.SetContent(ansi.Truncate("Enlarge to at least 40 × 12", m.width, "…"))
		return v
	}
	clip := func(s string) string { return ansi.Truncate(s, m.width, "…") }
	accent := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	statusStyle := muted
	if m.failed {
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	}
	e, p := m.panes()
	editor := lipgloss.NewStyle().Width(e.editorWidth + 1).Height(e.editorHeight).Render(m.editor.View())
	diagram := lipgloss.NewStyle().Width(p.width).Height(p.height).Render(m.pic.View())
	var body string
	if m.width >= 90 {
		rule := strings.TrimSuffix(strings.Repeat("│\n", e.editorHeight), "\n")
		body = lipgloss.JoinHorizontal(lipgloss.Top, editor, muted.Render(rule), diagram)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, editor, muted.Render(strings.Repeat("─", m.width)), diagram)
	}
	title := accent.Render("go-mermaid") + muted.Render(clip(fmt.Sprintf(" · example %s · %s", examples[m.example].name, "pure Go, no browser engine")))
	help := clip("Ctrl+N example · Ctrl+T theme · Ctrl+G graphics · Ctrl+C quit")
	v.SetContent(title + "\n\n" + body + "\n" + clip(statusStyle.Render(m.status)) + "\n" + muted.Render(help))
	return v
}

func main() {
	if err := demorun.Run(newModel(), "editor", "Live Mermaid editor", demorun.Keys{
		Primary: []key.Binding{
			demorun.Binding("ctrl+n", "Next example"),
			demorun.Binding("ctrl+t", "Theme"),
			demorun.Binding("ctrl+g", "Graphics"),
		},
		More: []key.Binding{demorun.Binding("ctrl+c", "Quit")},
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
