package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/leandrorosa/soundcore-q30/q30"
	"github.com/leandrorosa/soundcore-q30/rfcomm"
)

func init() { runtime.LockOSThread() }

const usage = `Usage: q30 [-a ADDR] <command> [args]

Commands:
  status                       battery, sound mode, eq, firmware
  anc [transport|indoor|outdoor]
  transparency
  normal
  toggle                       switch between anc and normal
  eq [PRESET]                  list presets, or apply one
  eq custom B1,...,B8          custom eq in dB (-12..12)
  connect                      connect bluetooth
  off                          disconnect bluetooth (the Q30 has no remote power off;
                               it shuts down by itself after a while disconnected)

The device is found by name ("Q30") unless -a or $Q30_ADDR is set.
`

func main() {
	addr := flag.String("a", os.Getenv("Q30_ADDR"), "bluetooth address")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usage) }
	flag.Parse()
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*addr, flag.Arg(0), flag.Args()[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "q30:", err)
		os.Exit(1)
	}
}

func run(addr, cmd string, args []string) error {
	if cmd == "eq" && len(args) == 0 {
		for _, p := range q30.Presets {
			fmt.Println(p.Name)
		}
		return nil
	}
	if addr == "" {
		var err error
		if addr, err = rfcomm.Find("Q30"); err != nil {
			return err
		}
	}
	switch cmd {
	case "connect":
		return rfcomm.Connect(addr)
	case "off", "disconnect":
		return rfcomm.Disconnect(addr)
	}
	if !rfcomm.Connected(addr) {
		return errors.New("headphones not connected, run: q30 connect")
	}

	d, err := open(addr)
	if err != nil {
		return err
	}
	defer d.Close()
	s, err := d.state()
	if err != nil {
		return err
	}

	switch cmd {
	case "status":
		printStatus(s)
		return nil
	case "anc":
		mode := s.SoundModes.ANC()
		if len(args) > 0 {
			if mode, err = q30.ParseANCMode(args[0]); err != nil {
				return err
			}
		}
		return d.setMode(s, q30.ANC, mode)
	case "transparency":
		return d.setMode(s, q30.Transparency, s.SoundModes.ANC())
	case "normal":
		return d.setMode(s, q30.Normal, s.SoundModes.ANC())
	case "toggle":
		next := q30.ANC
		if s.SoundModes.Ambient() == q30.ANC {
			next = q30.Normal
		}
		return d.setMode(s, next, s.SoundModes.ANC())
	case "eq":
		return d.setEQ(args)
	}
	return fmt.Errorf("unknown command %q", cmd)
}

func (d *device) setMode(s q30.State, a q30.AmbientMode, n q30.ANCMode) error {
	m := s.SoundModes.With(a, n)
	if _, err := d.call(q30.CmdSoundModes, m[:]); err != nil {
		return err
	}
	if a == q30.ANC {
		fmt.Println(a, n)
	} else {
		fmt.Println(a)
	}
	return nil
}

func (d *device) setEQ(args []string) error {
	var id uint16
	var bands [8]float64
	if args[0] == "custom" {
		if len(args) != 2 {
			return errors.New("usage: q30 eq custom B1,...,B8")
		}
		parts := strings.Split(args[1], ",")
		if len(parts) != len(bands) {
			return fmt.Errorf("want %d bands, got %d", len(bands), len(parts))
		}
		for i, p := range parts {
			v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil || v < -12 || v > 12 {
				return fmt.Errorf("band %d: %q is not a number in -12..12", i+1, p)
			}
			bands[i] = v
		}
		id = q30.CustomEQ
	} else {
		i := slices.IndexFunc(q30.Presets, func(p q30.Preset) bool { return p.Name == args[0] })
		if i < 0 {
			return fmt.Errorf("unknown preset %q, run `q30 eq` to list them", args[0])
		}
		id, bands = uint16(i), q30.Presets[i].Bands
	}
	body := binary.LittleEndian.AppendUint16(nil, id)
	if _, err := d.call(q30.CmdEqualizer, append(body, q30.EncodeBands(bands)...)); err != nil {
		return err
	}
	fmt.Println("eq", args[0])
	return nil
}

func printStatus(s q30.State) {
	battery := fmt.Sprintf("%d%%", s.Battery*20)
	if s.Charging {
		battery += " (charging)"
	}
	mode := s.SoundModes.Ambient().String()
	if s.SoundModes.Ambient() == q30.ANC {
		mode += " " + s.SoundModes.ANC().String()
	}
	eq := s.EQName()
	if s.EQPreset == q30.CustomEQ {
		eq += fmt.Sprintf(" %v", s.EQBands)
	}
	fmt.Printf("battery   %s\nmode      %s\neq        %s\n", battery, mode, eq)
	fmt.Printf("firmware  %s\nserial    %s\n", s.Firmware, s.Serial)
}
