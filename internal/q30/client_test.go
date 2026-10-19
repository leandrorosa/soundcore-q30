package q30

import (
	"bytes"
	"testing"
	"time"
)

type fakeConn struct {
	written [][]byte
	replies [][]byte
}

func (f *fakeConn) Write(b []byte) error {
	f.written = append(f.written, b)
	return nil
}

func (f *fakeConn) Read(timeout time.Duration) ([]byte, error) {
	if len(f.replies) == 0 {
		time.Sleep(timeout)
		return nil, nil
	}
	r := f.replies[0]
	f.replies = f.replies[1:]
	return r, nil
}

func inbound(cmd Command, body []byte) []byte {
	p := Encode(cmd, body)
	copy(p, inboundPrefix)
	p[len(p)-1] = checksum(p[:len(p)-1])
	return p
}

func TestStateSkipsUnrelatedPackets(t *testing.T) {
	conn := &fakeConn{replies: [][]byte{
		inbound(Command{0x06, 0x01}, []byte{0, 1, 1, 0}),
		statePacket[:30],
		statePacket[30:],
	}}
	s, err := NewClient(conn).State()
	if err != nil {
		t.Fatal(err)
	}
	if s.Serial != "3029086EC82F12AC" {
		t.Errorf("serial = %q", s.Serial)
	}
	if want := Encode(CmdState, nil); !bytes.Equal(conn.written[0], want) {
		t.Errorf("wrote %x, want %x", conn.written[0], want)
	}
}

func TestSetEqualizer(t *testing.T) {
	conn := &fakeConn{replies: [][]byte{inbound(CmdEqualizer, nil)}}
	if err := NewClient(conn).SetEqualizer(CustomEQ, [8]float64{-6, 6, 2.3, 4, 2.2, 6, -0.4, 1.6}); err != nil {
		t.Fatal(err)
	}
	want := []byte{0x08, 0xee, 0x00, 0x00, 0x00, 0x02, 0x81, 0x14, 0x00, 0xfe, 0xfe, 0x3c, 0xb4, 0x8f, 0xa0, 0x8e, 0xb4, 0x74, 0x88, 0xe6}
	if !bytes.Equal(conn.written[0], want) {
		t.Errorf("wrote %x, want %x", conn.written[0], want)
	}
}

func TestCallTimesOut(t *testing.T) {
	c := NewClient(&fakeConn{})
	c.timeout = 10 * time.Millisecond
	if err := c.SetSoundModes(SoundModes{}); err == nil {
		t.Fatal("want timeout error")
	}
}
