'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { EventEmitter } = require('node:events');
const { Worker, threadId } = require('node:worker_threads');
const addon = require(process.env.NAPI_GO_ADDON);

async function closed(control) {
  const deadline = Date.now() + 5000;
  while (!control.closed()) {
    assert(Date.now() < deadline, 'native event queue did not finalize');
    await new Promise(resolve => setTimeout(resolve, 5));
  }
}

function relay(emitter) {
  return ({ name, args }) => emitter.emit(name, ...args);
}

function collect(emitter, count) {
  const progress = [];
  const order = [];
  let done;
  let receivers = true;
  let jsThreads = true;
  emitter.on('progress', function (index, data) {
    receivers &&= this === emitter;
    jsThreads &&= threadId === 0;
    progress.push([index, data]);
    order.push('progress');
  });
  emitter.once('done', function (data) {
    receivers &&= this === emitter;
    jsThreads &&= threadId === 0;
    done = data;
    order.push('done');
  });
  return () => {
    assert.equal(progress.length, count);
    assert.equal(receivers, true, 'emit lost the EventEmitter receiver');
    assert.equal(jsThreads, true, 'listeners ran outside the Node thread');
    assert.equal(order.at(-1), 'done');
    const sum = count * (count - 1) * (2 * count - 1) / 6;
    assert.deepEqual(done, { count, sum });
    const indexes = progress.map(([index]) => index).sort((a, b) => a - b);
    assert.deepEqual(indexes, Array.from({ length: count }, (_, i) => i));
    for (const [index, data] of progress) assert.equal(data.value, index * index);
    return progress;
  };
}

test('EventEmitter receives background progress and completion after start returns', async () => {
  const emitter = new EventEmitter();
  const verify = collect(emitter, 96);
  let returned = false;
  emitter.on('progress', () => assert.equal(returned, true));
  const before = addon.stats().finalized;
  const control = addon.events(relay(emitter), 96, 1);
  returned = true;
  await closed(control);
  const progress = verify();
  assert.deepEqual(progress.map(([index]) => index), Array.from({ length: 96 }, (_, i) => i));
  assert.equal(addon.stats().finalized, before + 1);
  assert(control.rejects());
});

test('concurrent Go producers deliver every item before the done event', async () => {
  const emitter = new EventEmitter();
  const verify = collect(emitter, 300);
  const control = addon.events(relay(emitter), 300, 4);
  await closed(control);
  const progress = verify();
  for (let producer = 0; producer < 4; producer++) {
    const indexes = progress.filter(([, data]) => data.producer === producer).map(([index]) => index);
    assert.deepEqual(indexes, Array.from({ length: 75 }, (_, i) => producer + 4 * i));
  }
});

test('bound emitter.emit accepts native event names and multiple arguments directly', async () => {
  const emitter = new EventEmitter();
  const verify = collect(emitter, 32);
  const control = addon.eventsBound(emitter.emit.bind(emitter), 32, 2);
  await closed(control);
  verify();
});

test('an empty background task still emits completion exactly once', async () => {
  const emitter = new EventEmitter();
  const verify = collect(emitter, 0);
  let completions = 0;
  emitter.on('done', () => completions++);
  const control = addon.events(relay(emitter), 0, 4);
  await closed(control);
  verify();
  assert.equal(completions, 1);
});

test('background failure reaches an explicitly registered error listener', async () => {
  const emitter = new EventEmitter();
  const errors = [];
  emitter.on('error', function (message) {
    assert.equal(this, emitter);
    errors.push(message);
  });
  const control = addon.eventFailure(relay(emitter));
  await closed(control);
  assert.deepEqual(errors, ['processing failed']);
});

test('cancel aborts the queue and prevents completion or later events', async () => {
  const emitter = new EventEmitter();
  const events = [];
  emitter.on('progress', () => events.push('progress'));
  emitter.on('done', () => events.push('done'));
  const control = addon.events(relay(emitter), 10000, 4);
  control.cancel(); // Node has not returned to its event loop yet.
  await closed(control);
  assert.deepEqual(events, []);
  assert(control.rejects());
  assert.throws(() => control.cancel(), /napi_closing/);
});

test('EventEmitter tasks reject invalid arguments before starting producers', () => {
  const emit = relay(new EventEmitter());
  assert.throws(() => addon.events(emit), /requires/);
  assert.throws(() => addon.events(emit, -1, 1), /invalid/);
  assert.throws(() => addon.events(emit, 1, 0), /invalid/);
  assert.throws(() => addon.events(emit, 1, 33), /invalid/);
  assert.throws(() => addon.events(23, 1, 1), /function/);
  assert.throws(() => addon.eventFailure(), /requires/);
});

test('Worker EventEmitter listeners run in the owning Worker environment', async () => {
  const worker = new Worker(`
    const { parentPort, threadId } = require('node:worker_threads');
    const { EventEmitter } = require('node:events');
    const addon = require(process.env.NAPI_GO_ADDON);
    const emitter = new EventEmitter();
    const values = [];
    emitter.on('progress', function (index, data) {
      if (this !== emitter) throw Error('bad receiver');
      values.push(data.value);
    });
    emitter.once('done', result => parentPort.postMessage({ threadId, values, result }));
    addon.events(({ name, args }) => emitter.emit(name, ...args), 8, 1);
  `, { eval: true });
  const message = new Promise((resolve, reject) => {
    worker.once('message', resolve);
    worker.once('error', reject);
  });
  const exited = new Promise((resolve, reject) => {
    worker.once('exit', code => code === 0 ? resolve() : reject(Error('worker exit ' + code)));
    worker.once('error', reject);
  });
  const [result] = await Promise.all([message, exited]);
  assert(result.threadId > 0);
  assert.deepEqual(result.values, [0, 1, 4, 9, 16, 25, 36, 49]);
  assert.deepEqual(result.result, { count: 8, sum: 140 });
});
