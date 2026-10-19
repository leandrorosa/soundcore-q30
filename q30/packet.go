// Package q30 implements the Soundcore Life Q30 (A3028) RFCOMM protocol,
// as reverse engineered by OpenSCQ30.
package q30

import (
	"bytes"
	"encoding/binary"
)

type Command [2]byte

var (
	CmdState      = Command{0x01, 0x01}
	CmdEqualizer  = Command{0x02, 0x81}
	CmdSoundModes = Command{0x06, 0x81}
)

var (
	outboundPrefix = []byte{0x08, 0xee, 0x00, 0x00, 0x00}
	inboundPrefix  = []byte{0x09, 0xff, 0x00, 0x00, 0x01}
)

// header(5) + command(2) + length(2) + checksum(1)
const overhead = 10

type Packet struct {
	Command Command
	Body    []byte
}

func Encode(cmd Command, body []byte) []byte {
	n := overhead + len(body)
	p := append([]byte{}, outboundPrefix...)
	p = append(p, cmd[0], cmd[1])
	p = binary.LittleEndian.AppendUint16(p, uint16(n))
	p = append(p, body...)
	return append(p, checksum(p))
}

func checksum(b []byte) byte {
	var s byte
	for _, c := range b {
		s += c
	}
	return s
}

// Decoder reassembles inbound packets from an RFCOMM byte stream.
type Decoder struct{ buf []byte }

func (d *Decoder) Feed(b []byte) []Packet {
	d.buf = append(d.buf, b...)
	var out []Packet
	for {
		i := bytes.Index(d.buf, inboundPrefix)
		if i < 0 {
			if len(d.buf) > len(inboundPrefix) {
				d.buf = d.buf[len(d.buf)-len(inboundPrefix):]
			}
			return out
		}
		d.buf = d.buf[i:]
		if len(d.buf) < 9 {
			return out
		}
		n := int(binary.LittleEndian.Uint16(d.buf[7:9]))
		if n < overhead {
			d.buf = d.buf[1:]
			continue
		}
		if len(d.buf) < n {
			return out
		}
		raw := d.buf[:n]
		d.buf = d.buf[n:]
		if checksum(raw[:n-1]) != raw[n-1] {
			continue
		}
		out = append(out, Packet{
			Command: Command{raw[5], raw[6]},
			Body:    append([]byte{}, raw[9:n-1]...),
		})
	}
}
