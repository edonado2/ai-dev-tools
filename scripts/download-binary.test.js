// scripts/download-binary.test.js
const { test } = require('node:test');
const assert = require('node:assert');
const fs = require('fs');
const path = require('path');

test('download-binary.js exists and exports downloadBinary', () => {
  const mod = require('./download-binary.js');
  assert.ok(typeof mod.downloadBinary === 'function');
});

test('package.json has correct bin entry', () => {
  const pkg = JSON.parse(fs.readFileSync(path.join(__dirname, '../package.json'), 'utf8'));
  assert.ok(pkg.bin && pkg.bin['ai-dev'] === './bin/ai-dev');
});

test('bin/ai-dev wrapper exists', () => {
  const wrapperPath = path.join(__dirname, '../bin/ai-dev');
  assert.ok(fs.existsSync(wrapperPath));
  const content = fs.readFileSync(wrapperPath, 'utf8');
  assert.ok(content.includes('execFileSync'));
  assert.ok(content.includes('../lib/ai-dev'));
});