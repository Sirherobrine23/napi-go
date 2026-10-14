'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const { Worker } = require('node:worker_threads');
const { once } = require('node:events');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const crypto = require('node:crypto');
const zlib = require('node:zlib');
const { PassThrough, Readable } = require('node:stream');
const addon = require(process.env.NAPI_GO_ADDON);

function temporary(t) {
  const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'napi-builtins-'));
  t.after(() => fs.rmSync(directory, { recursive: true, force: true }));
  return directory;
}

test('generic module loader preserves builtin identity and supports arbitrary builtins', () => {
  assert.equal(addon.builtinLoad('path'), path);
  assert.equal(addon.builtinLoad('node:crypto'), crypto);
  assert.equal(addon.builtinLoad('node:http'), require('node:http'));
  assert.equal(addon.builtinLoad('node:fs/promises'), require('node:fs/promises'));
  assert.equal(addon.builtinCall('path', 'join', 'a', 'b'), path.join('a', 'b'));
  const namespace = { value: 10, method(amount) { assert.equal(this, namespace); return this.value + amount; } };
  assert.equal(addon.builtinWrap(namespace), 33);
  assert.throws(() => addon.builtinLoad('napi-go-does-not-exist'), /unknown Node builtin/);
  assert.throws(() => addon.builtinCall('path', 'sep'), /expected function/);
  assert.throws(() => addon.builtinCall('path', 'join', 23), TypeError);
});

test('path wrappers preserve native, POSIX and Windows semantics', () => {
  for (const [flavor, api, input] of [['native', path, path.join(os.tmpdir(), 'demo.txt')], ['posix', path.posix, '/a/b/../file.txt'], ['win32', path.win32, 'C:\\a\\b\\..\\file.txt']]) {
    const actual = addon.builtinPath(flavor, input);
    const parts = api.parse(input);
    assert.deepEqual(actual, { parsed: parts, formatted: api.format(parts), basename: api.basename(input), stem: api.basename(input, parts.ext), dirname: api.dirname(input), extname: api.extname(input), absolute: api.isAbsolute(input), normalize: api.normalize(input), join: api.join('a', 'b', '..', 'c'), resolve: api.resolve('.'), relative: api.relative('a', 'b'), sep: api.sep, delimiter: api.delimiter });
  }
});

test('os wrappers query the running Node host and decode structured CPU/user data', () => {
  const result = addon.builtinOS();
  for (const property of ['arch', 'platform', 'hostname', 'type', 'release', 'homedir', 'tmpdir', 'endianness', 'totalmem']) assert.equal(result[property], os[property]());
  assert.equal(result.parallel, os.availableParallelism());
  assert(result.freemem > 0 && result.freemem <= result.totalmem);
  assert(result.uptime > 0);
  assert.equal(result.loadavg.length, 3);
  assert.equal(result.cpus.length, os.cpus().length);
  if (result.cpus.length) assert.equal(result.cpus[0].model, os.cpus()[0].model);
  const user = os.userInfo();
  assert.equal(result.user.username, user.username);
  assert.equal(result.user.uid, user.uid);
  assert.equal(result.user.homedir, user.homedir);
});

test('os network interfaces preserve host results or native permission errors', () => {
  let native, failure;
  try { native = os.networkInterfaces(); } catch (error) { failure = error; }
  if (failure) assert.throws(() => addon.builtinNetworkInterfaces(), error => error.code === failure.code && error.syscall === failure.syscall);
  else assert.deepEqual(addon.builtinNetworkInterfaces(), native);
});

test('buffer wrappers preserve encodings, binary values and empty buffers', () => {
  const result = addon.builtinBuffer();
  const a = Buffer.from('héllo');
  assert.equal(result.text, 'héllo');
  assert.equal(result.hex, a.toString('hex'));
  assert.equal(result.length, a.length);
  assert.deepEqual(result.buffer, Buffer.concat([a, Buffer.from([0, 255, 1])]));
  assert.deepEqual(result.bytes, [0, 255, 1]);
  assert.equal(result.isBuffer, true);
  assert.equal(result.isArrayBuffer, false);
  assert.deepEqual(result.alloc, Buffer.alloc(3, 7));
  for (const key of ['empty', 'emptyCopy']) assert.deepEqual(result[key], Buffer.alloc(0));
  assert.equal(result.compare, 0);
  assert.equal(result.type, 'buffer');
});

test('CopyBytes detaches Go data from JavaScript memory before mutation', () => {
  const original = Buffer.from([1, 2, 255]);
  const clone = addon.builtinBufferClone(original);
  assert.deepEqual(original, Buffer.from([99, 2, 255]));
  assert.deepEqual(clone, Buffer.from([1, 2, 255]));
  assert.deepEqual(addon.builtinBufferClone(Buffer.alloc(0)), Buffer.alloc(0));
});

test('crypto wrappers delegate hashing/HMAC and propagate algorithm/finalization errors', () => {
  assert.equal(addon.builtinHash('sha256', 'hello'), crypto.createHash('sha256').update('hello').digest('hex'));
  assert.equal(addon.builtinHash('sha256', ''), crypto.createHash('sha256').update('').digest('hex'));
  const result = addon.builtinCrypto();
  assert.equal(result.hmac, crypto.createHmac('sha256', 'secret').update('data').digest('hex'));
  assert.deepEqual(result.emptyHash, [...crypto.createHash('sha256').digest()]);
  assert.deepEqual(result.emptyRandom, []);
  assert.equal(result.random.length, 16);
  assert.match(result.uuid, /^[\da-f]{8}-[\da-f]{4}-4[\da-f]{3}-[89ab][\da-f]{3}-[\da-f]{12}$/i);
  assert.equal(result.equal, true);
  assert.equal(result.different, false);
  assert.deepEqual(result.hashes, crypto.getHashes());
  assert.throws(() => addon.builtinHash('napi-go-no-such-hash', 'x'));
  assert.throws(() => addon.builtinCryptoUnequal(), RangeError);
  assert.throws(() => addon.builtinCryptoFinalized());
});

test('URL wrappers return native URLs and distinguish absent from empty parameters', () => {
  const input = '../item?q=one#hash';
  const result = addon.builtinURL(input);
  const expected = new URL(input, 'https://example.com/base/');
  expected.searchParams.append('q', 'two words');
  expected.searchParams.sort();
  assert(result.value instanceof URL);
  for (const property of ['href', 'origin', 'protocol', 'hostname', 'port', 'pathname', 'search', 'hash']) assert.equal(result[property], expected[property]);
  assert.equal(result.missing, '');
  assert.equal(result.exists, false);
  assert.equal(result.empty, '');
  assert.equal(result.present, true);
  assert.equal(result.hasQ, true);
  assert.deepEqual(result.all, ['one', 'two words']);
  assert.equal(result.query, expected.searchParams.toString());
  assert.equal(result.pairs, 'a=1&a=2');
  assert.equal(result.ascii, 'xn--mnich-kva.example');
  assert.equal(result.unicode, 'münich.example');
  assert.throws(() => addon.builtinURL('http://['), TypeError);
});

test('file URLs retain spaces, Unicode and native path round-trips', () => {
  const value = path.join(os.tmpdir(), 'two words é.txt');
  const result = addon.builtinFileURL(value);
  assert.equal(result.href, require('node:url').pathToFileURL(value).href);
  assert.equal(result.path, value);
});

test('util wrappers preserve formatting, inspection, deep equality and function adaptation', async () => {
  const result = addon.builtinUtil();
  assert.equal(result.formatted, 'value=23 go');
  assert.match(result.inspected, /answer: 42/);
  assert.equal(result.equal, true);
  assert.equal(result.unequal, false);
  const promised = addon.builtinPromisify((value, callback) => callback(null, value + 1));
  assert.equal(await promised(3), 4);
  const callback = addon.builtinCallbackify(async value => value + 2);
  assert.equal(await new Promise((resolve, reject) => callback(3, (error, result) => error ? reject(error) : resolve(result))), 5);
});

test('zlib wrappers round-trip native compressed data including empty payloads', () => {
  for (const input of [Buffer.from([0, 1, 255, 0, 9]), Buffer.alloc(0)]) {
    for (const [compress, decompress, api] of [['gzip', 'gunzip', zlib.gzipSync], ['deflate', 'inflate', zlib.deflateSync], ['deflateRaw', 'inflateRaw', zlib.deflateRawSync], ['brotli', 'unbrotli', zlib.brotliCompressSync]]) {
      const packed = addon.builtinCompress(compress, input);
      assert.deepEqual(packed, api(input));
      assert.deepEqual(addon.builtinCompress(decompress, packed), input);
    }
    assert.deepEqual(addon.builtinCompress('unzip', addon.builtinCompress('gzip', input)), input);
  }
  assert.deepEqual(addon.builtinCompress('gunzip', addon.builtinCompress('gzip', 'héllo')), Buffer.from('héllo'));
  assert.throws(() => addon.builtinCompress('gunzip', Buffer.from('bad data')));
});

test('fs sync wrappers read/write binary/text data and preserve native Stats', t => {
  const directory = temporary(t);
  const result = addon.builtinFS(directory);
  assert.deepEqual(result.data, [0, 255, 3, 4]);
  assert.equal(result.text, 'héllo');
  assert.equal(result.size, 4);
  assert.equal(result.file, true);
  assert.equal(result.directory, true);
  assert.equal(result.symlink, false);
  assert.equal(result.mode, fs.statSync(path.join(directory, 'text.txt')).mode);
  assert.deepEqual(result.entries.sort(), ['data.bin', 'text.txt']);
  assert.equal(result.exists, false);
  assert.deepEqual(result.raw, Buffer.from('héllo'));
  assert.throws(() => addon.builtinFSRead(path.join(directory, 'missing')), error => error.code === 'ENOENT');
  assert.equal(addon.builtinFSStatMaybe(path.join(directory, 'missing')), undefined);
  assert(addon.builtinFSStatMaybe(path.join(directory, 'text.txt')) instanceof fs.Stats);
});

test('fs/promises wrappers return genuine Promises and asynchronous native results', async t => {
  const directory = temporary(t);
  const filename = path.join(directory, 'async.bin');
  const write = addon.builtinFSAsync('write', filename);
  assert(write instanceof Promise);
  await write;
  assert.deepEqual(await addon.builtinFSAsync('read', filename), Buffer.from([1, 0, 255]));
  const stat = await addon.builtinFSAsync('stat', filename);
  assert(stat instanceof fs.Stats);
  assert(stat.isFile());
  assert.equal(stat.size, 3);
  assert.deepEqual(await addon.builtinFSAsync('readdir', directory), ['async.bin']);
  await addon.builtinFSAsync('mkdir', path.join(directory, 'child'));
  await addon.builtinFSAsync('rm', path.join(directory, 'child'));
  await assert.rejects(addon.builtinFSAsync('read', path.join(directory, 'missing')), error => error.code === 'ENOENT');
});

test('Promise Then/Catch/Finally adopt results and propagate JavaScript/Go failures', async () => {
  assert.equal(await addon.builtinPromise(Promise.resolve('hello'), 'upper'), 'HELLO');
  assert.equal(await addon.builtinPromise(Promise.resolve(23), 'identity'), 23);
  assert.equal(await addon.builtinPromise(Promise.resolve(), 'adopt'), 'adopted');
  assert.equal(await addon.builtinPromise(Promise.reject(new Error('original')), 'catch'), 'caught:original');
  await assert.rejects(addon.builtinPromise(Promise.resolve(), 'fail'), /Go promise handler failure/);
  let finalized = 0;
  assert.equal(await addon.builtinPromise(Promise.resolve('kept'), 'finally', () => finalized++), 'kept');
  assert.equal(finalized, 1);
  const rejection = new Error('kept rejection');
  await assert.rejects(addon.builtinPromise(Promise.reject(rejection), 'identity'), error => error === rejection);
  assert.throws(() => addon.builtinPromise(Promise.resolve(), 'invalid'), /requires a Go function/);
});

test('timers deliver arguments after returning and run Go callbacks on the Node thread', async () => {
  let returned = false;
  await new Promise(resolve => {
    const timer = addon.builtinTimer((text, details) => { assert.equal(returned, true); assert.equal(text, 'timeout'); assert.deepEqual(details, { count: 3 }); resolve(); }, 'timeout');
    assert.equal(timer.hasRef(), true);
    returned = true;
  });
  assert.equal(await addon.builtinGoTimer(), 'Go timer callback');
  assert.equal(await addon.builtinTimerPromise(), 'timer promise');
  assert.equal(await addon.builtinImmediatePromise(), 'immediate promise');
  await new Promise(resolve => addon.builtinTimer(text => { assert.equal(text, 'immediate'); resolve(); }, 'immediate'));
});

test('timer controls retain liveness, clear scheduled callbacks and stop intervals', async () => {
  let fired = false;
  const timer = addon.builtinTimer(() => { fired = true; }, 'timeout');
  assert.equal(addon.builtinTimerControl(timer, 'unref'), false);
  assert.equal(addon.builtinTimerControl(timer, 'ref'), true);
  assert.equal(addon.builtinTimerControl(timer, 'refresh'), true);
  addon.builtinTimerControl(timer, 'clear');
  const immediate = addon.builtinTimer(() => { fired = true; }, 'immediate');
  addon.builtinTimerControl(immediate, 'clearImmediate');
  let count = 0;
  await new Promise(resolve => {
    const interval = addon.builtinTimer(text => { assert.equal(text, 'interval'); if (++count === 2) { addon.builtinTimerControl(interval, 'clearInterval'); resolve(); } }, 'interval');
  });
  await new Promise(resolve => setTimeout(resolve, 15));
  assert.equal(fired, false);
  assert.equal(count, 2);
  assert.throws(() => addon.builtinTimer(23, 'timeout'));
});

test('streams return native objects, preserve write backpressure and support pipe', async () => {
  const source = addon.builtinStream();
  assert(source instanceof PassThrough);
  // With no consumer and a 1-byte highWaterMark, the producer must wait for drain.
  assert.equal(addon.builtinStreamWrite(source), false);
  assert.deepEqual(addon.builtinStreamRead(source), Buffer.from('abc'));
  assert.equal(addon.builtinStreamRead(source), null);
  const destination = new PassThrough();
  const chunks = [];
  destination.on('data', chunk => chunks.push(chunk));
  assert.equal(addon.builtinStreamPipe(source, destination), destination);
  const done = once(destination, 'end');
  addon.builtinStreamEnd(source);
  await done;
  assert.deepEqual(Buffer.concat(chunks), Buffer.from('end'));
});

test('stream listeners and Readable.from work with scoped native callbacks', async () => {
  const stream = addon.builtinStream();
  const values = [];
  const fn = addon.builtinStreamListen(stream, chunk => values.push(chunk.toString()));
  assert.equal(typeof fn, 'function');
  addon.builtinStreamWrite(stream);
  const end = once(stream, 'end');
  addon.builtinStreamEnd(stream);
  await end;
  assert.deepEqual(values, ['abc', 'end']);
  const readable = addon.builtinReadable();
  assert(readable instanceof Readable);
  const items = [];
  for await (const value of readable) items.push(value);
  assert.deepEqual(items, ['one', 'two']);
  const destroyed = addon.builtinStream();
  const closed = once(destroyed, 'close');
  addon.builtinStreamDestroy(destroyed);
  await closed;
});

test('DNS Promise lookup uses the host resolver for localhost', async () => {
  const result = addon.builtinDNS();
  assert(result instanceof Promise);
  assert.deepEqual(await result, await require('node:dns/promises').lookup('localhost', { family: 4 }));
});

test('EventEmitter convenience helpers convert values and keep removable listener identity', () => {
  const calls = [];
  const emitter = addon.builtinEvents(function (...args) { calls.push(args); assert.equal(this.constructor, require('node:events').EventEmitter); });
  assert.deepEqual(calls, [[23, { fromGo: true }], [24, { fromGo: true }]]);
  assert.equal(emitter.listenerCount('value'), 0);
  assert.equal(emitter.listenerCount('once'), 0);
});

test('generic wrappers reject invalid module/promise/buffer/callback inputs', () => {
  for (const [input, pattern] of [['nilModule', /value is nil/], ['nilPromise', /value is nil/], ['nilBuffer', /value is nil/], ['notPromise', /expected Promise/], ['notBuffer', /expected a Node Buffer/], ['callback', /non-nil function/]]) assert.throws(() => addon.builtinInvalid(input), pattern, input);
});

test('builtin wrappers work without global require in ESM', () => {
  const child = spawnSync(process.execPath, ['--input-type=module', '-e', `
    import { createRequire } from 'node:module';
    import assert from 'node:assert/strict';
    const addon = createRequire(import.meta.url)(process.env.NAPI_GO_ADDON);
    assert.equal(globalThis.require, undefined);
    assert.equal(addon.builtinHash('sha256','hello').length,64);
    assert.equal(await addon.builtinTimerPromise(),'timer promise');
  `], { env: process.env, timeout: 5000 });
  assert.equal(child.status, 0, child.stderr?.toString());
});

test('builtin wrappers keep Worker environments independent', async () => {
  const worker = new Worker(`
    const {parentPort}=require('node:worker_threads');
    const addon=require(process.env.NAPI_GO_ADDON);
    const info=addon.builtinOS();
    addon.builtinTimerPromise().then(value=>parentPort.postMessage({platform:info.platform,value,hash:addon.builtinHash('sha256','worker')}));
  `, { eval: true });
  const [message] = await once(worker, 'message');
  assert.equal(message.platform, process.platform);
  assert.equal(message.value, 'timer promise');
  assert.equal(message.hash, crypto.createHash('sha256').update('worker').digest('hex'));
  await once(worker, 'exit');
});

test('legacy builtin loading can use an explicit global require', () => {
  const original = process.getBuiltinModule;
  const previous = global.require;
  try {
    process.getBuiltinModule = undefined;
    global.require = require;
    assert.equal(addon.builtinLoad('path'), path);
    delete global.require;
    assert.throws(() => addon.builtinLoad('path'), /getBuiltinModule or a global require/);
  } finally {
    process.getBuiltinModule = original;
    if (previous === undefined) delete global.require; else global.require = previous;
  }
});
