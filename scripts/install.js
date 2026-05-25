#!/usr/bin/env node
"use strict";

const crypto = require("node:crypto");
const fs = require("node:fs");
const https = require("node:https");
const path = require("node:path");

const pkg = require("../package.json");

const supportedAssets = {
  "darwin-arm64": "ft-mcp-darwin-arm64",
  "darwin-x64": "ft-mcp-darwin-x64",
  "linux-arm64": "ft-mcp-linux-arm64",
  "linux-x64": "ft-mcp-linux-x64",
  "win32-arm64": "ft-mcp-win32-arm64.exe",
  "win32-x64": "ft-mcp-win32-x64.exe",
};

const platformKey = `${process.platform}-${process.arch}`;
const assetName = supportedAssets[platformKey];

if (process.env.FT_MCP_SKIP_DOWNLOAD === "1") {
  console.error("ft-mcp: skipped binary download because FT_MCP_SKIP_DOWNLOAD=1");
  process.exit(0);
}

if (!assetName) {
  const supported = Object.keys(supportedAssets).sort().join(", ");
  console.error(`ft-mcp: unsupported platform ${platformKey}. Supported platforms: ${supported}`);
  process.exit(1);
}

const repo = process.env.FT_MCP_GITHUB_REPO || repoFromPackageJson(pkg);
if (!repo) {
  console.error("ft-mcp: unable to determine GitHub repository for binary download");
  process.exit(1);
}

const releaseBase = `https://github.com/${repo}/releases/download/v${pkg.version}`;
const binaryURL = `${releaseBase}/${assetName}`;
const checksumURL = `${releaseBase}/checksums.txt`;
const binDir = path.join(__dirname, "..", "bin");
const targetName = process.platform === "win32" ? "ft-mcp-bin.exe" : "ft-mcp-bin";
const targetPath = path.join(binDir, targetName);

main().catch((err) => {
  console.error(`ft-mcp: ${err.message}`);
  process.exit(1);
});

async function main() {
  fs.mkdirSync(binDir, { recursive: true });

  const [binary, checksum] = await Promise.all([
    download(binaryURL),
    download(checksumURL),
  ]);

  verifyChecksum(binary, checksum.toString("utf8"));

  fs.writeFileSync(targetPath, binary);
  if (process.platform !== "win32") {
    fs.chmodSync(targetPath, 0o755);
  }

  console.error(`ft-mcp: installed ${assetName}`);
}

function repoFromPackageJson(packageJson) {
  const repository = packageJson.repository;
  const raw = typeof repository === "string" ? repository : repository && repository.url;
  if (!raw) {
    return "";
  }

  const match = raw.match(/github\.com[:/]([^/\s]+\/[^/\s#]+?)(?:\.git)?(?:[#?].*)?$/i);
  return match ? match[1].replace(/\.git$/i, "") : "";
}

function verifyChecksum(binary, checksumText) {
  const expected = findChecksum(checksumText, assetName);
  if (!/^[a-f0-9]{64}$/.test(expected)) {
    throw new Error(`invalid checksum for ${assetName}`);
  }

  const actual = crypto.createHash("sha256").update(binary).digest("hex");
  if (actual !== expected) {
    throw new Error(`checksum mismatch for ${assetName}`);
  }
}

function findChecksum(checksumText, filename) {
  for (const line of checksumText.split(/\r?\n/)) {
    const parts = line.trim().split(/\s+/);
    if (parts.length >= 2 && parts[parts.length - 1] === filename) {
      return parts[0].toLowerCase();
    }
  }

  throw new Error(`checksum not found for ${filename}`);
}

function download(url, redirects = 0) {
  return new Promise((resolve, reject) => {
    const req = https.get(
      url,
      {
        headers: {
          "User-Agent": `ft-mcp-npm/${pkg.version}`,
        },
        timeout: 30000,
      },
      (res) => {
        if ([301, 302, 303, 307, 308].includes(res.statusCode)) {
          res.resume();
          if (!res.headers.location) {
            reject(new Error(`redirect without location from ${url}`));
            return;
          }
          if (redirects >= 5) {
            reject(new Error(`too many redirects while downloading ${url}`));
            return;
          }
          resolve(download(new URL(res.headers.location, url).toString(), redirects + 1));
          return;
        }

        if (res.statusCode !== 200) {
          res.resume();
          reject(new Error(`download failed (${res.statusCode}) for ${url}`));
          return;
        }

        const chunks = [];
        res.on("data", (chunk) => chunks.push(chunk));
        res.on("end", () => resolve(Buffer.concat(chunks)));
      }
    );

    req.on("timeout", () => req.destroy(new Error(`download timed out for ${url}`)));
    req.on("error", reject);
  });
}
