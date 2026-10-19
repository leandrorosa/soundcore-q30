package q30

import (
	"encoding/binary"
	"fmt"
	"math"
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

var ancModes = []string{"transport", "outdoor", "indoor"}

func (m ANCMode) String() string {
	if int(m) < len(ancModes) {
		return ancModes[m]
	}
	return fmt.Sprintf("unknown(%d)", byte(m))
}

func ParseANCMode(s string) (ANCMode, error) {
	for i, n := range ancModes {
		if n == s {
			return ANCMode(i), nil
		}
	}
	return 0, fmt.Errorf("unknown anc mode %q, want one of: %s", s, strings.Join(ancModes, ", "))
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

const CustomEQ = 0xfefe

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

func EncodeBands(bands [8]float64) []byte {
	out := make([]byte, len(bands))
	for i, db := range bands {
		v := int(math.Round(db*10)) + 120
		out[i] = byte(min(max(v, 0), 240))
	}
	return out
}

type Preset struct {
	Name  string
	Bands [8]float64
}

// Presets are indexed by their protocol id.
var Presets = []Preset{
	{"signature", [8]float64{0, 0, 0, 0, 0, 0, 0, 0}},
	{"acoustic", [8]float64{4, 1, 2, 2, 4, 4, 4, 2}},
	{"bass-booster", [8]float64{4, 3, 1, 0, 0, 0, 0, 0}},
	{"bass-reducer", [8]float64{-4, -3, -1, 0, 0, 0, 0, 0}},
	{"classical", [8]float64{3, 3, -2, -2, 0, 2, 3, 4}},
	{"podcast", [8]float64{-3, 2, 4, 4, 3, 2, 0, -2}},
	{"dance", [8]float64{2, -3, -1, 1, 2, 2, 1, -3}},
	{"deep", [8]float64{2, 1, 3, 3, 2, -2, -4, -5}},
	{"electronic", [8]float64{3, 2, -2, 2, 1, 2, 3, 3}},
	{"flat", [8]float64{-2, -2, -1, 0, 0, 0, -2, -2}},
	{"hiphop", [8]float64{2, 3, -1, -1, 2, -1, 2, 3}},
	{"jazz", [8]float64{2, 2, -2, -2, 0, 2, 3, 4}},
	{"latin", [8]float64{0, 0, -2, -2, -2, 0, 3, 5}},
	{"lounge", [8]float64{-1, 2, 4, 3, 0, -2, 2, 1}},
	{"piano", [8]float64{0, 3, 3, 2, 4, 5, 3, 4}},
	{"pop", [8]float64{-1, 1, 3, 3, 1, -1, -2, -3}},
	{"rnb", [8]float64{6, 2, -2, -2, 2, 3, 3, 4}},
	{"rock", [8]float64{3, 2, -1, -1, 1, 3, 3, 3}},
	{"small-speakers", [8]float64{4, 3, 1, 0, -2, -3, -4, -4}},
	{"spoken-word", [8]float64{-3, -2, 1, 2, 2, 1, 0, -3}},
	{"treble-booster", [8]float64{-2, -2, -2, -1, 1, 2, 2, 4}},
	{"treble-reducer", [8]float64{0, 0, 0, -2, -3, -4, -4, -6}},
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
