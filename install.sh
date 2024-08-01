#!/bin/sh
# Copyright (c) 2023-2024 Zerops s.r.o. All rights reserved. MIT license.

set -e

case $(uname -sm) in
"Darwin x86_64") target="darwin-amd64" ;;
"Darwin arm64") target="darwin-arm64" ;;
"Linux i386") target="linux-i386" ;;
*) target="linux-amd64" ;;
esac

if [ $# -eq 0 ]; then
  zcli_uri="https://github.com/zeropsio/zcli/releases/latest/download/zcli-${target}"
else
  zcli_uri="https://github.com/zeropsio/zcli/releases/download/${1}/zcli-${target}"
fi

bin_dir="$HOME/.local/bin"
bin_path="$bin_dir/zcli"
bin_dir_existed=1

if [ ! -d "$bin_dir" ]; then
  mkdir -p "$bin_dir"
  bin_dir_existed=0

  # By default `~/.local/bin` isn't included in PATH if it doesn't exist
  # First try `.bash_profile`. It doesn't exist by default, but if it does, `.profile` is ignored by bash
  if [ "$(uname -s)" = "Linux" ]; then
    if [ -f "$HOME/.bash_profile" ]; then
      . "$HOME/.bash_profile"
    elif [ -f "$HOME/.profile" ]; then
      . "$HOME/.profile"
    fi
  fi
fi

curl --fail --location --progress-bar --output "$bin_path" "$zcli_uri"
chmod +x "$bin_path"

echo
echo "zCLI was installed successfully to '$bin_path'"

if command -v zcli >/dev/null; then
  echo "Run 'zcli --help' to get started"
  if [ "$bin_dir_existed" = 0 ]; then
    echo "ℹ️ You may need to relaunch your shell."
  fi
else
  if [ "$(uname -s)" = "Darwin" ]; then
    echo 'Add following line to the `/etc/paths` file and relaunch your shell.';
    echo "  $HOME/.local/bin"
    echo
    echo 'You can do so by running:'
    echo "sudo sh -c 'echo \"$HOME/.local/bin\" >> /etc/paths'"
  else
    echo "Manually add the directory to your '$HOME/.profile' (or similar) and relaunch your shell."
    echo '  export PATH="$HOME/.local/bin:$PATH"'
  fi
  echo
  echo "Run '$bin_path --help' to get started"
fi

echo
echo "Stuck? Join our Discord https://discord.com/invite/WDvCZ54"
