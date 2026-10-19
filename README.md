# q30

CLI to control Soundcore Life Q30 headphones from macOS over Bluetooth RFCOMM.
Protocol based on [OpenSCQ30](https://github.com/Oppzippy/OpenSCQ30), which has no macOS backend.

```sh
go build -o ~/bin/q30 .
```

```
q30 status                 # battery, mode, eq, firmware
q30 anc [transport|indoor|outdoor]
q30 transparency | normal
q30 toggle                 # anc <-> normal
q30 eq                     # list presets
q30 eq rock
q30 eq custom 3,2,0,0,0,1,2,3
q30 connect | off          # bluetooth connect / disconnect
```

The device is found by name; set `Q30_ADDR` or `-a` to pick a specific one.

The Q30 has no remote power-off command, so `off` disconnects Bluetooth and the
headphones shut down on their own after a while.

The terminal app needs Bluetooth permission (System Settings → Privacy & Security → Bluetooth).
