# q30

Control Soundcore Life Q30 headphones from the terminal on macOS and Linux.

```
$ q30 status
battery   100%
mode      anc indoor
eq        signature
firmware  02.30
serial    3028ABCDEF012345
```

## Install

Download an archive from [releases](https://github.com/leandrorosa/soundcore-q30/releases)
(universal binary for macOS, amd64/arm64 for Linux) and put `q30` on your `PATH`.

Or build from source (Go 1.27+):

```sh
git clone https://github.com/leandrorosa/soundcore-q30
cd soundcore-q30
go build -o ~/bin/q30 .
```

**macOS:** the terminal app needs Bluetooth access (System Settings → Privacy & Security → Bluetooth).

**Linux:** needs BlueZ with `bluetoothctl`. Build with `CGO_ENABLED=0` to get a static binary.

### Shell completions

```sh
q30 completion bash > ~/.local/share/bash-completion/completions/q30
q30 completion zsh  > "${fpath[1]}/_q30"
q30 completion fish > ~/.config/fish/completions/q30.fish
```

The bash script needs the [bash-completion](https://github.com/scop/bash-completion) package.
Release archives also ship the scripts under `completions/`.

## Usage

Pair the headphones with your OS first, then:

```
q30 status                          battery, sound mode, eq, firmware
q30 anc [transport|indoor|outdoor]  noise cancelling, optionally picking the mode
q30 transparency
q30 normal
q30 toggle                          switch between anc and normal
q30 eq                              list eq presets
q30 eq rock                         apply a preset
q30 eq custom 3,2,0,0,0,1,2,3       custom eq, 8 bands in dB (-12..12)
q30 connect                         connect bluetooth
q30 off                             disconnect bluetooth
q30 completion bash|zsh|fish        print shell completion script
q30 --version
```

Run `q30 <command> --help` for details.

The device is found by name. To target a specific one, pass `-a/--addr` or set `Q30_ADDR`
(`00:11:22:33:44:55` and `00-11-22-33-44-55` both work).

## Limitations

- **`off` does not power the headphones off.** The Q30 has no remote power-off command, so
  `off` drops the Bluetooth connection.
- **No ANC over the AUX cable.** Soundcore removed it in a firmware update; it can't be
  enabled from the control protocol.
- Tested on macOS with firmware 02.30. The Linux backend builds but hasn't been tested
  against real hardware yet.

## How it works

The Q30 (model A3028) speaks a small binary protocol over Bluetooth RFCOMM (SPP).
Each packet carries a 2-byte command, a little-endian length and a one-byte checksum.

| Package             | What it does                                                         |
| ------------------- | -------------------------------------------------------------------- |
| `internal/cli`      | Commands, built with [Cobra](https://github.com/spf13/cobra).         |
| `internal/q30`      | Protocol client: packet encoding/decoding, state parsing, eq presets. Pure Go, transport-agnostic. |
| `internal/rfcomm`   | Bluetooth transport: IOBluetooth via cgo on macOS, BlueZ RFCOMM socket and `bluetoothctl` on Linux. |

The protocol was learned from [OpenSCQ30](https://github.com/Oppzippy/OpenSCQ30), which
supports many more Soundcore devices but has no macOS backend.

## Releasing

Push a `v*` tag; the GitHub Actions workflow runs GoReleaser on macOS (the macOS build needs cgo).

```sh
git tag v0.1.0 && git push origin v0.1.0
```
