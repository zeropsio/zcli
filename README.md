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

[![CI](https://img.shields.io/github/actions/workflow/status/zeropsio/zcli/main.yml?labelColor=EDEFF3&color=8F9DA8)](https://github.com/zeropsio/zcli/actions/workflows/ci.yml)
[![NPM Downloads](https://img.shields.io/npm/d18m/%40zerops%2Fzcli?labelColor=EDEFF3&color=8F9DA8)](https://www.npmjs.com/package/@zerops/zcli)
[![npm version](https://img.shields.io/badge/dynamic/json?color=8F9DA8&labelColor=EDEFF3&label=@zerops/zcli&query=version&url=https%3A%2F%2Fbadge.fury.io%2Fjs%2F@zerops%252Fzcli.json)](https://badge.fury.io/js/@zerops%2Fzcli)
[![CPU](https://img.shields.io/badge/CPU-x86%2C%20x64%2C%20ARM%2C%20ARM64-8F9DA8?labelColor=EDEFF3)](https://docs.abblix.com/docs/technical-requirements)
[![Discord](https://img.shields.io/discord/735781031147208777?labelColor=EDEFF3&color=8F9DA8)](https://discord.gg/xxzmJSDKPT)

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
- ping - utilized by `zcli vpn` command.


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

## Using Service Commands

Service commands accept the service ID (UUID) as a positional parameter:

```Shell
# First, list your services to get the service ID
zcli service list

# Start a service by ID (UUID)
zcli service start 12345678-1234-1234-1234-123456789012

# Push code to a service
zcli push 12345678-1234-1234-1234-123456789012

# Get logs from a service
zcli service log 12345678-1234-1234-1234-123456789012

# Stop a service
zcli service stop 12345678-1234-1234-1234-123456789012
```

If you don't provide a service ID, zCLI will show an interactive selector to choose from available services.


<br/>



## Additional Documentation

For more information go through https://docs.zerops.io/references/cli.


<br/>


## Want to Contribute?

Contributions to zCLI are welcome and highly appreciated. However, We would like you to go through [CONTRIBUTING.md](https://github.com/zeropsio/zcli/blob/main/CONTRIBUTING.md).


<br/>

## Community

To chat with other community members, you can join the [Zerops Discord Server](https://discord.gg/xxzmJSDKPT).
