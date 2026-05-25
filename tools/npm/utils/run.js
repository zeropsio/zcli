#!/usr/bin/env node

const { existsSync } = require("fs");
const { run, install, getBinaryPath } = require("./binary");

if (!existsSync(getBinaryPath())) {
  process.stderr.write("binary not found downloading now...\n");
  install()
    .then(() => run())
    .catch((err) => {
      process.stderr.write(`download failed: ${err.message}\n`);
      process.exit(1);
    });
} else {
  run();
}