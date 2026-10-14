'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const addon = require(process.env.NAPI_GO_ADDON);

test('ObjectOf and Assign convert Go maps, nested values and nil handles', () => {
  const object = addon.objectCreate();
  assert.equal(object.name, 'Gopher');
  assert.equal(object.count, 2);
  assert.equal(object.ratio, 1.5);
  assert.deepEqual(object.nested, { active: true });
  assert.deepEqual(object.items, [1, 2]);
  assert.equal(object.unset, undefined);
  assert.equal(object.typedNil, undefined);
  assert.equal(object.add(2, 3), 5);
});

test('SetMethod supports an unnamed Callback and preserves the caller receiver', () => {
  const object = addon.objectCreate();
  assert.equal(object.increment(3), 5);
  assert.equal(object.count, 5);
  const other = { count: 10, increment: object.increment };
  assert.equal(other.increment(2), 12);
  assert.equal(object.count, 5);
  assert.throws(() => object.increment(), /argument 0 is missing/);
  assert.throws(() => object.increment('bad'), /argument 0/);
});

test('ordinary Go functions validate fixed and variadic arguments', () => {
  const object = addon.objectCreate();
  assert.equal(object.join('prefix:'), 'prefix:');
  assert.equal(object.join('prefix:', 'a', 'b'), 'prefix:a,b');
  assert.throws(() => object.join(), /minimum 1 arguments/);
  assert.throws(() => object.add(1), /require 2 arguments/);
  assert.throws(() => object.add(1, 2, 3), /require 2 arguments/);
  assert.throws(() => object.add('one', 2), /cannot set string/);
  assert.equal(object.sum([2, 3, 5]), 10);
  assert.throws(() => object.sum('bad'), /cannot set string/);
  assert.equal(object.noop(), undefined);
  assert.throws(() => object.noop(1), /require 0 arguments/);
  assert.deepEqual(object.anyArgs(), []);
  assert.deepEqual(object.anyArgs(1, { values: [true, 'x'] }), [1, { values: [true, 'x'] }]);
});

test('Go function results propagate nil, interface and concrete errors', () => {
  const object = addon.objectCreate();
  assert.equal(object.nilError(), 7);
  assert.throws(() => object.fail(), /Go method failed/);
  assert.throws(() => object.concreteError(), /concrete Go error/);
});

test('ArgAs and ObjectOf round-trip tagged structs, slices, maps and pointer fields', () => {
  const input = { name: 'config', enabled: true, items: [1, 3], labels: { region: 'br' }, child: { score: 7, Note: 'nested' }, Hidden: 'discard' };
  const expected = { name: 'config', enabled: true, items: [1, 3], labels: { region: 'br' }, child: { score: 7, Note: 'nested' } };
  assert.deepEqual(addon.objectDecode(input), expected);
  assert.deepEqual(addon.objectCreate().echo(input), expected);
  assert.deepEqual(addon.objectDecode({}), { name: '', enabled: false, items: [], labels: {}, child: undefined });
  assert.equal(addon.objectDecode({ child: null }).child, undefined);
  assert.throws(() => addon.objectDecode({ items: ['bad'] }), /argument 0/);
});

test('Decode any handles nested maps and arrays and detects JavaScript cycles', () => {
  const input = { nested: { a: [1, 'two', true] }, list: [{ value: 4 }] };
  assert.deepEqual(addon.objectAny(input), input);
  assert.deepEqual(addon.objectAny([input]), [input]);
  const cyclic = {};
  cyclic.self = cyclic;
  assert.throws(() => addon.objectAny(cyclic), /conversion depth.*cyclic/);
  // Raw Promise handles must not be treated as Go-owned napi_deferred values.
  const promise = Promise.resolve(3);
  assert.equal(addon.objectAny(promise), promise);
});

test('GetInto and GetAs support named scalar types and checked native wrappers', () => {
  assert.equal(addon.objectGetString({ title: 'ok' }, 'title'), 'ok');
  assert.equal(addon.objectGetBool({ flag: true }), true);
  assert.equal(addon.objectGetWrapper({ name: 'wrapped' }), 'wrapped');
  assert.throws(() => addon.objectGetWrapper({ name: 3 }), /property "name".*cannot decode number/);
  assert.throws(() => addon.objectGetString({}, 'missing'), /property "missing"/);
  assert.throws(() => addon.objectGetString({ title: 3 }, 'title'), /property "title"/);
  assert.throws(() => addon.objectGetBool({ flag: 'true' }), /property "flag"/);
  assert.throws(() => addon.objectGetWrapper(3), /argument 0/);
});

test('property getters throw their original JavaScript exception', () => {
  const failure = new Error('getter failure');
  assert.throws(() => addon.objectGetString({ get name() { throw failure; } }, 'name'), error => error === failure);
  assert.throws(() => addon.objectDecode({ get name() { throw failure; } }), error => error === failure);
  assert.equal(addon.objectGetString({ name: 'after exception' }, 'name'), 'after exception');
});

test('CallMethod converts Go arguments, resolves inherited methods and preserves this', () => {
  const prototype = { method(message, details) { assert.equal(this, object); return [this.name, message, details]; } };
  const object = Object.create(prototype);
  object.name = 'receiver';
  assert.deepEqual(addon.objectCall(object), ['receiver', 'from Go', { count: 3 }]);
  object.method = function () { assert.equal(this, object); return 42; };
  assert.equal(addon.objectCallZero(object), 42);
  object.method = 23;
  assert.throws(() => addon.objectCall(object), /property "method" is number, expected function/);
  assert.throws(() => addon.objectCall({}), /expected function/);
});

test('CallMethod propagates getter and method exceptions', () => {
  const failure = new TypeError('method failure');
  assert.throws(() => addon.objectCall({ method() { throw failure; } }), error => error === failure);
  assert.throws(() => addon.objectCall({ get method() { throw failure; } }), error => error === failure);
});

test('CallValues and CallWithThis convert Go values without losing the receiver', () => {
  assert.equal(addon.objectFunctionCall(function (value, items) {
    assert.equal(this, globalThis);
    assert.deepEqual(items, [5, 6]);
    return value + items[0] + items[1];
  }), 15);
  assert.throws(() => addon.objectFunctionCall(() => 'not a number'), /cannot set string/);
  const object = { count: 20 };
  assert.equal(addon.objectFunctionThis(function (amount) { assert.equal(this, object); return this.count + amount; }, object), 23);
});

test('NewValues accepts ordinary JavaScript constructors and converts Go arguments', () => {
  class Person { constructor(name, age) { this.name = name; this.age = age; } }
  const person = addon.objectClassNew(Person);
  assert(person instanceof Person);
  assert.deepEqual({ ...person }, { name: 'Gopher', age: 23 });
  assert.throws(() => addon.objectClassNew(() => {}));
  const failure = new Error('constructor failure');
  assert.throws(() => addon.objectClassNew(class { constructor() { throw failure; } }), error => error === failure);
});

test('UnwrapAs checks Go state types and rejects unwrapped objects', () => {
  const counter = new addon.Counter(5);
  assert.equal(addon.objectUnwrap(counter), 5);
  assert.throws(() => addon.objectUnwrapMismatch(counter), /wrapped state is .*counter, expected .*objectChild/);
  assert.throws(() => addon.objectUnwrap({}));
});

test('Assign uses deterministic setters and reports errors after partial assignment', () => {
  const calls = [];
  const object = { set a(value) { calls.push(['a', value]); }, set b(value) { calls.push(['b', value]); } };
  assert.equal(addon.objectAssign(object), object);
  assert.deepEqual(calls, [['a', 1], ['b', 2]]);
  const frozen = Object.freeze({});
  // Preserve Object.Set / napi_set_property behavior for non-writable targets.
  assert.equal(addon.objectAssign(frozen), frozen);
  assert.deepEqual(frozen, {});
  const failure = new Error('setter failure');
  const partial = { a: 0, set b(_) { throw failure; } };
  assert.throws(() => addon.objectAssign(partial), error => error === failure);
  assert.equal(partial.a, 1);
  const conversion = {};
  assert.throws(() => addon.objectAssignInvalid(conversion), /property "b".*conversion depth/);
  assert.deepEqual(conversion, { a: 1 });
});

test('invalid inputs return Go errors before unsafe native calls', () => {
  for (const [kind, pattern] of [
    ['primitive', /ObjectOf requires an object/],
    ['nil-function', /non-nil function/],
    ['not-function', /requires a Go function/],
    ['nil-target', /non-nil pointer/],
    ['value-target', /non-nil pointer/],
    ['nil-value', /JavaScript value is nil/],
    ['negative-argument', /argument -1 is missing/],
    ['cyclic', /conversion depth.*cyclic/],
    ['foreign', /another environment/],
  ]) assert.throws(() => addon.objectInvalid(kind), pattern, kind);
});
