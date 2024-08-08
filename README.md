# Zerops zCLI

Zerops zCLI is a command line utility for working with [zerops.io](https://zerops.io). It's used 
for **CI/CD** development and CLI lovers.

## Supported platforms

* Windows
* Linux
* MacOS (arm64, amd64)

## Requirements

* [wireguard](https://www.wireguard.com)

## Install zCLI

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

### Package managers

#### Npm
```
npm i -g @zerops/zcli
```

Currently, the zCLI is distributed for Linux (x86 & x64 architecture), macOS (x64 & M1 architecture) and Windows (x64 architecture).

To download the zCLI directly, use the [latest release](https://github.com/zeropsio/zcli/releases/latest/) on GitHub.

## Additional documentation

https://docs.zerops.io/references/cli
