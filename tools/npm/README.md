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
[![Discord](https://img.shields.io/discord/735781031147208777?labelColor=EDEFF3&color=8F9DA8)](https://discord.gg/xxzmJSDKPT)
  
</div>

<br/>

<h3 align="end">
<a href="https://docs.zerops.io/" target="_blank">Read the docs →</a>
<br/>
</h3>

## Install

```sh
npm i -g @zerops/zcli
```

Check out more installation ways at [zeropsio/zcli](https://github.com/zeropsio/zcli).

[!TIP]
> To download the zCLI directly, use the [latest release](https://github.com/zeropsio/zcli/releases/latest/) on GitHub.

## Requirements

- [Wireguard](https://www.wireguard.com/install/) - utilized by `zcli vpn` command.

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

## Support

Having trouble? Get help in the official [Zerops Discord Server](https://discord.gg/xxzmJSDKPT).


## Additional Documentation

For more information go through [zCLI Documentation](https://docs.zerops.io/references/cli).

## Want to Contribute?

Contributions to zCLI are welcome and highly appreciated. However, We would like you to go through [CONTRIBUTING.md](https://github.com/zeropsio/zcli/blob/main/CONTRIBUTING.md).
