const { Binary } = require("binary-install");
const os = require("os");

function getPlatform() {
  const type = os.type();
  const arch = os.arch();
  if (type === "Linux" && arch === "x64") return "linux-amd64";
  if (type === "Linux" && arch === "x86") return "linux-i386";
  if (type === "Darwin" && arch === "x64") return "darwin-amd64";
  if (type === "Windows_NT" && arch === "x64") return "win-x64";
  throw new Error(`Unsupported platform: ${type} ${arch}`);
}
function getBinary() {
  const platform_arch = getPlatform();
  const version = require("../package.json").version;
  const compatibleVersion = version.startsWith('v') ? version : `v${version}`;
  const url = `https://github.com/zeropsio/zcli/releases/download/${compatibleVersion}/zcli-${platform_arch}-npm.tar.gz`;
  const name = `zcli-${platform_arch}`;

  console.log('Downloading '+ name +' from', url);

  return new Binary(name, url);
}
const run = () => {
  const binary = getBinary();
  console.log('Running binary');
  binary.run();
};
const install = () => {
  const binary = getBinary();
  console.log('Installing binary');
  binary.install();
};
const uninstall = () => {
  const binary = getBinary();
  console.log('Uninstalling binary');
  binary.uninstall();
};
module.exports = {
  install,
  run,
  uninstall,
};
