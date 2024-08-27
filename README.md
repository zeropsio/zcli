![Zerops](https://github.com/zeropsio/recipe-shared-assets/blob/main/covers/svg/cover-zcli.svg)

<h2 align="center">
   Zerops zCLI
  <br/>
  <br/>
</h2>

<p align="center">
  <br/>
   A <b>Command Line Utility / Command Line Interface</b> used for interacting with <a href="https://zerops.io/" target="_blank">Zerops</a> platform.
  <br/>
</p>

<p align="center">
<b>Made with</b> ❤️ for <b>CI/CD</b> development and <b>CLI</b> lovers.
<br/>
</p>

<br />

<div align="center">

[![CI](https://github.com/zeropsio/zcli/actions/workflows/main.yml/badge.svg)](https://github.com/zeropsio/zcli/actions/workflows/ci.yml)
[![NPM Downloads](https://img.shields.io/npm/d18m/%40zerops%2Fzcli)](https://www.npmjs.com/package/@zerops/zcli)
[![npm version](https://badge.fury.io/js/@zerops%2Fzcli.svg)](https://badge.fury.io/js/@zerops%2Fzcli)
[![Discord](https://img.shields.io/discord/735781031147208777)](https://discord.gg/xxzmJSDKPT)
[![npm bundle size](https://img.shields.io/bundlephobia/min/%40zerops%2Fzcli)](https://www.npmjs.com/package/@zerops/zcli)

</div>

<br/>

<h3 align="end">
<a href="https://docs.zerops.io/" target="_blank">Read the docs →</a>
<br/>
</h3>

## Supported platforms

- Windows
- Linux
- MacOS (arm64, amd64)
- NixOS

## Requirements

- [Wireguard](https://www.wireguard.com/install/) - utilized by `zcli vpn` command.

## Install zCLI

### Package managers

#### Npm

```sh
npm i -g @zerops/zcli
```

### Windows

Execute following command in PowerShell:

```powershell
irm https://zerops.io/zcli/install.ps1 | iex
```

### Linux/MacOS

Execute following command in Terminal:

```shell
curl -L https://zerops.io/zcli/install.sh | sh
```

### NixOS

- Clone this repository
- `cd zcli` into the root of the cloned repository and run `nix develop`.
- Run `nix build` to build the binary / execuetable of zCli.
- zCLI's binary / execuetable will be present in `./result/bin/zcli`.

Currently, the zCLI is distributed for Linux (x86 & x64 architecture), macOS (x64 & M1 architecture) and Windows (x64 architecture).

> [!TIP]
> To download the zCLI directly, use the [latest release](https://github.com/zeropsio/zcli/releases/latest/) on GitHub.

## Quick Start

- Create a new personal access token from [settings/token-management](http://app.zerops.io/settings/token-management).

- Login to zCLI using the personal access token using the following command:

```Shell
zcli login <token>
```

- Push your project using the following command:

```Shell
zcli push
```


## Additional Documentation

For more information go through https://docs.zerops.io/references/cli.

## Want to Contribute?

Contributions to zCLI are welcome and highly appreciated. However, We would like you to go through [CONTRIBUTING.md](https://github.com/zeropsio/zcli/blob/main/CONTRIBUTING.md).

## Community

To chat with other community members, you can join the [Zerops Discord Server](https://discord.gg/xxzmJSDKPT).

