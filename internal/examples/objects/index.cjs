'use strict';
const assert = require('node:assert/strict');
const addon = require('../../../build/objects.node');

const counter = addon.createCounter(10);
assert.equal(counter.increment(), 11);
assert.equal(counter.increment(4), 15);
assert.equal(counter.count, 15);
assert.throws(() => counter.increment('invalid'), /argument 0/);
assert.equal(addon.describe(addon.defaults), 'Gopher: enabled=true, items=2');
console.log({ count: counter.count, description: addon.describe(addon.defaults) });
