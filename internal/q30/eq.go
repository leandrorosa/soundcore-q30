package q30

import (
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
)

const (
	CustomEQ = 0xfefe
	minBand  = -12.0
	maxBand  = 12.0
)

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

// PresetID returns the protocol id of the named preset.
func PresetID(name string) (uint16, bool) {
	i := slices.IndexFunc(Presets, func(p Preset) bool { return p.Name == name })
	return uint16(i), i >= 0
}

// ParseBands parses 8 comma-separated gains in dB, such as "3,2,0,0,0,1,2,3".
func ParseBands(s string) ([8]float64, error) {
	var bands [8]float64
	parts := strings.Split(s, ",")
	if len(parts) != len(bands) {
		return bands, fmt.Errorf("want %d bands, got %d", len(bands), len(parts))
	}
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil || v < minBand || v > maxBand {
			return bands, fmt.Errorf("band %d: %q is not a number in %v..%v", i+1, p, minBand, maxBand)
		}
		bands[i] = v
	}
	return bands, nil
}
