// Render every playground example in every theme with the real browser
// renderer under Node, and check that a broken diagram reports its position.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
import { examples, themes } from '../../web/playground.js';

vm.runInThisContext(await readFile('web/wasm_exec.js', 'utf8'));
const go = new Go();
const { instance } = await WebAssembly.instantiate(await readFile('web/demos/render.wasm'), go.importObject);
go.run(instance);
assert.equal(typeof globalThis.goMermaidRender, 'function', 'goMermaidRender exported');

let count = 0;
for (const [name, source] of Object.entries(examples)) {
  for (const theme of themes) {
    const result = goMermaidRender(source, theme);
    assert.equal(result.error, undefined, `${name} (${theme}): ${result.error}`);
    assert.match(result.svg, /^<svg[\s>]/, `${name} (${theme}): SVG output`);
    count++;
  }
}
const broken = goMermaidRender('flowchart LR\n  A -->', 'default');
assert.match(broken.error, /line 2 col \d+/, 'parse errors carry a position');
assert.match(goMermaidRender().error, /needs Mermaid source/, 'missing source is reported');
console.log(`Playground renderer: ${count} example renders and error reporting passed`);
process.exit(0);
