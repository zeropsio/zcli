// Single source of truth for the platforms we publish a prebuilt binary for.
// Keyed by `${process.platform} ${process.arch}` so the launcher can look up the
// current host directly, and reused by the release build script to emit one npm
// package per platform.
//
// Fields:
//   pkg    - unscoped package name; full name is `${SCOPE}/${pkg}`
//   binary - filename of the executable inside the platform package's bin/ dir
//   os/cpu - values for the platform package's package.json os/cpu fields, so npm
//            (and Bun, which honors them) installs only the matching package
//   asset  - name of the raw binary asset on the GitHub release to package

const SCOPE = "@zerops";

const PLATFORMS = [
  { key: "linux x64", pkg: "zcli-linux-amd64", binary: "zcli-linux-amd64", os: "linux", cpu: "x64", asset: "zcli-linux-amd64" },
  { key: "linux ia32", pkg: "zcli-linux-i386", binary: "zcli-linux-i386", os: "linux", cpu: "ia32", asset: "zcli-linux-i386" },
  { key: "darwin x64", pkg: "zcli-darwin-amd64", binary: "zcli-darwin-amd64", os: "darwin", cpu: "x64", asset: "zcli-darwin-amd64" },
  { key: "darwin arm64", pkg: "zcli-darwin-arm64", binary: "zcli-darwin-arm64", os: "darwin", cpu: "arm64", asset: "zcli-darwin-arm64" },
  { key: "win32 x64", pkg: "zcli-win-x64", binary: "zcli-win-x64.exe", os: "win32", cpu: "x64", asset: "zcli-win-x64.exe" },
];

module.exports = { SCOPE, PLATFORMS };
