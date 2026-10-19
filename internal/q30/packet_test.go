package q30

import (
	"bytes"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		cmd  Command
		body []byte
		want []byte
	}{
		{CmdState, nil, []byte{0x08, 0xee, 0x00, 0x00, 0x00, 0x01, 0x01, 0x0a, 0x00, 0x02}},
		{CmdSoundModes, []byte{0x00, 0x01, 0x01, 0x00}, []byte{0x08, 0xee, 0x00, 0x00, 0x00, 0x06, 0x81, 0x0e, 0x00, 0x00, 0x01, 0x01, 0x00, 0x8d}},
		{
			CmdEqualizer,
			append([]byte{0xfe, 0xfe}, EncodeBands([8]float64{-6, 6, 2.3, 4, 2.2, 6, -0.4, 1.6})...),
			[]byte{0x08, 0xee, 0x00, 0x00, 0x00, 0x02, 0x81, 0x14, 0x00, 0xfe, 0xfe, 0x3c, 0xb4, 0x8f, 0xa0, 0x8e, 0xb4, 0x74, 0x88, 0xe6},
		},
	}
	for _, tt := range tests {
		if got := Encode(tt.cmd, tt.body); !bytes.Equal(got, tt.want) {
			t.Errorf("Encode(%x) = %x, want %x", tt.cmd, got, tt.want)
		}
	}
}

var statePacket = []byte{
	0x09, 0xff, 0x00, 0x00, 0x01, 0x01, 0x01, 0x46, 0x00, 0x05, 0x00, 0x01, 0x00, 0x3c,
	0xb4, 0x8f, 0xa0, 0x8e, 0xb4, 0x74, 0x88, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x02, 0x00, 0x01, 0x00, 0x30, 0x32, 0x2e, 0x33, 0x30, 0x33, 0x30, 0x32,
	0x39, 0x30, 0x38, 0x36, 0x45, 0x43, 0x38, 0x32, 0x46, 0x31, 0x32, 0x41, 0x43, 0x35,
}

func TestDecodeStateAcrossChunks(t *testing.T) {
	var d Decoder
	noise := []byte{0x00, 0x09}
	if p := d.Feed(append(noise, statePacket[:20]...)); len(p) != 0 {
		t.Fatalf("got %d packets from partial input", len(p))
	}
	p := d.Feed(statePacket[20:])
	if len(p) != 1 || p[0].Command != CmdState {
		t.Fatalf("got %+v", p)
	}
	s, err := ParseState(p[0].Body)
	if err != nil {
		t.Fatal(err)
	}
	if s.Battery != 5 || s.Charging || s.EQName() != "acoustic" || s.Firmware != "02.30" ||
		s.Serial != "3029086EC82F12AC" || s.SoundModes.Ambient() != Normal || s.SoundModes.ANC() != Transport ||
		s.EQBands[0] != -6 {
		t.Errorf("unexpected state %+v", s)
	}
}

func TestDecodeSkipsBadChecksum(t *testing.T) {
	bad := append([]byte{}, statePacket...)
	bad[len(bad)-1]++
	var d Decoder
	if p := d.Feed(append(bad, statePacket...)); len(p) != 1 {
		t.Fatalf("got %d packets, want 1", len(p))
	}
}

func TestParseStateWithTrailingFields(t *testing.T) {
	body := []byte{
		5, 0, 14, 0, 120, 150, 150, 140, 160, 170, 150, 160, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0,
		48, 52, 46, 51, 51,
		51, 48, 50, 56, 54, 68, 67, 56, 57, 51, 52, 52, 52, 55, 57, 56,
		0, 0, 1, 1, 1, 1, 1,
	}
	s, err := ParseState(body)
	if err != nil {
		t.Fatal(err)
	}
	if s.EQName() != "piano" || s.Firmware != "04.33" {
		t.Errorf("unexpected state %+v", s)
	}
}
