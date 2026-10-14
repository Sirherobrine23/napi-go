'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { EventEmitter } = require('node:events');
const { Worker } = require('node:worker_threads');
const { spawnSync } = require('node:child_process');
const addon = require(process.env.NAPI_GO_ADDON);

async function finalized(control) {
  const deadline = Date.now() + 5000;
  while (!control.closed()) {
    assert(Date.now() < deadline, 'EventEmitterThread did not finalize');
    await new Promise(resolve => setTimeout(resolve, 5));
  }
}

test('native EventEmitter is constructed without a global require', () => {
  assert(addon.nativeEmitter() instanceof EventEmitter);
});

test('native emitter methods pass the event name, args and correct this', () => {
  const emitter = addon.nativeEmitter();
  const received = [];
  const callback = function (...args) { assert.equal(this, emitter); received.push(args); };
  addon.nativeListen(emitter, 'on', 'native', callback);
  assert.equal(addon.nativeEmit(emitter, 'native', 23, 'payload'), true);
  assert.deepEqual(received, [[23, 'payload']]);
  assert.deepEqual(addon.nativeInfo(emitter), { names: ['native'], count: 1, listeners: 1, max: 23 });
  addon.nativeListen(emitter, 'remove', 'native', callback);
  assert.equal(addon.nativeEmit(emitter, 'native', 42), false);
});

test('native once/prepend methods preserve listener order and removal', () => {
  const emitter = addon.nativeEmitter();
  const order = [];
  addon.nativeListen(emitter, 'on', 'native', () => order.push('on'));
  addon.nativeListen(emitter, 'once', 'native', () => order.push('once'));
  addon.nativeListen(emitter, 'prepend', 'native', () => order.push('prepend'));
  addon.nativeListen(emitter, 'prependOnce', 'native', () => order.push('prependOnce'));
  addon.nativeEmit(emitter, 'native');
  addon.nativeEmit(emitter, 'native');
  assert.deepEqual(order, ['prependOnce', 'prepend', 'on', 'once', 'prepend', 'on']);
  addon.nativeRemoveAll(emitter);
  assert.deepEqual(emitter.eventNames(), []);
});

test('Run processes in Go and delivers progress then completion through its emitter', async () => {
  const control = addon.threadRun('normal');
  const events = [];
  const emitter = control.value;
  emitter.on('progress', function (index, data) {
    assert.equal(this, emitter);
    events.push([index, data.square]);
  });
  emitter.once('done', (...args) => events.push(['done', args.length]));
  await finalized(control);
  assert.deepEqual(events, [...Array.from({ length: 32 }, (_, i) => [i, i*i]), ['done', 0]]);
  assert.throws(() => control.emitter(), /napi_closing/);
  assert.throws(() => control.emit(), /napi_closing/);
});

test('one EventEmitterThread accepts concurrent Go producers', async () => {
  const control = addon.threadRun('concurrent');
  const values = [];
  let countAtCompletion;
  control.value.on('progress', value => values.push(value));
  control.value.once('done', () => { countAtCompletion = values.length; });
  await finalized(control);
  assert.equal(countAtCompletion, 96);
  assert.deepEqual(values.sort((a,b) => a-b), Array.from({length:96}, (_,i)=>i));
});

test('Thread wraps an existing emitter and Close drains accepted events', async () => {
  const emitter = new EventEmitter();
  const control = addon.threadManual(emitter);
  assert.equal(control.emitter(), emitter);
  const data = [];
  emitter.on('manual', value => data.push(value));
  control.emit();
  control.close();
  control.close();
  await finalized(control);
  assert.deepEqual(data, ['data']);
  assert.equal(emitter.emit('still-alive'), false);
});

test('nonblocking Emit reports queue full and rejects JS handles', async () => {
  const control = addon.threadQueue({});
  const data = [];
  control.value.on('queued', value => data.push(value));
  await finalized(control);
  assert.deepEqual(data, [1]);
});

test('Abort cancels an active processor and discards pending events', async () => {
  const control = addon.threadRun('active');
  const received = [];
  control.value.on('progress', value => received.push(value));
  control.value.on('done', () => received.push('done'));
  control.abort();
  control.abort();
  await finalized(control);
  assert.deepEqual(received, []);
});

test('processor errors become JavaScript Error objects', async () => {
  const control = addon.threadRun('error');
  let error;
  control.value.on('error', value => { error = value; });
  await finalized(control);
  assert(error instanceof Error);
  assert.equal(error.message, 'processor failed');
});

test('manual Go error payloads become JavaScript Error objects', async () => {
  const control = addon.threadManualError();
  let error;
  control.value.on('error', value => { error = value; });
  await finalized(control);
  assert(error instanceof Error);
  assert.equal(error.message, 'manual error');
});

test('processor panics are recovered and emitted as errors', async () => {
  const control = addon.threadRun('panic');
  let error;
  control.value.on('error', value => { error = value; });
  await finalized(control);
  assert(error instanceof Error);
  assert.match(error.message, /processor panic/);
});

test('shutdown cancels Run context without emitting cancellation as an error', async () => {
  const control = addon.threadRun('wait');
  const errors = [];
  control.value.on('error', value => errors.push(value));
  control.close();
  await finalized(control);
  assert.deepEqual(errors, []);
});

test('persistent emitter reference survives callback scopes and forced GC', async () => {
  const control = addon.threadAfterScope();
  for (let i=0;i<4;i++) { global.gc(); await new Promise(resolve=>setImmediate(resolve)); }
  const emitter = control.emitter();
  let received;
  emitter.once('done', value => { received=value; });
  control.start();
  await finalized(control);
  assert.equal(received, 42);
});

test('thread validation rejects invalid processors and repeated Run', () => addon.threadValidation());

test('Unref lets an idle EventEmitterThread release its references at process exit', () => {
  const child = spawnSync(process.execPath, ['-e', 'const a=require(process.env.NAPI_GO_ADDON); a.threadManual().unref();'], { env:process.env, timeout:5000 });
  assert.equal(child.status, 0, child.stderr?.toString());
});

test('Worker termination closes EventEmitterThread with an active Go processor', async () => {
  const worker = new Worker(`
    const {parentPort}=require('node:worker_threads');
    const a=require(process.env.NAPI_GO_ADDON);
    const c=a.threadRun('active');
    c.value.on('progress',()=>{});
    c.unref();
    parentPort.postMessage('ready');
    setInterval(()=>{},1000);
  `, {eval:true});
  await new Promise((resolve,reject)=>{worker.once('message',resolve);worker.once('error',reject);});
  assert.equal(await worker.terminate(), 1);
});

test('builtin loader supports legacy hosts with an explicit global require', () => {
  const getter = process.getBuiltinModule;
  const previous = global.require;
  try {
    process.getBuiltinModule = undefined;
    global.require = require;
    assert(addon.nativeEmitter() instanceof EventEmitter);
  } finally {
    process.getBuiltinModule = getter;
    if (previous === undefined) delete global.require;
    else global.require = previous;
  }
});

test('missing synchronous builtin loader reports an error without leaking a queue', () => {
  const getter = process.getBuiltinModule;
  const previous = global.require;
  try {
    process.getBuiltinModule = undefined;
    delete global.require;
    assert.throws(() => addon.threadManual(), /getBuiltinModule or a global require/);
  } finally {
    process.getBuiltinModule = getter;
    if (previous === undefined) delete global.require;
    else global.require = previous;
  }
});

test('EventEmitterThread creates its emitter from ESM without a global require', () => {
  const child = spawnSync(process.execPath, ['--input-type=module', '-e', `
    import { createRequire } from 'node:module';
    import assert from 'node:assert/strict';
    import { EventEmitter } from 'node:events';
    const addon = createRequire(import.meta.url)(process.env.NAPI_GO_ADDON);
    const control = addon.threadRun('normal');
    assert(control.value instanceof EventEmitter);
    let count = 0;
    control.value.on('progress', () => count++);
    control.value.once('done', () => assert.equal(count, 32));
  `], { env: process.env, timeout: 5000 });
  assert.equal(child.status, 0, child.stderr?.toString());
});
