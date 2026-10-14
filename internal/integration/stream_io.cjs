'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { Readable, Writable, Duplex, PassThrough } = require('node:stream');
const { finished, pipeline } = require('node:stream/promises');
const { once } = require('node:events');
const { Worker } = require('node:worker_threads');
const { spawnSync } = require('node:child_process');
const addon = require(process.env.NAPI_GO_ADDON);

const delay = ms => new Promise(resolve => setTimeout(resolve, ms));
async function waitFor(check) {
  const deadline = Date.now() + 5000;
  while (!check()) { assert(Date.now() < deadline, 'Go I/O cleanup did not complete'); await delay(5); }
}
async function collect(readable) {
  const chunks = [];
  for await (const chunk of readable) { assert(Buffer.isBuffer(chunk)); chunks.push(chunk); }
  return Buffer.concat(chunks);
}

test('FromReader creates a native readable and delivers ordered chunks and EOF', async () => {
  const control = addon.ioReadable('bytes', 'hello world', true);
  assert(control.stream instanceof Readable);
  assert.equal((await collect(control.stream)).toString(), 'hello world');
  await waitFor(() => control.stats().closes === 1);
  assert.equal(control.stats().maxBusy, 1);
});

test('Go readable does not read before demand and bounds prefetch under backpressure', async () => {
  const control = addon.ioReadable('bytes', 'x'.repeat(128), false);
  await delay(10);
  assert.equal(control.stats().reads, 0);
  control.stream.read(0);
  await waitFor(() => control.stream.readableLength === 4);
  const count = control.stats().reads;
  await delay(15);
  assert.equal(control.stats().reads, count);
  assert.equal(control.stream.readableLength, 4);
  assert.equal((await collect(control.stream)).length, 128);
  assert.equal(control.stats().closes, 0);
});

test('Go reader handles empty streams and n > 0 with EOF without losing bytes', async () => {
  assert.deepEqual(await collect(addon.ioReadable('bytes', '', false).stream), Buffer.alloc(0));
  const control = addon.ioReadable('bothEOF', '', true);
  assert.equal((await collect(control.stream)).toString(), 'last');
  assert.equal(control.stats().reads, 1);
  await waitFor(() => control.stats().closes === 1);
});

test('blocking Go reads run outside the Node event loop', async () => {
  const control = addon.ioReadable('slow', 'abcdefgh', true);
  let ticks = 0;
  const interval = setInterval(() => ticks++, 1);
  try { assert.equal((await collect(control.stream)).toString(), 'abcdefgh'); }
  finally { clearInterval(interval); }
  assert(ticks >= 5, `Node was blocked during Read (${ticks} ticks)`);
});

test('Go reader errors, panics, invalid counts and no progress reach native stream errors', async () => {
  for (const [mode, pattern] of [['error', /reader failure/], ['panic', /reader panic/], ['invalid', /invalid count/], ['emptyReads', /no data or error/]]) {
    const control = addon.ioReadable(mode, '', true);
    const chunks = [];
    control.stream.on('data', chunk => chunks.push(chunk));
    await assert.rejects(finished(control.stream), pattern);
    if (mode === 'error') assert.equal(Buffer.concat(chunks).toString(), 'last');
    await waitFor(() => control.stats().closes === 1);
  }
});

test('context cancellation interrupts an owned pipe read and settles close', async () => {
  const control = addon.ioReadable('pipe', '', true);
  const end = finished(control.stream);
  control.stream.resume();
  await waitFor(() => control.stats().busy === 1);
  control.cancel();
  await assert.rejects(end, error => error.code === 'ABORT_ERR');
  await waitFor(() => control.stats().busy === 0 && control.stats().closes === 1);
});

test('destroy preserves an original JavaScript error and closes an owned reader once', async () => {
  const control = addon.ioReadable('pipe', '', true);
  const failure = new Error('original destroy error');
  const end = finished(control.stream);
  control.stream.resume();
  await waitFor(() => control.stats().busy === 1);
  control.stream.destroy(failure);
  control.stream.destroy(failure);
  await assert.rejects(end, error => error === failure);
  await waitFor(() => control.stats().closes === 1 && control.stats().busy === 0);
});

test('context deadline destroys a blocked owned reader with ABORT_ERR', async () => {
  const control = addon.ioReadable('deadline', '', true);
  const end = finished(control.stream);
  control.stream.resume();
  await assert.rejects(end, error => error.code === 'ABORT_ERR' && /deadline exceeded/.test(error.message));
  await waitFor(() => control.stats().busy === 0 && control.stats().closes === 1);
});

test('FromWriter serializes Go writes and flushes/closes before native finish', async () => {
  const control = addon.ioWritable('bytes', true);
  assert(control.stream instanceof Writable);
  const completion = finished(control.stream);
  control.stream.write(Buffer.from([0, 255]));
  control.stream.write('héllo');
  control.stream.end('end');
  await completion;
  assert.deepEqual(Buffer.from(control.stats().data), Buffer.concat([Buffer.from([0, 255]), Buffer.from('hélloend')]));
  assert.equal(control.stats().maxBusy, 1);
  assert.equal(control.stats().flushes, 1);
  assert.equal(control.stats().closes, 1);
});

test('Go writer respects native backpressure and acknowledges after I/O completes', async () => {
  const control = addon.ioWritable('slow', false);
  const stream = control.stream;
  const events = [];
  let returned = false;
  const drain = once(stream, 'drain');
  const acknowledgement = new Promise((resolve, reject) => {
    assert.equal(stream.write('abc', error => { assert(returned); events.push('callback'); error ? reject(error) : resolve(); }), false);
    returned = true;
  });
  await delay(5);
  assert.deepEqual(control.stats().data, []);
  await Promise.all([acknowledgement, drain]);
  assert.equal(Buffer.from(control.stats().data).toString(), 'abc');
  stream.end();
  await finished(stream);
  assert.deepEqual(events, ['callback']);
  assert.equal(control.stats().closes, 0);
});

test('Go writer receives a copy of mutable JS buffers including empty writes', async () => {
  const control = addon.ioWritable('slow', false);
  const data = Buffer.from([1, 0, 255]);
  const completion = finished(control.stream);
  control.stream.write(data);
  data.fill(9);
  control.stream.end(Buffer.alloc(0));
  await completion;
  assert.deepEqual(control.stats().data, [1, 0, 255]);
});

test('Go write/flush/close failures and panics reject native writable completion', async () => {
  for (const [mode, pattern] of [['short', /short write/], ['error', /writer failure/], ['invalid', /invalid count/], ['panic', /writer panic/], ['flushError', /flush failure/], ['closeError', /close failure/]]) {
    const control = addon.ioWritable(mode, true);
    const end = finished(control.stream);
    control.stream.end('data');
    await assert.rejects(end, pattern);
    await waitFor(() => control.stats().closes === 1);
  }
});

test('cancellation settles an uninterruptible borrowed writer callback without waiting on Node', async () => {
  const control = addon.ioWritable('blocked', false);
  const end = finished(control.stream);
  let callbacks = 0;
  const written = new Promise(resolve => control.stream.write('blocked', error => { callbacks++; assert(error); resolve(); }));
  await waitFor(() => control.stats().busy === 1);
  control.cancel();
  await assert.rejects(end, error => error.code === 'ABORT_ERR');
  await written;
  assert.equal(control.stats().closes, 0);
  control.release(); // A borrowed endpoint remains the caller's responsibility.
  await waitFor(() => control.stats().busy === 0);
  assert.equal(callbacks, 1);
});

test('owned writer cancellation closes the endpoint and wakes a blocked write', async () => {
  const control = addon.ioWritable('blocked', true);
  const end = finished(control.stream);
  control.stream.write('blocked');
  await waitFor(() => control.stats().busy === 1);
  control.cancel();
  await assert.rejects(end, error => error.code === 'ABORT_ERR');
  await waitFor(() => control.stats().busy === 0 && control.stats().closes === 1);
});

test('cancel settles every buffered write callback exactly once', async () => {
  const control = addon.ioWritable('blocked', false);
  const end = finished(control.stream);
  const counts = [0, 0, 0];
  const writes = counts.map((_, index) => new Promise(resolve => control.stream.write('chunk', error => { counts[index]++; assert(error); resolve(); })));
  await waitFor(() => control.stats().busy === 1);
  control.cancel();
  await assert.rejects(end, error => error.code === 'ABORT_ERR');
  await Promise.all(writes);
  control.release();
  await waitFor(() => control.stats().busy === 0);
  assert.deepEqual(counts, [1, 1, 1]);
});

test('FromDuplex supports an io.Pipe pair with independent halves and writer EOF', async () => {
  const duplex = addon.ioDuplex();
  assert(duplex instanceof Duplex);
  const output = collect(duplex);
  duplex.write(Buffer.from([0, 255, 1]));
  duplex.end('end');
  assert.deepEqual(await output, Buffer.concat([Buffer.from([0, 255, 1]), Buffer.from('end')]));
  await finished(duplex);
});

test('FromReadWriter supports a shared net.Conn without closing it at writable EOF', async () => {
  const duplex = addon.ioReadWriter();
  assert(duplex instanceof Duplex);
  let data = Buffer.alloc(0);
  const received = new Promise(resolve => duplex.on('data', chunk => { data = Buffer.concat([data, chunk]); if (data.length === 4) resolve(); }));
  duplex.end('echo');
  await received;
  assert.equal(data.toString(), 'echo');
  const closed = once(duplex, 'close');
  duplex.destroy();
  await closed;
});

test('owned shared ReadWriter uses CloseWrite and closes the endpoint only once', async () => {
  const control = addon.ioHalfClose();
  const data = collect(control.stream);
  control.stream.end('half close');
  assert.equal((await data).toString(), 'half close');
  await finished(control.stream);
  assert.deepEqual(control.stats(), { closed: 1, halfClosed: 1 });
});

test('CopyTo connects a native Node readable to a Go io.Writer through pipeline', async () => {
  const readable = Readable.from([Buffer.from([0, 255]), Buffer.from('héllo')]);
  const output = addon.ioCopyTo(readable);
  assert(output instanceof Promise);
  assert.deepEqual(await output, Buffer.concat([Buffer.from([0, 255]), Buffer.from('héllo')]));
  await assert.rejects(addon.ioCopyError(Readable.from([Buffer.from('data')])), /writer failure/);
});

test('CopyFrom connects a Go io.Reader to a native Node writable through pipeline', async () => {
  const destination = new PassThrough();
  const chunks = collect(destination);
  const data = Buffer.from([0, 1, 255, 2, 0, 7]);
  await addon.ioCopyFrom(destination, data);
  assert.deepEqual(await chunks, data);
});

test('Go streams work in native Node pipelines with error propagation', async () => {
  const source = addon.ioReadable('bytes', 'pipeline', true);
  const destination = addon.ioWritable('bytes', true);
  await pipeline(source.stream, destination.stream);
  assert.equal(Buffer.from(destination.stats().data).toString(), 'pipeline');
  const failed = addon.ioReadable('error', '', true);
  const sink = addon.ioWritable('bytes', true);
  await assert.rejects(pipeline(failed.stream, sink.stream), /reader failure/);
  await waitFor(() => sink.stats().closes === 1);
});

test('idle Go I/O streams do not keep the event loop alive', () => {
  const child = spawnSync(process.execPath, ['-e', `
    const addon=require(process.env.NAPI_GO_ADDON);
    addon.ioReadable('bytes','unused',true);
    addon.ioWritable('bytes',true);
  `], { env: process.env, timeout: 5000 });
  assert.equal(child.status, 0, child.stderr?.toString());
});

test('weak idle stream references allow forced GC and owned resource cleanup', async () => {
  const before = addon.ioClosedCount();
  let control = addon.ioReadable('bytes', 'unused', true);
  const weak = new WeakRef(control.stream);
  control = null;
  await delay(0);
  for (let i = 0; i < 40 && addon.ioClosedCount() === before; i++) { global.gc(); await delay(5); }
  assert.equal(weak.deref(), undefined);
  assert.equal(addon.ioClosedCount(), before + 1);
});

test('Worker termination cleans owned I/O while a Go pipe read is blocked', async () => {
  const before = addon.ioClosedCount();
  const worker = new Worker(`
    const {parentPort}=require('node:worker_threads');
    const addon=require(process.env.NAPI_GO_ADDON);
    const control=addon.ioReadable('pipe','',true);
    control.stream.resume();
    setTimeout(()=>parentPort.postMessage('ready'),20);
  `, { eval: true });
  await once(worker, 'message');
  assert.equal(await worker.terminate(), 1);
  await waitFor(() => addon.ioClosedCount() > before);
});

test('Go I/O constructors validate endpoints, options and already-canceled contexts', () => {
  for (const mode of ['reader', 'writer', 'duplex', 'rw', 'options', 'twoOptions', 'context']) assert.throws(() => addon.ioInvalid(mode));
});
