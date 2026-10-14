'use strict';
const assert = require('node:assert/strict');
const { Readable } = require('node:stream');
const { finished } = require('node:stream/promises');
const { createHash } = require('node:crypto');
const addon = require('../../../build/stream_io.node');

async function collect(stream) {
  const chunks = [];
  for await (const chunk of stream) chunks.push(chunk);
  return Buffer.concat(chunks);
}
async function main() {
  const text = (await collect(addon.read())).toString();
  assert.equal(text, 'hello from Go\n');
  const hash = await addon.digest(Readable.from(['alpha', 'beta']));
  assert.deepEqual(hash, createHash('sha256').update('alphabeta').digest());
  const duplex = addon.echo();
  const echo = collect(duplex);
  duplex.end(Buffer.from([0, 1, 255]));
  assert.deepEqual(await echo, Buffer.from([0, 1, 255]));
  await finished(duplex);
  console.log({ text: text.trim(), sha256: hash.toString('hex'), echo: [0, 1, 255] });
}
main().catch(error => { console.error(error); process.exitCode = 1; });
