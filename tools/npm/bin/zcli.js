#!/usr/bin/env node

const { spawnSync } = require("child_process");
const { getBinaryPath } = require("../lib/resolve");

let binaryPath;
try {
  binaryPath = getBinaryPath();
} catch (e) {
  process.stderr.write(e.message + "\n");
  process.exit(1);
}

const result = spawnSync(binaryPath, process.argv.slice(2), {
  cwd: process.cwd(),
  stdio: "inherit",
});

if (result.error) {
  process.stderr.write(`zcli: failed to start binary: ${result.error.message}\n`);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
