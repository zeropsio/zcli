# Zerops zCLI

Zerops zCLI is a command line utility for working with [zerops.io](https://zerops.io). It's used 
for **CI/CD** development and CLI lovers.

## Supported platforms

* Windows
* Linux
* MacOS (arm64, amd64)
* NixOS

## Requirements

* [wireguard](https://www.wireguard.com)

## Install zCLI

### Package managers

#### Npm
```
npm i -g @zerops/zcli
```

### Windows
Execute following command in PowerShell
```powershell
irm https://zerops.io/zcli/install.ps1 | iex
```

### Linux/MacOS
Execute following command in Terminal
```shell
curl -L https://zerops.io/zcli/install.sh | sh
```

### NixOS

- Clone this repository
- `cd zcli` into the root of the cloned repository and run `nix develop`.
- Run `nix build` to build the binary / execuetable of zCli.
- zCLI's binary / execuetable will be present in `./result/bin/zcli`.



Currently, the zCLI is distributed for Linux (x86 & x64 architecture), macOS (x64 & M1 architecture) and Windows (x64 architecture).

To download the zCLI directly, use the [latest release](https://github.com/zeropsio/zcli/releases/latest/) on GitHub.

## Additional documentation

https://docs.zerops.io/references/cli
