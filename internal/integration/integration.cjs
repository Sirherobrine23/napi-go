'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { Worker } = require('node:worker_threads');
const { spawnSync } = require('node:child_process');
const addon = require(process.env.NAPI_GO_ADDON);

test('status errors match by code through stack traces and wrapped errors', () => {
  assert.equal(addon.statusErrors(), undefined);
});

async function until(check) {
  for (let i = 0; i < 200; i++) {
    if (check()) return;
    if (global.gc) global.gc();
    await new Promise(resolve => setTimeout(resolve, 5));
  }
  assert.fail('condition did not complete');
}

test('classes: constructor, native new, instance state, accessors, static and data properties', () => {
  const c = new addon.Counter(3);
  assert.equal(c.value, 3);
  assert.equal(c.increment(), 4);
  c.value = 10;
  assert.equal(c.value, 10);
  assert.equal(c.label, 'counter');
  assert.equal(addon.Counter.version(), '1');
  assert.equal(c.version, undefined);
  assert(c instanceof addon.Counter);
  assert.equal(addon.zero.value, 0);
  assert.equal(addon.one.value, 1);
  assert(new addon.Empty() instanceof addon.Empty);
  assert.throws(() => addon.Counter(), TypeError);
  assert.throws(() => c.fail(), /class error/);
  assert.throws(() => c.panic(), /class panic/);
  assert.equal(Object.getOwnPropertyDescriptor(addon.Counter.prototype, 'increment').enumerable, false);
  class Derived extends addon.Counter {}
  assert(new Derived(8) instanceof Derived);
  assert.equal(new Derived(8).increment(), 9);
});

test('properties: symbol names, methods, getters and empty descriptor list', () => {
  const symbol = Symbol('native');
  const obj = addon.properties(symbol);
  assert.equal(obj[symbol], 23);
  assert.equal(obj.method(), 7);
  assert.equal(obj.getter, 8);
});

test('wraps: remove transfers ownership, prevents finalizer, allows rewrap', () => {
  const c = new addon.Counter(9);
  assert.equal(addon.detach(c), 9);
  assert.throws(() => addon.unwrap(c));
  assert.throws(() => addon.unwrap({}));
  addon.rewrap(c);
  assert.equal(c.value, 42);
  assert.throws(() => addon.rewrap(c));
});

test('callbacks survive constructor/prototype collection and extraction', async () => {
  let C = addon.makeClass();
  let c = new C(23);
  const method = C.prototype.increment;
  const getter = Object.getOwnPropertyDescriptor(C.prototype, 'value').get;
  C = null;
  for (let i = 0; i < 5; i++) { global.gc(); await new Promise(r => setImmediate(r)); }
  assert.equal(method.call(c), 24);
  assert.equal(getter.call(c), 24);
  c = null;
  global.gc();
});

test('wrapped state finalizes on GC', async () => {
  const before = addon.stats().wrapped;
  (() => { for (let i = 0; i < 20; i++) new addon.Counter(i); })();
  await until(() => addon.stats().wrapped >= before + 20);
});

test('threadsafe: bounded blocking queue, concurrent producers and final release', async () => {
  const values = [];
  const control = addon.stream(value => values.push(value), 300, 4);
  await until(() => control.closed());
  assert.deepEqual(values.sort((a, b) => a - b), Array.from({ length: 300 }, (_, i) => i));
  assert(control.rejects());
});

test('threadsafe: nonblocking queue reports full and drains accepted data', async () => {
  const values = [];
  const control = addon.queue(value => values.push(value));
  await until(() => control.closed());
  assert.deepEqual(values, [1]);
  assert(control.rejects());
});

test('threadsafe: custom callback with nil JavaScript function and Go context', async () => {
  const before = addon.stats();
  const control = addon.custom();
  await until(() => control.closed());
  assert.equal(addon.stats().delivered, before.delivered + 1);
  assert.equal(addon.stats().finalized, before.finalized + 1);
});

test('threadsafe: abort discards data and frees all abandoned Go handles', async () => {
  const before = addon.stats();
  const control = addon.abort();
  await until(() => control.closed());
  assert.equal(addon.stats().delivered, before.delivered);
  await until(() => addon.stats().collected >= before.collected + 1000);
});

test('threadsafe: abort wakes a producer blocked on a full queue', async () => {
  const control = addon.abortBlocking();
  await until(() => control.closed());
  assert(control.rejects());
});

test('threadsafe: no user finalizer, nil payload and Go Callback compatibility', async () => {
  let argc;
  const control = addon.noFinalizer((...args) => { argc = args.length; });
  await until(() => control.closed());
  assert.equal(argc, 0);
  assert(control.rejects());
  const before = addon.stats().delivered;
  const goControl = addon.goCallback();
  await until(() => goControl.closed());
  assert.equal(addon.stats().delivered, before + 1);
});

test('threadsafe: Go values are converted on the Node thread', async () => {
  const values = [];
  const control = addon.converted(value => values.push(value));
  await until(() => control.closed());
  assert.deepEqual(values, [23, 1.5, { value: 5, nil: null }]);
});

test('validation and idempotent event-loop Ref/Unref', () => addon.validation(23));

test('threadsafe: concurrent acquire/release never closes another producer', async () => {
  const control = addon.acquisitions();
  await until(() => control.closed());
  assert(control.rejects());
});

test('Unref allows the process to exit with a live native function', () => {
  const child = spawnSync(process.execPath, ['-e', 'require(process.env.NAPI_GO_ADDON).unref()'], { env: process.env, timeout: 5000 });
  assert.equal(child.status, 0, child.stderr?.toString());
});

test('independent Worker environments create and clean up callbacks', async () => {
  await Promise.all(Array.from({ length: 4 }, () => new Promise((resolve, reject) => {
    const worker = new Worker(`
      const a = require(process.env.NAPI_GO_ADDON);
      const c = new a.Counter(1);
      if (c.increment() !== 2) throw Error('bad class state');
      a.unref();
    `, { eval: true });
    worker.once('error', reject);
    worker.once('exit', code => code === 0 ? resolve() : reject(Error('worker exit ' + code)));
  })));
});

test('Worker termination stops active Go producers and native queue cleanup', async () => {
  const worker = new Worker(`
    const { parentPort } = require('node:worker_threads');
    require(process.env.NAPI_GO_ADDON).workerActive();
    parentPort.postMessage('ready');
    setInterval(() => {}, 1000);
  `, { eval: true });
  await new Promise((resolve, reject) => {
    worker.once('message', resolve);
    worker.once('error', reject);
  });
  assert.equal(await worker.terminate(), 1);
});
