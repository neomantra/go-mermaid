// The SVG playground: go-mermaid compiled to WebAssembly exports
// goMermaidRender(source, theme), and this page owns everything else.
export const examples = {
  flowchart: `flowchart LR
    A[Write Mermaid] --> B{Parses?}
    B -->|yes| C[Lay out]
    B -->|no| D[Report line and column]
    C --> E[SVG]`,
  sequence: `sequenceDiagram
    participant B as Browser
    participant W as WebAssembly
    B->>W: goMermaidRender(source)
    W-->>B: SVG string`,
  class: `classDiagram
    class Diagram {
      +Render(src) []byte
    }
    class Theme
    Diagram --> Theme`,
  state: `stateDiagram-v2
    [*] --> Typing
    Typing --> Rendered: valid
    Typing --> Error: invalid
    Error --> Typing
    Rendered --> [*]`,
  er: `erDiagram
    CUSTOMER ||--o{ ORDER : places
    ORDER ||--|{ LINE_ITEM : contains`,
  pie: `pie title Diagram types
    "Flowchart" : 45
    "Sequence" : 30
    "Other" : 25`,
  gantt: `gantt
    title Release
    dateFormat YYYY-MM-DD
    section Build
    Parser :a1, 2026-10-01, 7d
    Layout :after a1, 5d
    section Ship
    Demo :2026-10-14, 3d`,
  mindmap: `mindmap
  root((go-mermaid))
    Parse
    Lay out
    Render
      SVG
      PNG`,
};
export const themes = ['default', 'dark', 'neutral', 'forest', 'base'];

const $ = id => document.getElementById(id);

function remember(key, value) {
  try { if (value === undefined) return localStorage.getItem(key); localStorage.setItem(key, value); } catch { return null; }
  return null;
}

function start() {
  const source = $('source'), diagram = $('diagram'), status = $('status');
  const themeSelect = $('theme'), exampleSelect = $('example');
  for (const name of Object.keys(examples)) exampleSelect.add(new Option(name, name));
  for (const name of themes) themeSelect.add(new Option(name, name));
  themeSelect.value = themes.includes(remember('go-mermaid-theme')) ? remember('go-mermaid-theme') : 'default';
  source.value = examples.flowchart;
  let lastSVG = '', timer = 0;

  function draw() {
    if (!globalThis.goMermaidRender) return;
    const started = performance.now();
    const result = globalThis.goMermaidRender(source.value, themeSelect.value);
    const ms = (performance.now() - started).toFixed(1);
    diagram.dataset.theme = themeSelect.value;
    diagram.replaceChildren();
    if (result.error) {
      const p = document.createElement('p');
      p.className = 'error';
      p.textContent = result.error;
      diagram.append(p);
      status.textContent = `Not drawn after ${ms} ms. The previous download stays available.`;
      return;
    }
    lastSVG = result.svg;
    // Parse rather than assign innerHTML: the SVG is data, never markup to run.
    const doc = new DOMParser().parseFromString(result.svg, 'image/svg+xml');
    diagram.append(document.importNode(doc.documentElement, true));
    status.textContent = `Drawn in ${ms} ms, locally in your browser.`;
  }

  source.addEventListener('input', () => { clearTimeout(timer); timer = setTimeout(draw, 120); });
  themeSelect.addEventListener('change', () => { remember('go-mermaid-theme', themeSelect.value); draw(); });
  exampleSelect.addEventListener('change', () => { source.value = examples[exampleSelect.value]; draw(); });
  $('download').addEventListener('click', () => {
    if (!lastSVG) return;
    const a = document.createElement('a');
    a.href = URL.createObjectURL(new Blob([lastSVG], { type: 'image/svg+xml' }));
    a.download = 'diagram.svg';
    a.click();
    URL.revokeObjectURL(a.href);
  });

  const go = new Go();
  WebAssembly.instantiateStreaming(fetch('demos/render.wasm'), go.importObject)
    .then(({ instance }) => { go.run(instance); draw(); })
    .catch(error => {
      status.dataset.error = '';
      status.textContent = `Could not load WebAssembly: ${error.message}`;
    });
}

if (typeof document !== 'undefined') start();
