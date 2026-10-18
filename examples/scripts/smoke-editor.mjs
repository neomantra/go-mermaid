// Exercise the real terminal-editor binary through the booba bridge under
// Node. This checks application behavior, not browser rendering.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import vm from 'node:vm';
import { setTimeout as delay } from 'node:timers/promises';

vm.runInThisContext(await readFile('web/wasm_exec.js', 'utf8'));
const go = new Go();
let exitCode;
go.exit = code => { exitCode = code; };
const { instance } = await WebAssembly.instantiate(await readFile('web/demos/editor.wasm'), go.importObject);
const timeout = setTimeout(() => { console.error('editor timed out'); process.exit(1); }, 30000);
const running = go.run(instance);

async function expectOutput(pattern) {
  let output = '';
  for (let n = 0; n < 200 && !pattern.test(output); n++) {
    await delay(25);
    output += bubbletea_read();
  }
  assert.match(output, pattern, `expected terminal output: ${pattern}`);
  return output;
}

try {
  for (let n = 0; typeof globalThis.bubbletea_read !== 'function' && n < 100; n++) await delay(20);
  assert.equal(typeof globalThis.bubbletea_read, 'function');
  bubbletea_resize(120, 40);
  await expectOutput(/go-mermaid/);
  // Bubble Tea redraws only changed cells: when "Drawing…" becomes "Drew …"
  // the shared "Dr" is not resent, so match the part that always is.
  await expectOutput(/\d+×\d+ in [\d.]+m?s/);
  // Negotiate the same program-status support as the browser action bar.
  bubbletea_write('\x1b]7501;?\x1b\\');
  const advertised = await expectOutput(/:msg=/);
  const encoded = advertised.match(/:msg=([A-Za-z0-9+/=]+)/)?.[1];
  assert.ok(encoded, 'action status payload');
  const description = Buffer.from(encoded, 'base64').toString('utf8');
  for (const label of ['Next example', 'Theme', 'Graphics', 'Quit']) {
    assert.ok(description.includes(label), `advertised control: ${label}`);
  }
  // Bubble Tea sends only changed cells, so match text each redraw sends whole.
  bubbletea_write('\x0e'); // Ctrl+N: the sequence example
  await expectOutput(/rasterize and show[\s\S]*▄/); // New source, then new picture cells.
  bubbletea_write('\x14'); // Ctrl+T: next theme repaints the picture.
  await expectOutput(/▄/);
  bubbletea_write('\r    E->>'); // A message with no receiver: an error; the last diagram stays.
  await expectOutput(/sender and receiver/);
  bubbletea_resize(30, 8);
  await expectOutput(/Enlarge/);
  bubbletea_resize(120, 40);
  bubbletea_write('\x03');
  await running;
  assert.equal(exitCode, 0);
  console.log('Terminal editor: startup, controls, examples, themes, errors, resize, and quit passed');
} finally {
  clearTimeout(timeout);
}
process.exit(0);
