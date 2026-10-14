'use strict';
const path = require('node:path');
const addon = require(process.env.NAPI_GO_ADDON || path.resolve(__dirname, '../../../build/events_thread.node'));

const emitter = addon.start(10);
emitter.on('progress', value => console.log('progress:', value));
emitter.once('done', result => console.log('done:', result));
emitter.on('error', error => { console.error(error); process.exitCode = 1; });
// emitter.cancel() stops the Go processor and discards pending events.
