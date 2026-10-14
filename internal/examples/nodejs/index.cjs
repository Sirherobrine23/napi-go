'use strict';
const assert = require('node:assert/strict');
const path = require('node:path');
const fs = require('node:fs');
const crypto = require('node:crypto');
const addon = require('../../../build/nodejs.node');

async function main() {
  assert.equal(addon.join(['a', 'b', '..', 'c']), path.join('a', 'b', '..', 'c'));
  const hash = await addon.hashFile(__filename);
  assert.equal(hash, crypto.createHash('sha256').update(fs.readFileSync(__filename)).digest('hex'));
  assert.equal(await addon.later(), 'from the Node event loop');
  console.log({ path: addon.join(['a', 'c']), sha256: hash });
}
main().catch(error => { console.error(error); process.exitCode = 1; });
