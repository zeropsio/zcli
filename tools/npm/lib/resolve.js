const { SCOPE, PLATFORMS } = require("./platforms");

function currentPlatform() {
  const key = `${process.platform} ${process.arch}`;
  const entry = PLATFORMS.find(p => p.key === key);
  if (!entry) {
    const supported = PLATFORMS.map(p => p.key).join(", ");
    throw new Error(`zcli: unsupported platform "${key}". Supported: ${supported}`);
  }
  return entry;
}

function getBinaryPath() {
  const { pkg, binary } = currentPlatform();
  const fullName = `${SCOPE}/${pkg}`;
  try {
    return require.resolve(`${fullName}/bin/${binary}`);
  } catch {
    throw new Error(
      [
        `zcli: could not find the platform binary package "${fullName}".`,
        "This usually means optional dependencies were skipped during install",
        "(e.g. --no-optional, --ignore-optional, or a package manager configured to skip them).",
        "Reinstall @zerops/zcli with optional dependencies enabled.",
      ].join("\n")
    );
  }
}

module.exports = { currentPlatform, getBinaryPath };
