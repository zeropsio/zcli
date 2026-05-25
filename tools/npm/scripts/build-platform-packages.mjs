#!/usr/bin/env node
// Generates one publishable npm package per platform from the raw release binaries.
//
// Usage:
//   node scripts/build-platform-packages.mjs --binaries <dir> --version <v> --out <dir>
//
//   --binaries  directory containing the raw release binaries (named per PLATFORMS[].asset)
//   --version   version to stamp into every generated package.json
//   --out       directory to write the generated packages into (one subdir per platform)
//
// Each generated package carries os/cpu fields so the package manager (npm, and Bun,
// which honors them) installs only the one matching the host.

import { mkdirSync, copyFileSync, writeFileSync, chmodSync, existsSync, rmSync } from "node:fs";
import { join } from "node:path";
import { createRequire } from "node:module";

const require = createRequire(import.meta.url);
const { SCOPE, PLATFORMS } = require("../lib/platforms.js");

function parseArgs(argv) {
  const args = {};
  for (let i = 0; i < argv.length; i += 2) {
    const key = argv[i];
    if (!key.startsWith("--")) {
      throw new Error(`unexpected argument: ${key}`);
    }
    args[key.slice(2)] = argv[i + 1];
  }
  for (const required of ["binaries", "version", "out"]) {
    if (!args[required]) {
      throw new Error(`missing required --${required}`);
    }
  }
  return args;
}

const { binaries, version, out } = parseArgs(process.argv.slice(2));

rmSync(out, { recursive: true, force: true });

for (const p of PLATFORMS) {
  const src = join(binaries, p.asset);
  if (!existsSync(src)) {
    throw new Error(`binary not found for ${p.pkg}: ${src}`);
  }

  const pkgDir = join(out, p.pkg);
  const binDir = join(pkgDir, "bin");
  mkdirSync(binDir, { recursive: true });

  const dest = join(binDir, p.binary);
  copyFileSync(src, dest);
  chmodSync(dest, 0o755);

  const pkgJson = {
    name: `${SCOPE}/${p.pkg}`,
    version,
    description: `zcli prebuilt binary for ${p.os}/${p.cpu}.`,
    license: "MIT",
    os: [p.os],
    cpu: [p.cpu],
    files: ["bin/"],
  };
  writeFileSync(join(pkgDir, "package.json"), JSON.stringify(pkgJson, null, 2) + "\n");

  console.log(`built ${SCOPE}/${p.pkg}@${version} (${p.os}/${p.cpu}) -> ${dest}`);
}

console.log(`\nGenerated ${PLATFORMS.length} platform packages in ${out}`);
