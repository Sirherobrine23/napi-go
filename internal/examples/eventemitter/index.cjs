'use strict';
const { EventEmitter } = require('node:events');
const path = require('node:path');
const addon = require(process.env.NAPI_GO_ADDON || path.resolve(__dirname, '../../../build/eventemitter.node'));

class BackgroundTask extends EventEmitter {
  start(count = 10) {
    addon.start(({ name, data }) => {
      if (name === 'error') this.emit(name, new Error(data));
      else this.emit(name, data);
    }, count);
    return this;
  }
}

module.exports = BackgroundTask;

if (require.main === module) {
  const task = new BackgroundTask();
  task.on('progress', data => console.log('progress:', data));
  task.once('done', data => console.log('done:', data));
  task.on('error', error => { console.error(error); process.exitCode = 1; });
  task.start();
}
