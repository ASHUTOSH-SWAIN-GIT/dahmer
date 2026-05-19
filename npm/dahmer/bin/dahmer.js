#!/usr/bin/env node
"use strict";

const { spawnSync } = require("child_process");
const path = require("path");
const fs = require("fs");

function resolveBinary() {
  const { platform, arch } = process;
  const pkgName = `@dahmercli/${platform}-${arch}`;
  const binName = platform === "win32" ? "dahmer.exe" : "dahmer";

  try {
    // Per-platform packages export `bin/<binName>` as a file.
    return require.resolve(`${pkgName}/bin/${binName}`);
  } catch (_) {
    return null;
  }
}

const binary = resolveBinary();
if (!binary || !fs.existsSync(binary)) {
  console.error(
    `dahmer: no prebuilt binary for ${process.platform}-${process.arch}.\n` +
      `Supported: darwin-arm64, darwin-x64, linux-arm64, linux-x64, win32-arm64, win32-x64.`
  );
  process.exit(1);
}

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
});

if (result.error) {
  console.error("dahmer:", result.error.message);
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);
