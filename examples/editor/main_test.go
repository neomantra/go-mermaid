package main

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

// draw runs one debounced render the way the event loop would.
func draw(t *testing.T, m *model) rendered {
	t.Helper()
	m.schedule()
	m.Update(renderTick{m.generation})
	msg, ok := m.render()().(rendered)
	if !ok {
		t.Fatal("render did not return a rendered message")
	}
	m.Update(msg)
	return msg
}

func TestEditorDrawsEveryExampleInEveryTheme(t *testing.T) {
	m := newModel()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	for i, ex := range examples {
		m.example = i
		m.editor.SetValue(ex.source)
		for th := range themes {
			m.theme = th
			if got := draw(t, m); got.err != nil || got.image == nil {
				t.Fatalf("%s in %s: %v", ex.name, themes[th], got.err)
			}
		}
	}
}

func TestEditorKeepsLastDiagramOnError(t *testing.T) {
	m := newModel()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	draw(t, m)
	m.editor.SetValue("flowchart LR\n  A -->")
	if got := draw(t, m); got.err == nil {
		t.Fatal("a dangling arrow should not parse")
	}
	if !m.failed || !strings.Contains(m.status, "parse error") {
		t.Fatalf("status = %q, want the parse error", m.status)
	}
}

func TestEditorDropsStaleRenders(t *testing.T) {
	m := newModel()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	old := m.generation
	m.schedule() // A newer edit arrives before the old render finishes.
	m.Update(rendered{generation: old, err: errTest})
	if m.failed {
		t.Fatal("a stale render replaced the status")
	}
}

func TestEditorLayoutStacksWhenNarrow(t *testing.T) {
	m := newModel()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	e, p := m.panes()
	if e.editorHeight != p.height {
		t.Fatalf("wide layout should sit side by side: %+v %+v", e, p)
	}
	m.Update(tea.WindowSizeMsg{Width: 60, Height: 40})
	e, p = m.panes()
	if p.width != 60 || e.editorHeight+p.height+1 != 40-chromeRows {
		t.Fatalf("narrow layout should stack: %+v %+v", e, p)
	}
	if v := m.View().Content; !strings.Contains(v, "Ctrl+N example") {
		t.Fatal("help line missing")
	}
}

type testError string

func (e testError) Error() string { return string(e) }

const errTest = testError("stale")
