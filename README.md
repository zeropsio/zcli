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

</div>

<br/>

<h3 align="end">
<a href="https://docs.zerops.io/" target="_blank">Read the docs →</a>
<br/>
</h3>

### Supported platforms

- Windows
- Linux
- MacOS (arm64, amd64)
- NixOS

### Optional requirements

- [Wireguard](https://www.wireguard.com/install/) - utilized by `zcli vpn` command.


<br/>


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



<br/>

<br/>


> [!TIP]
> To download the zCLI directly, locate the binary for your OS in the [latest release](https://github.com/zeropsio/zcli/releases/latest/) on GitHub.


<br/>


## Quick Start

- Create a new personal access token at [settings/token-management](http://app.zerops.io/settings/token-management) in Zerops GUI.

- Login to zCLI using the personal access token using the following command:

```Shell
zcli login <token>
```

- Run zcli to list commands and the current status

```Shell
zcli
```


<br/>



## Additional Documentation

For more information go through https://docs.zerops.io/references/cli.


<br/>


## Want to Contribute?

Contributions to zCLI are welcome and highly appreciated. However, We would like you to go through [CONTRIBUTING.md](https://github.com/zeropsio/zcli/blob/main/CONTRIBUTING.md).


<br/>

## Community

To chat with other community members, you can join the [Zerops Discord Server](https://discord.gg/xxzmJSDKPT).

