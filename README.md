# otter-release

This repository hosts the statically compiled Go installer for the otter
toolchain. The installer binary is published as a GitHub Release asset.

## Install

```bash
curl -fsSL -o otter-install https://github.com/rainoffallingstar/otter-release/releases/latest/download/otter-install-linux-amd64-static
chmod +x otter-install
./otter-install
```

## Source

The installer source lives in the `installer/` directory of the
[otter](https://github.com/rainoffallingstar/otter) repository.
