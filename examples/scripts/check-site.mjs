// Validate the generated Pages artifact without a local HTTP server.
// Run from the repository root after `make wasm-site`.
import assert from 'node:assert/strict';
import { readFile, access } from 'node:fs/promises';
import { BoobaTerminal, installActionBar, installRendererHud, parseRendererFromURL } from '../../web/booba/booba.js';

// Link the actual browser runtime exports: syntax checks alone miss missing APIs.
for (const api of [BoobaTerminal, installActionBar, installRendererHud, parseRendererFromURL]) {
  assert.equal(typeof api, 'function', 'required browser runtime API');
}
for (const file of [
  'index.html', 'playground.html', 'site.css', 'site.js', 'playground.js', 'wasm_exec.js',
  'booba/booba.js', 'ghostty-web/ghostty-web.js', 'ghostty-web/ghostty-vt.wasm',
  'LICENSE', 'THIRD_PARTY_NOTICES.md',
]) await access(`web/${file}`);
for (const demo of ['editor', 'render']) {
  const bytes = await readFile(`web/demos/${demo}.wasm`);
  assert.deepEqual([...bytes.subarray(0, 4)], [0, 97, 115, 109], `${demo}: WASM header`);
  assert.ok(WebAssembly.validate(bytes), `${demo}: valid WASM`);
}
for (const page of ['index.html', 'playground.html']) {
  const html = await readFile(`web/${page}`, 'utf8');
  for (const [, path] of html.matchAll(/(?:src|href)="([^"?#]+)"/g)) {
    if (/^(https?:|\.\/$)/.test(path)) continue;
    await access(`web/${path}`);
  }
}
console.log('Site pages, links, runtime assets, and 2 WASM binaries checked.');
