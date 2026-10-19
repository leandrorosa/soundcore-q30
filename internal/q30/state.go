package q30

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type AmbientMode byte

const (
	ANC AmbientMode = iota
	Transparency
	Normal
)

func (m AmbientMode) String() string {
	return [...]string{"anc", "transparency", "normal"}[m%3]
}

type ANCMode byte

const (
	Transport ANCMode = iota
	Outdoor
	Indoor
)

var ANCModes = []string{"transport", "outdoor", "indoor"}

func (m ANCMode) String() string {
	if int(m) < len(ANCModes) {
		return ANCModes[m]
	}
	return fmt.Sprintf("unknown(%d)", byte(m))
}

func ParseANCMode(s string) (ANCMode, error) {
	for i, n := range ANCModes {
		if n == s {
			return ANCMode(i), nil
		}
	}
	return 0, fmt.Errorf("unknown anc mode %q, want one of: %s", s, strings.Join(ANCModes, ", "))
}

// SoundModes is the 4-byte block sent with CmdSoundModes; the last two bytes
// (transparency mode, custom anc level) are unused by the Q30 but must round-trip.
type SoundModes [4]byte

func (s SoundModes) Ambient() AmbientMode { return AmbientMode(s[0]) }
func (s SoundModes) ANC() ANCMode         { return ANCMode(s[1]) }

func (s SoundModes) With(a AmbientMode, n ANCMode) SoundModes {
	s[0], s[1] = byte(a), byte(n)
	return s
}

type State struct {
	Battery    int // 0..5
	Charging   bool
	EQPreset   uint16
	EQBands    [8]float64 // dB
	SoundModes SoundModes
	Firmware   string
	Serial     string
}

const stateLen = 60

func ParseState(b []byte) (State, error) {
	if len(b) < stateLen {
		return State{}, fmt.Errorf("state packet too short: %d bytes", len(b))
	}
	s := State{
		Battery:  int(b[0]),
		Charging: b[1] == 1,
		EQPreset: binary.LittleEndian.Uint16(b[2:4]),
		Firmware: string(b[39:44]),
		Serial:   string(b[44:60]),
	}
	for i, v := range b[4:12] {
		s.EQBands[i] = float64(int(v)-120) / 10
	}
	copy(s.SoundModes[:], b[35:39])
	return s, nil
}

func (s State) EQName() string {
	if s.EQPreset == CustomEQ {
		return "custom"
	}
	if int(s.EQPreset) < len(Presets) {
		return Presets[s.EQPreset].Name
	}
	return fmt.Sprintf("unknown(%#04x)", s.EQPreset)
}
