#!/usr/bin/env node
"use strict";

const { spawn } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const binaryName = process.platform === "win32" ? "ft-mcp-bin.exe" : "ft-mcp-bin";
const binaryPath = process.env.FT_MCP_BINARY || path.join(__dirname, binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(
    [
      "ft-mcp: binary not found.",
      `Expected: ${binaryPath}`,
      "Run `npm rebuild ft-mcp` or reinstall the package.",
      "For source checkouts, run the Go binary directly with `go run .` or set FT_MCP_BINARY.",
    ].join("\n")
  );
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), {
  env: process.env,
  stdio: "inherit",
});

for (const signal of ["SIGINT", "SIGTERM"]) {
  process.on(signal, () => {
    if (!child.killed) {
      child.kill(signal);
    }
  });
}

child.on("error", (err) => {
  console.error(`ft-mcp: failed to start binary: ${err.message}`);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    const signalExitCodes = {
      SIGHUP: 129,
      SIGINT: 130,
      SIGTERM: 143,
    };
    process.exit(signalExitCodes[signal] || 1);
  }
  process.exit(code ?? 1);
});
