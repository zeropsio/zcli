const { Binary } = require("binary-install");
const os = require("os");

function getPlatform() {
  const type = os.type();
  const arch = os.arch();

  if (type === "Linux" && arch === "x64") return "linux-amd64";
  if (type === "Linux" && arch === "x86") return "linux-i386";
  if (type === "Darwin" && arch === "x64") return "darwin-amd64";

  throw new Error(`Unsupported platform: ${type} ${arch}`);
}

function getBinary() {
  const platform_arch = getPlatform();
  const version = require("../../package.json").version;
  const url = `https://github.com/zeropsio/zcli/releases/download/v${version}/zcli-${platform_arch}.tar.gz`;
  const name = `zcli-${platform_arch}`;
  return new Binary(url, { name });
}
const run = () => {
  const binary = getBinary();
  binary.run();
};

const install = () => {
  const binary = getBinary();
  binary.install();
};

const uninstall = () => {
  const binary = getBinary();
  binary.uninstall();
};

module.exports = {
  install,
  run,
  uninstall,
};
